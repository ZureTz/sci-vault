package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"gateway/internal/dto"
	"gateway/internal/model"
	"gateway/internal/repo"
	"gateway/pkg/app_error"
	"gateway/pkg/cache"
	"gateway/pkg/grpc_client"
	"gateway/pkg/storage"
)

// allowedUploadTypes is the set of MIME types accepted by the upload endpoints.
// PDFs and plaintext flow straight through; the three OOXML office formats are
// converted to PDF inside svc-recommender before enrichment. Keep this in sync
// with the conversion module in svc-recommender (src/conversion/converter.py).
var allowedUploadTypes = map[string]struct{}{
	"application/pdf": {},
	"text/plain":      {},
	"text/markdown":   {},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {},
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": {},
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         {},
}

// extensionContentTypes is the authoritative ext→MIME map for our supported
// upload formats. It's checked before Go's `mime.TypeByExtension`, which on
// many systems doesn't know about `.md` (and varies by OS for the OOXML
// formats too) — making the allowlist effectively dependent on the host's
// /etc/mime.types is fragile.
var extensionContentTypes = map[string]string{
	".pdf":  "application/pdf",
	".txt":  "text/plain",
	".md":   "text/markdown",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
}

// resolveUploadContentType normalises the multipart Content-Type header.
// Browsers often send empty or generic `application/octet-stream` for `.md`
// and other less-common types, so fall back to the filename extension. We
// also re-resolve from extension when the header disagrees with our known
// extension mapping — some browsers report `.md` as `text/x-markdown` or
// the OOXML formats as `application/zip`.
func resolveUploadContentType(header, filename string) string {
	ct := strings.ToLower(strings.TrimSpace(header))
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if known, ok := extensionContentTypes[ext]; ok {
		// Trust the extension over an octet-stream / empty / mismatched header.
		if ct == "" || ct == "application/octet-stream" || ct != known {
			return known
		}
		return ct
	}

	if ct == "" || ct == "application/octet-stream" {
		if sniffed := mime.TypeByExtension(ext); sniffed != "" {
			ct = strings.ToLower(strings.TrimSpace(strings.SplitN(sniffed, ";", 2)[0]))
		}
	}
	return ct
}

func isAllowedUploadType(contentType string) bool {
	_, ok := allowedUploadTypes[contentType]
	return ok
}

// storageContentType is the value sent to S3 as the object's Content-Type
// header. Text-y formats get an explicit `charset=utf-8` so browsers render
// non-ASCII content correctly when hitting the presigned download URL —
// without it they fall back to a system default and mangle UTF-8. The DB
// `Document.ContentType` field stays the bare MIME so the recommender's
// dispatcher (which matches exact strings) keeps working.
func storageContentType(ct string) string {
	switch ct {
	case "text/plain", "text/markdown":
		return ct + "; charset=utf-8"
	}
	return ct
}

const (
	maxDocumentSize   = 100 << 20 // 100 MB
	downloadURLExpiry = 15 * time.Minute
	enrichStatusTTL   = 24 * time.Hour
)

func enrichStatusKey(docID uint) string {
	return fmt.Sprintf("doc:enrich:%d", docID)
}

type DocumentService struct {
	repo              repo.DocumentRepository
	interactionRepo   repo.DocumentInteractionRepository
	labRepo           repo.LabRepository
	storageClient     *storage.Client
	recommenderClient *grpc_client.RecommenderClient
	cacheConn         *cache.CacheConnector
}

func NewDocumentService(
	repo repo.DocumentRepository,
	interactionRepo repo.DocumentInteractionRepository,
	labRepo repo.LabRepository,
	storageClient *storage.Client,
	recommenderClient *grpc_client.RecommenderClient,
	cacheConn *cache.CacheConnector,
) *DocumentService {
	return &DocumentService{
		repo:              repo,
		interactionRepo:   interactionRepo,
		labRepo:           labRepo,
		storageClient:     storageClient,
		recommenderClient: recommenderClient,
		cacheConn:         cacheConn,
	}
}

// dedupViolation classifies a DB error as a unique-violation raised by one of
// the dedup partial indexes (private-per-user or lab-wide). Returns the matching
// sentinel (ErrDocumentDuplicate or ErrDocumentDuplicateInLab) or nil when err
// is not a dedup-index violation. Used to detect a race where a concurrent
// upload persisted the same content_sha256 first.
func dedupViolation(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "idx_documents_private_user_sha"):
		return app_error.ErrDocumentDuplicate
	case strings.Contains(msg, "idx_documents_lab_sha"):
		return app_error.ErrDocumentDuplicateInLab
	}
	return nil
}

// duplicateErrorFor returns the appropriate duplicate sentinel for a given
// destination visibility.
func duplicateErrorFor(visibility string) error {
	if visibility == model.DocVisibilityLab {
		return app_error.ErrDocumentDuplicateInLab
	}
	return app_error.ErrDocumentDuplicate
}

// canAccessDocument reports whether userID is allowed to read doc.
// Access is granted when the user is the uploader, or when the document
// is lab-visible and the user is a member of that lab.
func (s *DocumentService) canAccessDocument(ctx context.Context, userID uint, doc *model.Document) (bool, error) {
	if doc.UploadedByUserID == userID {
		return true, nil
	}
	if doc.Visibility == model.DocVisibilityLab && doc.LabID != nil {
		if _, err := s.labRepo.FindMember(ctx, *doc.LabID, userID); err == nil {
			return true, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return false, err
		}
	}
	return false, nil
}

// canManageDocument reports whether userID is allowed to mutate doc
// (edit metadata, restart enrichment, delete). Two paths grant authority:
// being the uploader, or being the OWNER of the lab the doc is shared into.
// Sits alongside canAccessDocument — same shape, but for write authority.
func (s *DocumentService) canManageDocument(ctx context.Context, userID uint, doc *model.Document) (bool, error) {
	if doc.UploadedByUserID == userID {
		return true, nil
	}
	if doc.Visibility == model.DocVisibilityLab && doc.LabID != nil {
		member, err := s.labRepo.FindMember(ctx, *doc.LabID, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, nil
			}
			return false, err
		}
		if member.Role == model.LabRoleOwner {
			return true, nil
		}
	}
	return false, nil
}

// resolveVisibility validates a visibility/labID pair and returns the canonical values to persist.
// If visibility is "lab", labID must be provided AND the user must be a member of that lab.
// If visibility is "private", labID is forced to nil.
func (s *DocumentService) resolveVisibility(ctx context.Context, userID uint, visibility string, labID *uint) (string, *uint, error) {
	switch visibility {
	case model.DocVisibilityPrivate:
		return model.DocVisibilityPrivate, nil, nil
	case model.DocVisibilityLab:
		if labID == nil || *labID == 0 {
			return "", nil, app_error.ErrLabRequiredForLabVis
		}
		if _, err := s.labRepo.FindMember(ctx, *labID, userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", nil, app_error.ErrNotMember
			}
			return "", nil, err
		}
		return model.DocVisibilityLab, labID, nil
	default:
		return "", nil, app_error.ErrInvalidVisibility
	}
}

func (s *DocumentService) UploadDocument(ctx context.Context, userID uint, file io.Reader, form dto.UploadDocumentForm) (*dto.DocumentResponse, error) {
	if form.File.Size > maxDocumentSize {
		return nil, app_error.ErrDocumentTooLarge
	}

	contentType := resolveUploadContentType(form.File.Header.Get("Content-Type"), form.File.Filename)
	if !isAllowedUploadType(contentType) {
		return nil, app_error.ErrDocumentInvalidType
	}

	// Resolve visibility (default: private)
	visibility := model.DocVisibilityPrivate
	if form.Visibility != nil && *form.Visibility != "" {
		visibility = *form.Visibility
	}
	resolvedVis, resolvedLabID, err := s.resolveVisibility(ctx, userID, visibility, form.LabID)
	if err != nil {
		return nil, err
	}

	// Buffer the whole upload so we can hash it before committing to S3.
	// Size was validated above (<= 100 MB), so holding it in memory is acceptable.
	buf, err := io.ReadAll(io.LimitReader(file, maxDocumentSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read document body: %w", err)
	}
	if int64(len(buf)) > maxDocumentSize {
		return nil, app_error.ErrDocumentTooLarge
	}
	contentHash := sha256.Sum256(buf)
	contentSHA := hex.EncodeToString(contentHash[:])

	// Duplicate guard: scoped to the destination context.
	// - private → the same user's private library
	// - lab → the selected lab (regardless of uploader)
	if _, err := s.repo.FindExistingByHash(ctx, resolvedVis, userID, resolvedLabID, contentSHA); err == nil {
		return nil, duplicateErrorFor(resolvedVis)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check for duplicate: %w", err)
	}

	ts := time.Now().UTC()
	keyHash := sha256.Sum256([]byte(form.File.Filename + ts.Format(time.RFC3339Nano)))
	key := fmt.Sprintf("documents/%s/%x", ts.Format("20060102"), keyHash)

	if err := s.storageClient.PutObject(ctx, key, bytes.NewReader(buf), storageContentType(contentType), true); err != nil {
		return nil, fmt.Errorf("failed to upload document: %w", err)
	}

	doc := &model.Document{
		Title:            form.Title,
		OriginalFileName: form.File.Filename,
		FileKey:          key,
		FileSize:         form.File.Size,
		ContentType:      contentType,
		ContentSHA256:    contentSHA,
		Year:             form.Year,
		DOI:              form.DOI,
		UploadedByUserID: userID,
		Visibility:       resolvedVis,
		LabID:            resolvedLabID,
	}
	if err := s.repo.Create(ctx, doc); err != nil {
		// Race with a partial unique index — another concurrent upload won.
		if dupErr := dedupViolation(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("failed to save document: %w", err)
	}

	// Set the enrichment status in Redis with a TTL to not_started
	if err := s.cacheConn.Set(ctx, enrichStatusKey(doc.ID), model.EnrichStatusNotStarted, enrichStatusTTL); err != nil {
		slog.Warn("Failed to set enrich status in cache", "docID", doc.ID, "err", err)
	}

	// Invalidate the user's dashboard stats cache so they see the fresh data
	if _, err := s.cacheConn.Del(ctx, dashboardStatsKey(userID)); err != nil {
		slog.Warn("Failed to invalidate dashboard stats cache", "userID", userID, "err", err)
	}

	// Trigger async enrichment on the Python microservice.
	// The call returns an immediate ACK; Python owns all status updates from this point.
	// A failure here is non-fatal — the document is already saved.
	if _, err := s.recommenderClient.EnrichDocument(ctx, uint64(doc.ID), key, contentType); err != nil {
		slog.Warn("EnrichDocument gRPC call failed", "docID", doc.ID, "err", err)
	}

	downloadURL, err := s.storageClient.PrivateObjectURL(ctx, key, downloadURLExpiry, downloadFilename(doc.OriginalFileName))
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return toDocumentResponse(doc, downloadURL, false), nil
}

// BatchUploadDocuments uploads multiple files sharing one metadata envelope
// (visibility + optional lab_id). DB work is batched — one query for duplicate
// detection across all files, one INSERT for all surviving records — instead
// of N round trips. Per-file failures are reported via BatchUploadItemResult.Error
// as the raw sentinel error message; the handler translates to i18n codes.
func (s *DocumentService) BatchUploadDocuments(ctx context.Context, userID uint, form dto.BatchUploadDocumentForm) (*dto.BatchUploadDocumentResponse, error) {
	// Whole-batch visibility resolution: invalid visibility or non-member lab
	// fails the entire request rather than repeating the same error N times.
	visibility := model.DocVisibilityPrivate
	if form.Visibility != nil && *form.Visibility != "" {
		visibility = *form.Visibility
	}
	resolvedVis, resolvedLabID, err := s.resolveVisibility(ctx, userID, visibility, form.LabID)
	if err != nil {
		return nil, err
	}

	// Phase 1 — buffer + hash + per-file pre-flight validation.
	type pending struct {
		fh          *multipart.FileHeader
		buf         []byte
		contentType string
		hash        string
		err         string // non-empty when this file is already doomed (size/type/read)
	}
	pendingDocs := make([]pending, len(form.Files))
	hashesToCheck := make([]string, 0, len(form.Files))
	for i, fh := range form.Files {
		p := pending{fh: fh}
		if fh.Size > maxDocumentSize {
			p.err = app_error.ErrDocumentTooLarge.Error()
			pendingDocs[i] = p
			continue
		}
		ct := resolveUploadContentType(fh.Header.Get("Content-Type"), fh.Filename)
		if !isAllowedUploadType(ct) {
			p.err = app_error.ErrDocumentInvalidType.Error()
			pendingDocs[i] = p
			continue
		}
		p.contentType = ct

		file, openErr := fh.Open()
		if openErr != nil {
			p.err = openErr.Error()
			pendingDocs[i] = p
			continue
		}
		buf, readErr := io.ReadAll(io.LimitReader(file, maxDocumentSize+1))
		file.Close()
		if readErr != nil {
			p.err = readErr.Error()
			pendingDocs[i] = p
			continue
		}
		if int64(len(buf)) > maxDocumentSize {
			p.err = app_error.ErrDocumentTooLarge.Error()
			pendingDocs[i] = p
			continue
		}
		sum := sha256.Sum256(buf)
		p.buf = buf
		p.hash = hex.EncodeToString(sum[:])
		pendingDocs[i] = p
		hashesToCheck = append(hashesToCheck, p.hash)
	}

	// Phase 2 — single-query dedup check scoped to the destination (user's private
	// library or the target lab).
	existingHashes := map[string]bool{}
	if len(hashesToCheck) > 0 {
		found, err := s.repo.FindExistingHashesInSet(ctx, resolvedVis, userID, resolvedLabID, hashesToCheck)
		if err != nil {
			return nil, fmt.Errorf("failed to batch check duplicates: %w", err)
		}
		for _, h := range found {
			existingHashes[h] = true
		}
	}

	// Phase 3 — S3 put for each survivor, collect docs to insert.
	// Intra-batch dedup: if the same file appears twice in a private batch,
	// only the first is persisted; subsequent ones are marked duplicate.
	results := make([]dto.BatchUploadItemResult, len(form.Files))
	toInsert := make([]*model.Document, 0, len(form.Files))
	insertIdx := make([]int, 0, len(form.Files)) // result index per inserted doc
	seenInBatch := map[string]bool{}
	ts := time.Now().UTC()
	dayPrefix := ts.Format("20060102")
	tsNano := ts.Format(time.RFC3339Nano)
	for i := range pendingDocs {
		p := pendingDocs[i]
		results[i].Filename = p.fh.Filename
		if p.err != "" {
			results[i].Error = p.err
			continue
		}
		if existingHashes[p.hash] || seenInBatch[p.hash] {
			results[i].Error = duplicateErrorFor(resolvedVis).Error()
			continue
		}
		seenInBatch[p.hash] = true

		keyHash := sha256.Sum256([]byte(p.fh.Filename + tsNano + fmt.Sprint(i)))
		key := fmt.Sprintf("documents/%s/%x", dayPrefix, keyHash)
		if err := s.storageClient.PutObject(ctx, key, bytes.NewReader(p.buf), storageContentType(p.contentType), true); err != nil {
			results[i].Error = err.Error()
			continue
		}
		toInsert = append(toInsert, &model.Document{
			OriginalFileName: p.fh.Filename,
			FileKey:          key,
			FileSize:         p.fh.Size,
			ContentType:      p.contentType,
			ContentSHA256:    p.hash,
			UploadedByUserID: userID,
			Visibility:       resolvedVis,
			LabID:            resolvedLabID,
		})
		insertIdx = append(insertIdx, i)
	}

	// Phase 4 — single batch INSERT, then post-insert side effects per doc.
	if len(toInsert) > 0 {
		if err := s.repo.CreateBatch(ctx, toInsert); err != nil {
			// Unique-index race: another concurrent request persisted the same hash.
			// Fall back to per-row inserts so one colliding file doesn't fail the batch.
			if dedupViolation(err) != nil {
				for n, doc := range toInsert {
					if createErr := s.repo.Create(ctx, doc); createErr != nil {
						if dupErr := dedupViolation(createErr); dupErr != nil {
							results[insertIdx[n]].Error = dupErr.Error()
							continue
						}
						results[insertIdx[n]].Error = createErr.Error()
						continue
					}
					id := doc.ID
					results[insertIdx[n]].DocID = &id
				}
			} else {
				return nil, fmt.Errorf("failed to batch insert documents: %w", err)
			}
		} else {
			for n, doc := range toInsert {
				id := doc.ID
				results[insertIdx[n]].DocID = &id
			}
		}

		for _, doc := range toInsert {
			if doc.ID == 0 {
				continue
			}
			if err := s.cacheConn.Set(ctx, enrichStatusKey(doc.ID), model.EnrichStatusNotStarted, enrichStatusTTL); err != nil {
				slog.Warn("Failed to set enrich status in cache", "docID", doc.ID, "err", err)
			}
			if _, err := s.recommenderClient.EnrichDocument(ctx, uint64(doc.ID), doc.FileKey, doc.ContentType); err != nil {
				slog.Warn("EnrichDocument gRPC call failed", "docID", doc.ID, "err", err)
			}
		}
		if _, err := s.cacheConn.Del(ctx, dashboardStatsKey(userID)); err != nil {
			slog.Warn("Failed to invalidate dashboard stats cache", "userID", userID, "err", err)
		}
	}

	resp := &dto.BatchUploadDocumentResponse{Results: results}
	for _, r := range results {
		if r.Error != "" {
			resp.Failed++
		} else {
			resp.Succeeded++
		}
	}
	return resp, nil
}

func (s *DocumentService) GetDocument(ctx context.Context, userID, docID uint) (*dto.DocumentResponse, error) {
	doc, err := s.repo.FindByID(ctx, docID)
	if err != nil {
		return nil, app_error.ErrDocumentNotFound
	}

	ok, err := s.canAccessDocument(ctx, userID, &doc)
	if err != nil {
		return nil, fmt.Errorf("failed to check document access: %w", err)
	}
	if !ok {
		return nil, app_error.ErrDocumentNotFound
	}

	if inserted, err := s.interactionRepo.RecordView(ctx, userID, docID, model.ViewThrottleWindow); err != nil {
		return nil, fmt.Errorf("failed to record view: %w", err)
	} else if inserted {
		doc.ViewCount++
	}

	likedByMe, err := s.interactionRepo.IsLikedBy(ctx, userID, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to load like state: %w", err)
	}

	downloadURL, err := s.storageClient.PrivateObjectURL(ctx, doc.FileKey, downloadURLExpiry, downloadFilename(doc.OriginalFileName))
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return toDocumentResponse(&doc, downloadURL, likedByMe), nil
}

func (s *DocumentService) GetEnrichStatus(ctx context.Context, userID, docID uint) (string, error) {
	// Authorize against DB — we need the row to check ownership/lab visibility anyway.
	doc, err := s.repo.FindByID(ctx, docID)
	if err != nil {
		return "", app_error.ErrDocumentNotFound
	}
	ok, err := s.canAccessDocument(ctx, userID, &doc)
	if err != nil {
		return "", fmt.Errorf("failed to check document access: %w", err)
	}
	if !ok {
		return "", app_error.ErrDocumentNotFound
	}

	// Fast path: Redis
	if status, err := s.cacheConn.Get(ctx, enrichStatusKey(docID)); err == nil {
		return status, nil
	}
	return doc.EnrichStatus, nil
}

func (s *DocumentService) RestartEnrichment(ctx context.Context, userID, docID uint) error {
	doc, err := s.repo.FindByID(ctx, docID)
	if err != nil {
		return app_error.ErrDocumentNotFound
	}
	if ok, err := s.canManageDocument(ctx, userID, &doc); err != nil {
		return fmt.Errorf("failed to check document manage permission: %w", err)
	} else if !ok {
		return app_error.ErrNotDocumentOwner
	}

	// Set the enrichment status in Redis to not_started
	if err := s.cacheConn.Set(ctx, enrichStatusKey(doc.ID), model.EnrichStatusNotStarted, enrichStatusTTL); err != nil {
		slog.Warn("Failed to set enrich status in cache", "docID", doc.ID, "err", err)
	}

	// Invalidate the dashboard stats cache so status breakdown counts are fresh
	if _, err := s.cacheConn.Del(ctx, dashboardStatsKey(doc.UploadedByUserID)); err != nil {
		slog.Warn("Failed to invalidate dashboard stats cache", "userID", doc.UploadedByUserID, "err", err)
	}

	// Trigger async enrichment on the Python microservice.
	if _, err := s.recommenderClient.EnrichDocument(ctx, uint64(doc.ID), doc.FileKey, doc.ContentType); err != nil {
		slog.Warn("EnrichDocument gRPC call failed", "docID", doc.ID, "err", err)
		return fmt.Errorf("failed to restart enrichment: %w", err)
	}

	return nil
}

func (s *DocumentService) ListMyDocuments(ctx context.Context, userID uint, query dto.ListMyDocumentsQuery) (*dto.ListMyDocumentsResponse, error) {
	page := query.Page
	if page == 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize == 0 {
		pageSize = 20
	}
	filter := repo.ListMyDocumentsFilter{
		Search:     strings.TrimSpace(query.Search),
		Status:     query.Status,
		Visibility: query.Visibility,
		LabID:      query.LabID,
		SortBy:     query.SortBy,
		SortOrder:  query.SortOrder,
		Offset:     (page - 1) * pageSize,
		Limit:      pageSize,
	}
	docs, total, err := s.repo.FindByUserID(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	items := make([]dto.DocumentListItem, 0, len(docs))
	for i := range docs {
		items = append(items, toDocumentListItem(&docs[i]))
	}

	return &dto.ListMyDocumentsResponse{
		Documents: items,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// ListLabDocuments lists every lab-visible document in a lab, regardless of
// uploader. Owner-only: members see ErrNotOwner, non-members see ErrNotMember.
func (s *DocumentService) ListLabDocuments(ctx context.Context, requesterID, labID uint, query dto.ListLabDocumentsQuery) (*dto.ListMyDocumentsResponse, error) {
	member, err := s.labRepo.FindMember(ctx, labID, requesterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, app_error.ErrNotMember
		}
		return nil, fmt.Errorf("failed to load lab membership: %w", err)
	}
	if member.Role != model.LabRoleOwner {
		return nil, app_error.ErrNotOwner
	}

	page := query.Page
	if page == 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize == 0 {
		pageSize = 20
	}
	filter := repo.ListLabDocumentsFilter{
		Search:    strings.TrimSpace(query.Search),
		Status:    query.Status,
		SortBy:    query.SortBy,
		SortOrder: query.SortOrder,
		Offset:    (page - 1) * pageSize,
		Limit:     pageSize,
	}
	docs, total, err := s.repo.FindByLabID(ctx, labID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list lab documents: %w", err)
	}

	items := make([]dto.DocumentListItem, 0, len(docs))
	for i := range docs {
		items = append(items, toDocumentListItem(&docs[i]))
	}

	return &dto.ListMyDocumentsResponse{
		Documents: items,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

func (s *DocumentService) ListPendingDocuments(ctx context.Context, userID uint) (*dto.ListMyDocumentsResponse, error) {
	docs, total, err := s.repo.FindByUserIDAndStatus(ctx, userID, model.EnrichStatusNotStarted, 0, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending documents: %w", err)
	}

	items := make([]dto.DocumentListItem, 0, len(docs))
	for i := range docs {
		items = append(items, toDocumentListItem(&docs[i]))
	}

	return &dto.ListMyDocumentsResponse{
		Documents: items,
		Total:     total,
		Page:      1,
		PageSize:  50,
	}, nil
}

func (s *DocumentService) UpdateVisibility(ctx context.Context, docID, userID uint, req dto.UpdateVisibilityRequest) error {
	resolvedVis, resolvedLabID, err := s.resolveVisibility(ctx, userID, req.Visibility, req.LabID)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateVisibility(ctx, docID, userID, resolvedVis, resolvedLabID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return app_error.ErrNotDocumentOwner
		}
		return err
	}
	return nil
}

func (s *DocumentService) BatchUpdateVisibility(ctx context.Context, userID uint, req dto.BatchUpdateVisibilityRequest) (int64, error) {
	resolvedVis, resolvedLabID, err := s.resolveVisibility(ctx, userID, req.Visibility, req.LabID)
	if err != nil {
		return 0, err
	}

	updated, err := s.repo.BatchUpdateVisibility(ctx, req.DocIDs, userID, resolvedVis, resolvedLabID)
	if err != nil {
		return 0, fmt.Errorf("failed to batch update visibility: %w", err)
	}
	if updated != int64(len(req.DocIDs)) {
		return updated, app_error.ErrSomeDocsNotAccessible
	}
	return updated, nil
}

// UpdateMetadata patches user-editable metadata on a document the caller can manage
// (the uploader, or — for lab-visible docs — the owner of the lab).
// Only non-nil fields on the request are applied; a nil field means "leave as-is".
// The three fields are nullable columns, so clients can clear them by sending
// an explicit empty string / zero — callers send `null` via JSON to mean "no change".
func (s *DocumentService) UpdateMetadata(ctx context.Context, userID, docID uint, req dto.UpdateDocumentMetadataRequest) error {
	doc, err := s.repo.FindByID(ctx, docID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return app_error.ErrDocumentNotFound
		}
		return err
	}
	if ok, err := s.canManageDocument(ctx, userID, &doc); err != nil {
		return fmt.Errorf("failed to check document manage permission: %w", err)
	} else if !ok {
		return app_error.ErrNotDocumentOwner
	}

	patch := repo.DocumentMetadataPatch{
		Title: req.Title,
		Year:  req.Year,
		DOI:   req.DOI,
	}
	if err := s.repo.UpdateMetadata(ctx, docID, patch); err != nil {
		return err
	}
	return nil
}

// DeleteDocument soft-deletes a document the caller can manage (uploader or
// lab owner) and best-effort removes the S3 object. Dashboard stats are
// invalidated so the breakdown refreshes.
func (s *DocumentService) DeleteDocument(ctx context.Context, userID, docID uint) error {
	doc, err := s.repo.FindByID(ctx, docID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return app_error.ErrDocumentNotFound
		}
		return err
	}
	if ok, err := s.canManageDocument(ctx, userID, &doc); err != nil {
		return fmt.Errorf("failed to check document manage permission: %w", err)
	} else if !ok {
		return app_error.ErrNotDocumentOwner
	}

	if err := s.repo.DeleteByID(ctx, docID); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	// Best effort: remove the S3 object and any cached enrich status.
	// Failures here don't roll back the DB delete — the row is already gone.
	if err := s.storageClient.DeleteObject(ctx, doc.FileKey, true); err != nil {
		slog.Warn("Failed to delete S3 object for document", "docID", docID, "key", doc.FileKey, "err", err)
	}
	if _, err := s.cacheConn.Del(ctx, enrichStatusKey(docID)); err != nil {
		slog.Warn("Failed to clear enrich status cache", "docID", docID, "err", err)
	}
	// Invalidate dashboard stats for the original uploader (their breakdown changed).
	if _, err := s.cacheConn.Del(ctx, dashboardStatsKey(doc.UploadedByUserID)); err != nil {
		slog.Warn("Failed to invalidate dashboard stats cache", "userID", doc.UploadedByUserID, "err", err)
	}
	return nil
}

// downloadFilename returns the original upload filename. Files are stored
// byte-for-byte in their original format (the recommender converts to PDF
// only for embedding), so callers download exactly what they uploaded.
func downloadFilename(original string) string {
	return original
}

func toDocumentResponse(doc *model.Document, downloadURL string, likedByMe bool) *dto.DocumentResponse {
	authors := []string(doc.Authors)
	if authors == nil {
		authors = []string{}
	}
	tags := []string(doc.Tags)
	if tags == nil {
		tags = []string{}
	}
	var labName *string
	if doc.Lab != nil {
		name := doc.Lab.Name
		labName = &name
	}
	var uploaderName *string
	if doc.UploadedBy.ID != 0 {
		name := doc.UploadedBy.Username
		uploaderName = &name
	}
	return &dto.DocumentResponse{
		ID:                 doc.ID,
		Title:              doc.Title,
		OriginalFileName:   doc.OriginalFileName,
		FileSize:           doc.FileSize,
		ContentType:        doc.ContentType,
		Year:               doc.Year,
		DOI:                doc.DOI,
		EnrichStatus:       doc.EnrichStatus,
		Visibility:         doc.Visibility,
		LabID:              doc.LabID,
		LabName:            labName,
		Authors:            authors,
		Summary:            doc.Summary,
		Tags:               tags,
		ViewCount:          doc.ViewCount,
		LikeCount:          doc.LikeCount,
		LikedByMe:          likedByMe,
		UploadedByUserID:   doc.UploadedByUserID,
		UploadedByUsername: uploaderName,
		DownloadURL:        downloadURL,
		CreatedAt:          doc.CreatedAt,
	}
}

func toDocumentListItem(doc *model.Document) dto.DocumentListItem {
	var labName *string
	if doc.Lab != nil {
		name := doc.Lab.Name
		labName = &name
	}
	// UploadedBy is only preloaded for the lab-list endpoint; the personal
	// list endpoint omits it because every row is the caller's own upload.
	var uploaderName *string
	if doc.UploadedBy.ID != 0 {
		name := doc.UploadedBy.Username
		uploaderName = &name
	}
	return dto.DocumentListItem{
		ID:                 doc.ID,
		Title:              doc.Title,
		OriginalFileName:   doc.OriginalFileName,
		FileSize:           doc.FileSize,
		EnrichStatus:       doc.EnrichStatus,
		Visibility:         doc.Visibility,
		LabID:              doc.LabID,
		LabName:            labName,
		UploadedByUserID:   doc.UploadedByUserID,
		UploadedByUsername: uploaderName,
		CreatedAt:          doc.CreatedAt,
	}
}
