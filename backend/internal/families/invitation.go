package families

import (
	"time"

	"ostos/backend/internal/platform"
)

type Invitation struct {
	ID              platform.ID `gorm:"column:id;primaryKey;autoIncrement:false" json:"id"`
	FamilyID        platform.ID `gorm:"column:family_id" json:"family_id"`
	InvitedUserID   platform.ID `gorm:"column:invited_user_id" json:"invited_user_id"`
	InvitedByUserID platform.ID `gorm:"column:invited_by_user_id" json:"invited_by_user_id"`
	ExpiresAt       time.Time   `gorm:"column:expires_at" json:"expires_at"`
	CreatedAt       time.Time   `gorm:"column:created_at" json:"created_at"`
}

func (Invitation) TableName() string { return "family_invitations" }

type InvitationView struct {
	ID         platform.ID `json:"id"`
	FamilyID   platform.ID `json:"family_id"`
	FamilyName string      `json:"family_name"`
	ExpiresAt  time.Time   `json:"expires_at"`
	CreatedAt  time.Time   `json:"created_at"`
}

type SentInvitationView struct {
	ID          platform.ID `json:"id"`
	UserID      platform.ID `json:"user_id"`
	Username    string      `json:"username"`
	DisplayName string      `json:"display_name"`
	ExpiresAt   time.Time   `json:"expires_at"`
	CreatedAt   time.Time   `json:"created_at"`
}

type AuditEvent struct {
	ID          platform.ID  `gorm:"column:id;primaryKey;autoIncrement:false"`
	ActorUserID *platform.ID `gorm:"column:actor_user_id"`
	FamilyID    *platform.ID `gorm:"column:family_id"`
	ActionCode  string       `gorm:"column:action_code"`
	TargetType  string       `gorm:"column:target_type"`
	TargetID    *platform.ID `gorm:"column:target_id"`
	Metadata    []byte       `gorm:"column:metadata;type:json"`
	RequestID   string       `gorm:"column:request_id"`
	CreatedAt   time.Time    `gorm:"column:created_at"`
}

func (AuditEvent) TableName() string { return "audit_events" }
