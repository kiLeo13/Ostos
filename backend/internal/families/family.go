package families

import (
	"time"

	"ostos/backend/internal/platform"
)

type Family struct {
	ID        platform.ID `gorm:"column:id;primaryKey;autoIncrement:false" json:"id"`
	Name      string      `gorm:"column:name" json:"name"`
	OwnerID   platform.ID `gorm:"column:owner_id" json:"owner_id"`
	CreatedAt time.Time   `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time   `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time  `gorm:"column:deleted_at" json:"-"`
}

func (Family) TableName() string { return "families" }

type Summary struct {
	ID        platform.ID `json:"id"`
	Name      string      `json:"name"`
	Role      Role        `json:"role"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
