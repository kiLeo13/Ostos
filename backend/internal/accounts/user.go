package accounts

import (
	"strings"
	"time"

	"ostos/backend/internal/platform"
)

const (
	MaxDisplayNameLength = 100
	MaxDeviceNameLength  = 100
)

type User struct {
	ID              platform.ID `gorm:"column:id;primaryKey;autoIncrement:false" json:"id"`
	Email           string      `gorm:"column:email" json:"email"`
	Username        string      `gorm:"column:username" json:"username"`
	DisplayName     string      `gorm:"column:display_name" json:"display_name"`
	EmailVerifiedAt *time.Time  `gorm:"column:email_verified_at" json:"-"`
	CreatedAt       time.Time   `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time   `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt       *time.Time  `gorm:"column:deleted_at" json:"-"`
}

func (User) TableName() string { return "users" }

type PublicUser struct {
	ID            platform.ID `json:"id"`
	Email         string      `json:"email"`
	Username      string      `json:"username"`
	DisplayName   string      `json:"display_name"`
	EmailVerified bool        `json:"email_verified"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

func (user User) Public() PublicUser {
	return PublicUser{
		ID: user.ID, Email: user.Email, Username: user.Username,
		DisplayName: user.DisplayName, EmailVerified: user.EmailVerifiedAt != nil,
		CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}
}

func CanonicalEmail(value string) string { return strings.ToLower(strings.TrimSpace(value)) }
