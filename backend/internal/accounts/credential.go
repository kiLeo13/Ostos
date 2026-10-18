package accounts

import (
	"time"

	"ostos/backend/internal/platform"
)

type Credential struct {
	ID                platform.ID `gorm:"column:id;primaryKey;autoIncrement:false"`
	UserID            platform.ID `gorm:"column:user_id"`
	PasswordHash      string      `gorm:"column:password_hash"`
	PasswordChangedAt time.Time   `gorm:"column:password_changed_at"`
	CreatedAt         time.Time   `gorm:"column:created_at"`
	UpdatedAt         time.Time   `gorm:"column:updated_at"`
}

func (Credential) TableName() string { return "user_credentials" }
