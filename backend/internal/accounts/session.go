package accounts

import (
	"time"

	"ostos/backend/internal/platform"
)

type Session struct {
	ID                  platform.ID  `gorm:"column:id;primaryKey;autoIncrement:false"`
	UserID              platform.ID  `gorm:"column:user_id"`
	RefreshTokenHash    []byte       `gorm:"column:refresh_token_hash;type:binary(32)"`
	DeviceName          *string      `gorm:"column:device_name"`
	CreatedAt           time.Time    `gorm:"column:created_at"`
	LastUsedAt          time.Time    `gorm:"column:last_used_at"`
	ExpiresAt           time.Time    `gorm:"column:expires_at"`
	RevokedAt           *time.Time   `gorm:"column:revoked_at"`
	ReplacedBySessionID *platform.ID `gorm:"column:replaced_by_session_id"`
}

func (Session) TableName() string { return "auth_sessions" }

type TokenPurpose string

const (
	TokenPurposeVerifyEmail   TokenPurpose = "verify_email"
	TokenPurposeResetPassword TokenPurpose = "reset_password"
)

type AccountToken struct {
	ID              platform.ID  `gorm:"column:id;primaryKey;autoIncrement:false"`
	UserID          platform.ID  `gorm:"column:user_id"`
	Purpose         TokenPurpose `gorm:"column:purpose"`
	BearerTokenHash []byte       `gorm:"column:bearer_token_hash;type:binary(32)"`
	ExpiresAt       time.Time    `gorm:"column:expires_at"`
	UsedAt          *time.Time   `gorm:"column:used_at"`
	CreatedAt       time.Time    `gorm:"column:created_at"`
}

func (AccountToken) TableName() string { return "account_tokens" }

type EmailTemplate string

const (
	EmailTemplateVerifyEmail   EmailTemplate = "verify_email"
	EmailTemplateResetPassword EmailTemplate = "reset_password"
)

type EmailOutbox struct {
	ID                    platform.ID   `gorm:"column:id;primaryKey;autoIncrement:false"`
	UserID                platform.ID   `gorm:"column:user_id"`
	Template              EmailTemplate `gorm:"column:template"`
	Recipient             string        `gorm:"column:recipient"`
	TemplateData          []byte        `gorm:"column:template_data;type:json"`
	EncryptedTokenPayload []byte        `gorm:"column:encrypted_token_payload"`
	DeliveryState         string        `gorm:"column:delivery_state"`
	AttemptCount          uint16        `gorm:"column:attempt_count"`
	NextAttemptAt         time.Time     `gorm:"column:next_attempt_at"`
	SentAt                *time.Time    `gorm:"column:sent_at"`
	CreatedAt             time.Time     `gorm:"column:created_at"`
	UpdatedAt             time.Time     `gorm:"column:updated_at"`
}

func (EmailOutbox) TableName() string { return "email_outbox" }

type AuditEvent struct {
	ID          platform.ID  `gorm:"column:id;primaryKey;autoIncrement:false"`
	ActorUserID *platform.ID `gorm:"column:actor_user_id"`
	ActionCode  string       `gorm:"column:action_code"`
	TargetType  string       `gorm:"column:target_type"`
	TargetID    *platform.ID `gorm:"column:target_id"`
	Metadata    []byte       `gorm:"column:metadata;type:json"`
	RequestID   string       `gorm:"column:request_id"`
	CreatedAt   time.Time    `gorm:"column:created_at"`
}

func (AuditEvent) TableName() string { return "audit_events" }
