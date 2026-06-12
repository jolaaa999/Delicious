package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"size:64;not null;uniqueIndex:uk_users_username" json:"username"`
	Email        *string        `gorm:"size:128;uniqueIndex:uk_users_email" json:"email,omitempty"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	Nickname     string         `gorm:"size:64;not null;default:''" json:"nickname"`
	AvatarURL    *string        `gorm:"size:512" json:"avatar_url,omitempty"`
	Status       int8           `gorm:"not null;default:1" json:"status"`
	CreatedAt    time.Time      `gorm:"precision:3" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"precision:3" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index:idx_users_deleted_at" json:"-"`

	Recipes []MyRecipe `gorm:"foreignKey:UserID" json:"recipes,omitempty"`
}

func (User) TableName() string { return "users" }
