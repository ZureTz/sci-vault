package database

import (
	"fmt"
	"log/slog"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"gateway/internal/config"
)

func getDSN(c *config.DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode, c.TimeZone,
	)
}

func New(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(getDSN(cfg)), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Info("database connected", "host", cfg.Host, "dbname", cfg.DBName)
	return db, nil
}

// HasCustomIndexes is implemented by models that need raw-SQL indexes which
// AutoMigrate cannot express (e.g. partial unique indexes, pgvector HNSW).
// Each returned statement must be idempotent (CREATE INDEX IF NOT EXISTS ...).
type HasCustomIndexes interface {
	CustomIndexes() []string
}

// Setup runs one-time database setup: installs required extensions, auto-migrates
// the given models, then applies any custom indexes declared by those models.
func Setup(db *gorm.DB, models ...any) error {
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		return fmt.Errorf("failed to create pgvector extension: %w", err)
	}
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto migrate database: %w", err)
	}
	return ensureCustomIndexes(db, models)
}

func ensureCustomIndexes(db *gorm.DB, models []any) error {
	for _, m := range models {
		indexer, ok := m.(HasCustomIndexes)
		if !ok {
			continue
		}
		for _, stmt := range indexer.CustomIndexes() {
			if err := db.Exec(stmt).Error; err != nil {
				return fmt.Errorf("failed to create custom index for %T: %w", m, err)
			}
		}
	}
	return nil
}
