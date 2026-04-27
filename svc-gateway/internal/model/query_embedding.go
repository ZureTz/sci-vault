package model

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

// QueryEmbedding is a persistent cache of search-query → embedding mappings.
// The mapping is deterministic for a fixed embedding model, so once we've paid
// the Gemini token cost for a query we never want to pay it again. Redis fronts
// this table for hot reads (24h TTL); this table is the durable backstop.
//
// The gateway owns the schema (so AutoMigrate creates the table) but never
// reads or writes it — the recommender service is the only consumer.
type QueryEmbedding struct {
	QueryHash  []byte          `gorm:"type:bytea;primaryKey"` // sha256(query) raw bytes
	Query      string          `gorm:"type:text;not null"`    // kept for debugging / observability
	Embedding  pgvector.Vector `gorm:"type:vector(768);not null"`
	CreatedAt  time.Time
	LastUsedAt time.Time `gorm:"index"`
}
