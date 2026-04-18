// Package reenrich provides a background scheduler that retries stuck enrichments.
// It periodically scans for documents whose enrich_status is still "not_started"
// well after upload (i.e. the recommender's in-process retry budget was exhausted)
// and re-invokes EnrichDocument against the Python microservice.
package reenrich

import (
	"context"
	"log/slog"
	"time"

	"gateway/internal/repo"
	"gateway/pkg/grpc_client"
)

const (
	// gracePeriod skips docs uploaded within this window, to avoid racing with
	// the recommender's own in-process retry burst.
	gracePeriod = 15 * time.Minute
	// batchSize caps how many stale docs are rescheduled per tick.
	batchSize = 100
)

type Scheduler struct {
	repo              repo.DocumentRepository
	recommenderClient *grpc_client.RecommenderClient
	interval          time.Duration
	stop              chan struct{}
}

func NewScheduler(repo repo.DocumentRepository, recommenderClient *grpc_client.RecommenderClient, interval time.Duration) *Scheduler {
	return &Scheduler{
		repo:              repo,
		recommenderClient: recommenderClient,
		interval:          interval,
		stop:              make(chan struct{}),
	}
}

// Start runs the scheduler loop. This is a blocking call; the caller owns
// the goroutine (e.g. `go scheduler.Start()`). Call Stop to terminate.
func (s *Scheduler) Start() {
	slog.Info("re-enrich scheduler started", "interval", s.interval)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			slog.Info("re-enrich scheduler stopped")
			return
		case <-ticker.C:
			s.runOnce()
		}
	}
}

// Stop signals the scheduler loop to exit after the current iteration.
func (s *Scheduler) Stop() {
	close(s.stop)
}

// runOnce finds stale not_started docs and re-queues them with the recommender.
// One slow call should not block the next tick, so each run has a bounded timeout.
func (s *Scheduler) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), s.interval/2)
	defer cancel()

	olderThan := time.Now().Add(-gracePeriod)
	docs, err := s.repo.FindStaleNotStarted(ctx, olderThan, batchSize)
	if err != nil {
		slog.Error("re-enrich scheduler: failed to query stale docs", "err", err)
		return
	}
	if len(docs) == 0 {
		return
	}

	slog.Info("re-enrich scheduler: rescheduling stale docs", "count", len(docs))
	for i := range docs {
		doc := &docs[i]
		if _, err := s.recommenderClient.EnrichDocument(ctx, uint64(doc.ID), doc.FileKey); err != nil {
			slog.Warn("re-enrich: EnrichDocument call failed", "docID", doc.ID, "err", err)
			continue
		}
	}
}
