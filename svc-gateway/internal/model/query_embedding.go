package model

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

// QueryEmbedding is a persistent cache of (text, task_type) → embedding
// mappings. The mapping is deterministic for a fixed embedding model + task
// type, so once we've paid the Gemini token cost we never want to pay it
// again. Redis fronts this table for hot reads (24h TTL); this table is the
// durable backstop.
//
// task_type is part of the primary key because Gemini's RETRIEVAL_QUERY and
// RETRIEVAL_DOCUMENT produce intentionally asymmetric vectors that live in
// incompatible spaces — the same string under two task types is two distinct
// embeddings and must not collide.
//
// The gateway owns the schema (so AutoMigrate creates the table) but never
// reads or writes it — the recommender service is the only consumer.
type QueryEmbedding struct {
	QueryHash  []byte          `gorm:"type:bytea;primaryKey"`       // sha256(text) raw bytes
	TaskType   string          `gorm:"type:varchar(32);primaryKey"` // RETRIEVAL_QUERY | RETRIEVAL_DOCUMENT
	Query      string          `gorm:"type:text;not null"`          // kept for debugging / observability
	Embedding  pgvector.Vector `gorm:"type:vector(768);not null"`
	CreatedAt  time.Time
	LastUsedAt time.Time `gorm:"index"`
}
