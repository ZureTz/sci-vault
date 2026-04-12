package model

import "gorm.io/gorm"

const (
	LabRoleOwner  = "owner"
	LabRoleMember = "member"
)

type Lab struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description *string
	InviteCode  string      `gorm:"uniqueIndex;not null"`
	OwnerID     uint        `gorm:"not null;index"`
	Owner       User        `gorm:"foreignKey:OwnerID"`
	Members     []LabMember `gorm:"foreignKey:LabID"`
}

type LabMember struct {
	gorm.Model
	LabID  uint   `gorm:"not null;index"`
	UserID uint   `gorm:"not null;index"`
	Role   string `gorm:"not null;default:'member'"`
	User   User   `gorm:"foreignKey:UserID"`
	Lab    Lab    `gorm:"foreignKey:LabID"`
}
