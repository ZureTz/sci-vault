package model

import "gorm.io/gorm"

// UserAvatar records each avatar upload for a user, preserving full upload history.
type UserAvatar struct {
	gorm.Model
	UserID    uint   `gorm:"not null;index;constraint:OnDelete:CASCADE"`
	User      User   `gorm:"foreignKey:UserID"`
	AvatarURL string `gorm:"not null"`
}
