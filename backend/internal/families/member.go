package families

import (
	"time"

	"ostos/backend/internal/platform"
)

type Role string

const (
	RoleOwner   Role = "owner"
	RoleManager Role = "manager"
	RoleMember  Role = "member"
)

type Member struct {
	ID        platform.ID `gorm:"column:id;primaryKey;autoIncrement:false"`
	FamilyID  platform.ID `gorm:"column:family_id"`
	UserID    platform.ID `gorm:"column:user_id"`
	Role      Role        `gorm:"column:role"`
	CreatedAt time.Time   `gorm:"column:created_at"`
	UpdatedAt time.Time   `gorm:"column:updated_at"`
}

func (Member) TableName() string { return "family_members" }

type MemberView struct {
	ID          *platform.ID `json:"id,omitempty"`
	UserID      platform.ID  `json:"user_id"`
	Username    string       `json:"username"`
	DisplayName string       `json:"display_name"`
	Role        Role         `json:"role"`
	JoinedAt    time.Time    `json:"joined_at"`
}
