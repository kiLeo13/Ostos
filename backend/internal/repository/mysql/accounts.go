package mysql

import (
	"context"
	"errors"
	"strings"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"ostos/backend/internal/accounts"
	"ostos/backend/internal/families"
	"ostos/backend/internal/platform"
)

type AccountRepository struct{ database *gorm.DB }

func NewAccountRepository(database *gorm.DB) *AccountRepository {
	return &AccountRepository{database: database}
}

func (repository *AccountRepository) WithinTransaction(ctx context.Context, operation func(accounts.Repository) error) error {
	return WithinTransaction(ctx, repository.database, nil, func(transaction *gorm.DB) error {
		return operation(&AccountRepository{database: transaction})
	})
}

func (repository *AccountRepository) CreateUser(ctx context.Context, user *accounts.User) error {
	err := repository.database.WithContext(ctx).Create(user).Error
	if duplicateKey(err, "uq_users_email") {
		return accounts.ErrEmailTaken
	}
	if duplicateKey(err, "uq_users_username") {
		return accounts.ErrUsernameTaken
	}
	return err
}

func (repository *AccountRepository) CreateCredential(ctx context.Context, credential *accounts.Credential) error {
	return repository.database.WithContext(ctx).Create(credential).Error
}

func (repository *AccountRepository) CreateSession(ctx context.Context, session *accounts.Session) error {
	return repository.database.WithContext(ctx).Create(session).Error
}

func (repository *AccountRepository) CreateAccountToken(ctx context.Context, token *accounts.AccountToken) error {
	return repository.database.WithContext(ctx).Create(token).Error
}

func (repository *AccountRepository) CreateEmailOutbox(ctx context.Context, email *accounts.EmailOutbox) error {
	return repository.database.WithContext(ctx).Create(email).Error
}

func (repository *AccountRepository) CreateAuditEvent(ctx context.Context, event *accounts.AuditEvent) error {
	return repository.database.WithContext(ctx).Create(event).Error
}

func (repository *AccountRepository) FindUserByID(ctx context.Context, id platform.ID) (*accounts.User, error) {
	var user accounts.User
	err := repository.database.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, accounts.ErrUserNotFound
	}
	return &user, err
}

func (repository *AccountRepository) FindUserByEmail(ctx context.Context, email string) (*accounts.User, error) {
	var user accounts.User
	err := repository.database.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, accounts.ErrUserNotFound
	}
	return &user, err
}

func (repository *AccountRepository) FindUserAndCredentialByIdentifier(ctx context.Context, identifier string) (*accounts.User, *accounts.Credential, error) {
	type row struct {
		accounts.User
		CredentialID        platform.ID `gorm:"column:credential_id"`
		PasswordHash        string      `gorm:"column:password_hash"`
		PasswordChangedAt   time.Time   `gorm:"column:password_changed_at"`
		CredentialCreatedAt time.Time   `gorm:"column:credential_created_at"`
		CredentialUpdatedAt time.Time   `gorm:"column:credential_updated_at"`
	}
	var result row
	err := repository.database.WithContext(ctx).Table("users AS u").
		Select("u.*, c.id AS credential_id, c.password_hash, c.password_changed_at, c.created_at AS credential_created_at, c.updated_at AS credential_updated_at").
		Joins("JOIN user_credentials AS c ON c.user_id = u.id").
		Where("u.deleted_at IS NULL AND (u.email = ? OR u.username = ?)", identifier, identifier).Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, accounts.ErrUserNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	credential := &accounts.Credential{ID: result.CredentialID, UserID: result.ID, PasswordHash: result.PasswordHash,
		PasswordChangedAt: result.PasswordChangedAt, CreatedAt: result.CredentialCreatedAt, UpdatedAt: result.CredentialUpdatedAt}
	return &result.User, credential, nil
}

func (repository *AccountRepository) FindCredentialByUserID(ctx context.Context, userID platform.ID) (*accounts.Credential, error) {
	var credential accounts.Credential
	err := repository.database.WithContext(ctx).Where("user_id = ?", userID).Take(&credential).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, accounts.ErrUserNotFound
	}
	return &credential, err
}

func (repository *AccountRepository) FindSessionByHashForUpdate(ctx context.Context, hash []byte) (*accounts.Session, error) {
	var session accounts.Session
	err := repository.database.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("refresh_token_hash = ?", hash).Take(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, accounts.ErrInvalidRefreshToken
	}
	return &session, err
}

func (repository *AccountRepository) FindAccountTokenForUpdate(ctx context.Context, hash []byte, purpose accounts.TokenPurpose) (*accounts.AccountToken, error) {
	var token accounts.AccountToken
	err := repository.database.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("bearer_token_hash = ? AND purpose = ?", hash, purpose).Take(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, accounts.ErrInvalidAccountToken
	}
	return &token, err
}

func (repository *AccountRepository) UpdateUserProfile(ctx context.Context, userID platform.ID, username, displayName string, now time.Time) error {
	result := repository.database.WithContext(ctx).Model(&accounts.User{}).Where("id = ? AND deleted_at IS NULL", userID).
		Updates(map[string]any{"username": username, "display_name": displayName, "updated_at": now})
	if duplicateKey(result.Error, "uq_users_username") {
		return accounts.ErrUsernameTaken
	}
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return accounts.ErrUserNotFound
	}
	return nil
}

func (repository *AccountRepository) MarkEmailVerified(ctx context.Context, userID platform.ID, now time.Time) error {
	return repository.database.WithContext(ctx).Model(&accounts.User{}).Where("id = ? AND deleted_at IS NULL", userID).
		Updates(map[string]any{"email_verified_at": now, "updated_at": now}).Error
}

func (repository *AccountRepository) UpdatePassword(ctx context.Context, userID platform.ID, hash string, now time.Time) error {
	result := repository.database.WithContext(ctx).Model(&accounts.Credential{}).Where("user_id = ?", userID).
		Updates(map[string]any{"password_hash": hash, "password_changed_at": now, "updated_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return accounts.ErrUserNotFound
	}
	return nil
}

func (repository *AccountRepository) MarkAccountTokenUsed(ctx context.Context, tokenID platform.ID, now time.Time) error {
	return repository.database.WithContext(ctx).Model(&accounts.AccountToken{}).Where("id = ? AND used_at IS NULL", tokenID).Update("used_at", now).Error
}

func (repository *AccountRepository) InvalidateAccountTokens(ctx context.Context, userID platform.ID, purpose accounts.TokenPurpose, now time.Time) error {
	return repository.database.WithContext(ctx).Model(&accounts.AccountToken{}).
		Where("user_id = ? AND purpose = ? AND used_at IS NULL", userID, purpose).Update("used_at", now).Error
}

func (repository *AccountRepository) RotateSession(ctx context.Context, oldSessionID platform.ID, replacement *accounts.Session, now time.Time) error {
	if err := repository.CreateSession(ctx, replacement); err != nil {
		return err
	}
	result := repository.database.WithContext(ctx).Model(&accounts.Session{}).
		Where("id = ? AND revoked_at IS NULL AND replaced_by_session_id IS NULL", oldSessionID).
		Updates(map[string]any{"revoked_at": now, "last_used_at": now, "replaced_by_session_id": replacement.ID})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return accounts.ErrInvalidRefreshToken
	}
	return nil
}

func (repository *AccountRepository) RevokeSession(ctx context.Context, userID, sessionID platform.ID, now time.Time) error {
	return repository.database.WithContext(ctx).Model(&accounts.Session{}).
		Where("id = ? AND user_id = ? AND revoked_at IS NULL", sessionID, userID).Update("revoked_at", now).Error
}

func (repository *AccountRepository) RevokeSessionChain(ctx context.Context, userID, rootSessionID platform.ID, now time.Time) error {
	var sessionIDs []platform.ID
	err := repository.database.WithContext(ctx).Raw(`
		WITH RECURSIVE session_chain AS (
			SELECT id, replaced_by_session_id FROM auth_sessions WHERE id = ? AND user_id = ?
			UNION ALL
			SELECT child.id, child.replaced_by_session_id FROM auth_sessions child
			JOIN session_chain parent ON child.id = parent.replaced_by_session_id
			WHERE child.user_id = ?
		)
		SELECT id FROM session_chain`, rootSessionID, userID, userID).Scan(&sessionIDs).Error
	if err != nil { return err }
	if len(sessionIDs) == 0 { return accounts.ErrInvalidRefreshToken }
	return repository.database.WithContext(ctx).Model(&accounts.Session{}).
		Where("user_id = ? AND id IN ? AND revoked_at IS NULL", userID, sessionIDs).Update("revoked_at", now).Error
}

func (repository *AccountRepository) RevokeAllSessions(ctx context.Context, userID platform.ID, now time.Time) error {
	return repository.database.WithContext(ctx).Model(&accounts.Session{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).Update("revoked_at", now).Error
}

func (repository *AccountRepository) SoftDeleteUserIfNoOwnedFamilies(ctx context.Context, userID platform.ID, now time.Time) error {
	var user accounts.User
	if err := repository.database.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND deleted_at IS NULL", userID).Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return accounts.ErrUserNotFound
		}
		return err
	}
	var owned int64
	if err := repository.database.WithContext(ctx).Table("families").Where("owner_id = ? AND deleted_at IS NULL", userID).Count(&owned).Error; err != nil {
		return err
	}
	if owned != 0 {
		return accounts.ErrOwnsFamilies
	}
	if err := repository.database.WithContext(ctx).Where("user_id = ?", userID).Delete(&families.Member{}).Error; err != nil {
		return err
	}
	if err := repository.database.WithContext(ctx).Where("invited_user_id = ? OR invited_by_user_id = ?", userID, userID).Delete(&families.Invitation{}).Error; err != nil {
		return err
	}
	if err := repository.database.WithContext(ctx).Model(&accounts.AccountToken{}).
		Where("user_id = ? AND used_at IS NULL", userID).Update("used_at", now).Error; err != nil {
		return err
	}
	if err := repository.database.WithContext(ctx).Model(&accounts.EmailOutbox{}).
		Where("user_id = ? AND delivery_state IN ('pending', 'processing')", userID).
		Updates(map[string]any{"delivery_state": "dead", "encrypted_token_payload": nil, "updated_at": now}).Error; err != nil {
		return err
	}
	return repository.database.WithContext(ctx).Model(&accounts.User{}).Where("id = ?", userID).
		Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error
}

func (repository *AccountRepository) ClaimEmail(ctx context.Context, now, leaseUntil time.Time) (claimed *accounts.EmailOutbox, err error) {
	err = repository.WithinTransaction(ctx, func(transaction accounts.Repository) error {
		concrete := transaction.(*AccountRepository)
		var email accounts.EmailOutbox
		findErr := concrete.database.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("delivery_state IN ('pending', 'processing') AND next_attempt_at <= ?", now).
			Order("next_attempt_at, id").Take(&email).Error
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			return accounts.ErrNoPendingEmail
		}
		if findErr != nil {
			return findErr
		}
		if updateErr := concrete.database.WithContext(ctx).Model(&accounts.EmailOutbox{}).Where("id = ?", email.ID).
			Updates(map[string]any{"delivery_state": "processing", "next_attempt_at": leaseUntil, "updated_at": now}).Error; updateErr != nil {
			return updateErr
		}
		email.DeliveryState, email.NextAttemptAt, email.UpdatedAt = "processing", leaseUntil, now
		claimed = &email
		return nil
	})
	return claimed, err
}

func (repository *AccountRepository) MarkEmailSent(ctx context.Context, id platform.ID, now time.Time) error {
	return repository.database.WithContext(ctx).Model(&accounts.EmailOutbox{}).Where("id = ? AND delivery_state = 'processing'", id).
		Updates(map[string]any{"delivery_state": "sent", "sent_at": now, "encrypted_token_payload": nil, "updated_at": now}).Error
}

func (repository *AccountRepository) MarkEmailFailed(ctx context.Context, id platform.ID, attempts uint16, retryAt time.Time, dead bool) error {
	state := "pending"
	if dead {
		state = "dead"
	}
	updates := map[string]any{"delivery_state": state, "attempt_count": attempts, "next_attempt_at": retryAt, "updated_at": time.Now().UTC()}
	if dead {
		updates["encrypted_token_payload"] = nil
	}
	return repository.database.WithContext(ctx).Model(&accounts.EmailOutbox{}).Where("id = ?", id).Updates(updates).Error
}

func duplicateKey(err error, key string) bool {
	var mysqlError *mysqldriver.MySQLError
	return errors.As(err, &mysqlError) && mysqlError.Number == 1062 && strings.Contains(mysqlError.Message, key)
}
