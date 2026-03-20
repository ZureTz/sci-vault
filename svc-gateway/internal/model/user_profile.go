package model

import "gorm.io/gorm"

// UserProfile stores the public-facing profile information for a user (1:1 with User).
type UserProfile struct {
	gorm.Model
	UserID    uint   `gorm:"uniqueIndex;not null;constraint:OnDelete:CASCADE"`
	User      User   `gorm:"foreignKey:UserID"`
	Nickname  string `gorm:"size:50"`
	Bio       string `gorm:"size:500"`
	AvatarURL string `gorm:"size:1024"`
	Website   string `gorm:"size:255"`
	Location  string `gorm:"size:100"`
}
