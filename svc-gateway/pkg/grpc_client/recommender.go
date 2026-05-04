package grpc_client

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "gateway/internal/pb/recommender/v1"
)

// RecommenderClient wraps the gRPC connection to svc-recommender.
type RecommenderClient struct {
	conn   *grpc.ClientConn
	client pb.RecommenderServiceClient
}

// NewRecommenderClient dials the recommender service at addr (e.g. "localhost:50051")
// and returns a ready-to-use client. Call Close when done.
func NewRecommenderClient(addr string) (*RecommenderClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial recommender at %s: %w", addr, err)
	}

	slog.Info("connected to recommender grpc service", "addr", addr)

	return &RecommenderClient{
		conn:   conn,
		client: pb.NewRecommenderServiceClient(conn),
	}, nil
}

// Health calls the Health RPC on the recommender service.
// A short deadline is applied so a slow peer does not block the gateway's own health route.
func (r *RecommenderClient) Health(ctx context.Context) (*pb.HealthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return r.client.Health(ctx, &pb.HealthRequest{})
}

// EnrichDocument sends a fire-and-forget enrichment request to the Python service.
// The Python service ACKs immediately and processes asynchronously. contentType
// is the MIME of the uploaded file — the recommender uses it to pick a
// passthrough or office-conversion strategy before feeding bytes to Gemini.
func (r *RecommenderClient) EnrichDocument(ctx context.Context, docID uint64, fileKey, contentType string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	resp, err := r.client.EnrichDocument(ctx, &pb.EnrichDocumentRequest{
		DocId:       docID,
		FileKey:     fileKey,
		ContentType: contentType,
	})
	if err != nil {
		return false, err
	}
	return resp.Accepted, nil
}

// TranslateTextStream opens a server-streaming call for text translation.
// The caller must consume and close the returned stream.
func (r *RecommenderClient) TranslateTextStream(ctx context.Context, text, targetLanguage string) (pb.RecommenderService_TranslateTextClient, error) {
	return r.client.TranslateText(ctx, &pb.TranslateTextRequest{
		Text:           text,
		TargetLanguage: targetLanguage,
	})
}

// SemanticSearch calls the Python service to embed a query and search for similar documents.
// Uses a generous timeout since query embedding involves a Gemini API call.
func (r *RecommenderClient) SemanticSearch(ctx context.Context, query string, userID, labID uint64, limit uint32) (*pb.SemanticSearchResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return r.client.SemanticSearch(ctx, &pb.SemanticSearchRequest{
		Query:  query,
		UserId: userID,
		LabId:  labID,
		Limit:  limit,
	})
}

// RecommendSimilar asks the Python service for documents most similar to a
// source doc. Unlike SemanticSearch this doesn't require an embedding call
// (the source doc's stored embedding is reused), so a short timeout is fine.
func (r *RecommenderClient) RecommendSimilar(ctx context.Context, docID, userID, labID uint64, limit uint32) (*pb.RecommendSimilarResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.client.RecommendSimilar(ctx, &pb.RecommendSimilarRequest{
		DocId:  docID,
		UserId: userID,
		LabId:  labID,
		Limit:  limit,
	})
}

// RecommendForUser asks the Python service for a personalized feed built from
// the caller's like / view / search-query history. The recent_queries list may
// trigger Gemini calls (one per uncached query), so a generous timeout is
// applied — same as SemanticSearch.
func (r *RecommenderClient) RecommendForUser(
	ctx context.Context,
	userID, labID uint64,
	limit uint32,
	likedDocIDs, viewedDocIDs []uint64,
	recentQueries []string,
) (*pb.RecommendForUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return r.client.RecommendForUser(ctx, &pb.RecommendForUserRequest{
		UserId:        userID,
		LabId:         labID,
		Limit:         limit,
		LikedDocIds:   likedDocIDs,
		ViewedDocIds:  viewedDocIDs,
		RecentQueries: recentQueries,
	})
}

// Close releases the underlying gRPC connection.
func (r *RecommenderClient) Close() error {
	return r.conn.Close()
}
