package model

import "gorm.io/gorm"

// SearchHistory records each successful semantic search a user performs,
// for the "recent searches" UI and for future analytics.
type SearchHistory struct {
	gorm.Model
	UserID      uint   `gorm:"not null;index:idx_search_history_user_created,priority:1"`
	LabID       *uint  `gorm:"index"`
	Query       string `gorm:"not null;type:text"`
	ResultCount int    `gorm:"not null;default:0"`
}
