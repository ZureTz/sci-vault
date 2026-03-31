package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"gateway/internal/dto"
	"gateway/internal/model"
	"gateway/internal/repo"
	"gateway/pkg/app_error"
	"gateway/pkg/cache"
	"gateway/pkg/grpc_client"
	"gateway/pkg/storage"
)

const (
	maxDocumentSize   = 100 << 20 // 100 MB
	downloadURLExpiry = 15 * time.Minute
	enrichStatusTTL   = 24 * time.Hour
)

// DB enrich_status values (source of truth for persistent state).
// Fine-grained transient states (pending, processing, failed) live in Redis only,
// managed entirely by the Python microservice.
const (
	EnrichStatusNotStarted = "not_started"
	EnrichStatusDone       = "done"
)

func enrichStatusKey(docID uint) string {
	return fmt.Sprintf("doc:enrich:%d", docID)
}

type DocumentService struct {
	repo              repo.DocumentRepository
	storageClient     *storage.Client
	recommenderClient *grpc_client.RecommenderClient
	cacheConn         *cache.CacheConnector
}

func NewDocumentService(
	repo repo.DocumentRepository,
	storageClient *storage.Client,
	recommenderClient *grpc_client.RecommenderClient,
	cacheConn *cache.CacheConnector,
) *DocumentService {
	return &DocumentService{
		repo:              repo,
		storageClient:     storageClient,
		recommenderClient: recommenderClient,
		cacheConn:         cacheConn,
	}
}

func (s *DocumentService) UploadDocument(ctx context.Context, userID uint, file io.Reader, form dto.UploadDocumentForm) (*dto.DocumentResponse, error) {
	if form.File.Size > maxDocumentSize {
		return nil, app_error.ErrDocumentTooLarge
	}

	contentType := strings.ToLower(form.File.Header.Get("Content-Type"))
	if contentType != "application/pdf" {
		return nil, app_error.ErrDocumentInvalidType
	}

	ts := time.Now().UTC()
	hash := sha256.Sum256([]byte(form.File.Filename + ts.Format(time.RFC3339Nano)))
	key := fmt.Sprintf("documents/%s/%x", ts.Format("20060102"), hash)

	if err := s.storageClient.PutObject(ctx, key, file, contentType, true); err != nil {
		return nil, fmt.Errorf("failed to upload document: %w", err)
	}

	title := form.Title
	if title == "" {
		title = form.File.Filename
	}

	doc := &model.Document{
		Title:            title,
		OriginalFileName: form.File.Filename,
		FileKey:          key,
		FileSize:         form.File.Size,
		ContentType:      contentType,
		Year:             form.Year,
		DOI:              form.DOI,
		UploadedByUserID: userID,
	}
	if err := s.repo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to save document: %w", err)
	}

	// Trigger async enrichment on the Python microservice.
	// The call returns an immediate ACK; Python owns all status updates from this point.
	// A failure here is non-fatal — the document is already saved.
	if _, err := s.recommenderClient.EnrichDocument(ctx, uint64(doc.ID), key); err != nil {
		slog.Warn("EnrichDocument gRPC call failed", "docID", doc.ID, "err", err)
	}

	downloadURL, err := s.storageClient.PrivateObjectURL(ctx, key, downloadURLExpiry, downloadFilename(doc.OriginalFileName))
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return toDocumentResponse(doc, downloadURL), nil
}

func (s *DocumentService) GetDocument(ctx context.Context, docID uint) (*dto.DocumentResponse, error) {
	doc, err := s.repo.FindByID(ctx, docID)
	if err != nil {
		return nil, app_error.ErrDocumentNotFound
	}

	if err := s.repo.IncrementViewCount(ctx, docID); err != nil {
		return nil, fmt.Errorf("failed to increment view count: %w", err)
	}
	doc.ViewCount++

	downloadURL, err := s.storageClient.PrivateObjectURL(ctx, doc.FileKey, downloadURLExpiry, downloadFilename(doc.OriginalFileName))
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return toDocumentResponse(&doc, downloadURL), nil
}

func (s *DocumentService) GetEnrichStatus(ctx context.Context, docID uint) (string, error) {
	// Fast path: Redis
	if status, err := s.cacheConn.Get(ctx, enrichStatusKey(docID)); err == nil {
		return status, nil
	}
	// Fallback: DB (handles Redis TTL expiry and cache misses)
	doc, err := s.repo.FindByID(ctx, docID)
	if err != nil {
		return "", app_error.ErrDocumentNotFound
	}
	return doc.EnrichStatus, nil
}

// downloadFilename ensures the filename ends with ".pdf".
func downloadFilename(original string) string {
	if strings.HasSuffix(strings.ToLower(original), ".pdf") {
		return original
	}
	return original + ".pdf"
}

func toDocumentResponse(doc *model.Document, downloadURL string) *dto.DocumentResponse {
	authors := []string(doc.Authors)
	if authors == nil {
		authors = []string{}
	}
	tags := []string(doc.Tags)
	if tags == nil {
		tags = []string{}
	}
	return &dto.DocumentResponse{
		ID:               doc.ID,
		Title:            doc.Title,
		OriginalFileName: doc.OriginalFileName,
		FileSize:         doc.FileSize,
		ContentType:      doc.ContentType,
		Year:             doc.Year,
		DOI:              doc.DOI,
		EnrichStatus:     doc.EnrichStatus,
		Authors:          authors,
		Summary:          doc.Summary,
		Tags:             tags,
		ViewCount:        doc.ViewCount,
		LikeCount:        doc.LikeCount,
		UploadedByUserID: doc.UploadedByUserID,
		DownloadURL:      downloadURL,
		CreatedAt:        doc.CreatedAt,
	}
}
