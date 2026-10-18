package accounts

import (
	"context"
	"time"

	"ostos/backend/internal/platform"
)

type Repository interface {
	WithinTransaction(context.Context, func(Repository) error) error
	CreateUser(context.Context, *User) error
	CreateCredential(context.Context, *Credential) error
	CreateSession(context.Context, *Session) error
	CreateAccountToken(context.Context, *AccountToken) error
	CreateEmailOutbox(context.Context, *EmailOutbox) error
	CreateAuditEvent(context.Context, *AuditEvent) error
	FindUserByID(context.Context, platform.ID) (*User, error)
	FindUserByEmail(context.Context, string) (*User, error)
	FindUserAndCredentialByIdentifier(context.Context, string) (*User, *Credential, error)
	FindCredentialByUserID(context.Context, platform.ID) (*Credential, error)
	FindSessionByHashForUpdate(context.Context, []byte) (*Session, error)
	FindAccountTokenForUpdate(context.Context, []byte, TokenPurpose) (*AccountToken, error)
	UpdateUserProfile(context.Context, platform.ID, string, string, time.Time) error
	MarkEmailVerified(context.Context, platform.ID, time.Time) error
	UpdatePassword(context.Context, platform.ID, string, time.Time) error
	MarkAccountTokenUsed(context.Context, platform.ID, time.Time) error
	InvalidateAccountTokens(context.Context, platform.ID, TokenPurpose, time.Time) error
	RotateSession(context.Context, platform.ID, *Session, time.Time) error
	RevokeSession(context.Context, platform.ID, platform.ID, time.Time) error
	RevokeSessionChain(context.Context, platform.ID, platform.ID, time.Time) error
	RevokeAllSessions(context.Context, platform.ID, time.Time) error
	SoftDeleteUserIfNoOwnedFamilies(context.Context, platform.ID, time.Time) error
	ClaimEmail(context.Context, time.Time, time.Time) (*EmailOutbox, error)
	MarkEmailSent(context.Context, platform.ID, time.Time) error
	MarkEmailFailed(context.Context, platform.ID, uint16, time.Time, bool) error
}
