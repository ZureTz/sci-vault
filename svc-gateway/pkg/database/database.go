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

// Setup runs one-time database setup: installs required extensions and auto-migrates the given models.
func Setup(db *gorm.DB, models ...any) error {
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		return fmt.Errorf("failed to create pgvector extension: %w", err)
	}
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto migrate database: %w", err)
	}
	return nil
}
