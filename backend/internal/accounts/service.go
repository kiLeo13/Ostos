package accounts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"ostos/backend/internal/platform"
)

const (
	verificationTokenTTL = 24 * time.Hour
	passwordResetTTL     = time.Hour
	refreshTokenTTL      = 30 * 24 * time.Hour
)

type Service struct {
	repository Repository
	ids        platform.IDGenerator
	clock      platform.Clock
	passwords  PasswordHasher
	jwt        *JWTManager
	cipher     *TokenCipher
}

func NewService(repository Repository, ids platform.IDGenerator, clock platform.Clock, passwords PasswordHasher, jwt *JWTManager, cipher *TokenCipher) (*Service, error) {
	if repository == nil || ids == nil || clock == nil || passwords == nil || jwt == nil || cipher == nil {
		return nil, errors.New("all account service dependencies are required")
	}
	return &Service{repository: repository, ids: ids, clock: clock, passwords: passwords, jwt: jwt, cipher: cipher}, nil
}

type RegisterInput struct {
	Email       string
	Username    string
	DisplayName string
	Password    string
}

func (service *Service) Register(ctx context.Context, input RegisterInput) (PublicUser, error) {
	now := service.clock.Now().UTC()
	userID, err := service.ids.Next(ctx)
	if err != nil {
		return PublicUser{}, err
	}
	credentialID, err := service.ids.Next(ctx)
	if err != nil {
		return PublicUser{}, err
	}
	hash, err := service.passwords.Hash(input.Password)
	if err != nil {
		return PublicUser{}, err
	}
	user := &User{ID: userID, Email: CanonicalEmail(input.Email), Username: input.Username,
		DisplayName: strings.TrimSpace(input.DisplayName), CreatedAt: now, UpdatedAt: now}
	credential := &Credential{ID: credentialID, UserID: userID, PasswordHash: hash,
		PasswordChangedAt: now, CreatedAt: now, UpdatedAt: now}
	err = service.repository.WithinTransaction(ctx, func(repository Repository) error {
		if err := repository.CreateUser(ctx, user); err != nil {
			return err
		}
		if err := repository.CreateCredential(ctx, credential); err != nil {
			return err
		}
		if err := service.createAccountEmail(ctx, repository, user, TokenPurposeVerifyEmail, EmailTemplateVerifyEmail, verificationTokenTTL, now); err != nil {
			return err
		}
		return service.audit(ctx, repository, user.ID, "account.created", "user", user.ID, now)
	})
	if err != nil {
		return PublicUser{}, err
	}
	return user.Public(), nil
}

func (service *Service) GetProfile(ctx context.Context, userID platform.ID) (PublicUser, error) {
	user, err := service.repository.FindUserByID(ctx, userID)
	if err != nil {
		return PublicUser{}, err
	}
	return user.Public(), nil
}

type UpdateProfileInput struct {
	Username    *string
	DisplayName *string
}

func (service *Service) UpdateProfile(ctx context.Context, userID platform.ID, input UpdateProfileInput) (PublicUser, error) {
	user, err := service.repository.FindUserByID(ctx, userID)
	if err != nil {
		return PublicUser{}, err
	}
	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.DisplayName != nil {
		user.DisplayName = strings.TrimSpace(*input.DisplayName)
	}
	now := service.clock.Now().UTC()
	if err := service.repository.UpdateUserProfile(ctx, userID, user.Username, user.DisplayName, now); err != nil {
		return PublicUser{}, err
	}
	user.UpdatedAt = now
	return user.Public(), nil
}

func (service *Service) ChangePassword(ctx context.Context, userID platform.ID, currentPassword, newPassword string) error {
	credential, err := service.repository.FindCredentialByUserID(ctx, userID)
	if err != nil || !service.passwords.Compare(credential.PasswordHash, currentPassword) {
		return ErrInvalidCredentials
	}
	hash, err := service.passwords.Hash(newPassword)
	if err != nil {
		return err
	}
	now := service.clock.Now().UTC()
	return service.repository.WithinTransaction(ctx, func(repository Repository) error {
		if err := repository.UpdatePassword(ctx, userID, hash, now); err != nil {
			return err
		}
		if err := repository.RevokeAllSessions(ctx, userID, now); err != nil {
			return err
		}
		return service.audit(ctx, repository, userID, "account.password_changed", "user", userID, now)
	})
}

func (service *Service) DeleteAccount(ctx context.Context, userID platform.ID, currentPassword string) error {
	credential, err := service.repository.FindCredentialByUserID(ctx, userID)
	if err != nil || !service.passwords.Compare(credential.PasswordHash, currentPassword) {
		return ErrInvalidCredentials
	}
	now := service.clock.Now().UTC()
	return service.repository.WithinTransaction(ctx, func(repository Repository) error {
		if err := repository.SoftDeleteUserIfNoOwnedFamilies(ctx, userID, now); err != nil {
			return err
		}
		if err := repository.RevokeAllSessions(ctx, userID, now); err != nil {
			return err
		}
		return service.audit(ctx, repository, userID, "account.deleted", "user", userID, now)
	})
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (service *Service) Login(ctx context.Context, identifier, password string, deviceName *string) (TokenPair, error) {
	identifier = strings.TrimSpace(identifier)
	if strings.Contains(identifier, "@") {
		identifier = CanonicalEmail(identifier)
	}
	user, credential, err := service.repository.FindUserAndCredentialByIdentifier(ctx, identifier)
	if err != nil || !service.passwords.Compare(credential.PasswordHash, password) {
		return TokenPair{}, ErrInvalidCredentials
	}
	if user.EmailVerifiedAt == nil {
		return TokenPair{}, ErrEmailNotVerified
	}
	var pair TokenPair
	now := service.clock.Now().UTC()
	err = service.repository.WithinTransaction(ctx, func(repository Repository) error {
		var createErr error
		pair, createErr = service.newSession(ctx, repository, user.ID, deviceName, now)
		if createErr != nil {
			return createErr
		}
		return service.audit(ctx, repository, user.ID, "session.created", "user", user.ID, now)
	})
	return pair, err
}

func (service *Service) Refresh(ctx context.Context, refreshToken string) (pair TokenPair, err error) {
	hash := HashBearerToken(refreshToken)
	now := service.clock.Now().UTC()
	var reused bool
	err = service.repository.WithinTransaction(ctx, func(repository Repository) error {
		session, findErr := repository.FindSessionByHashForUpdate(ctx, hash)
		if findErr != nil {
			return ErrInvalidRefreshToken
		}
		if session.ReplacedBySessionID != nil {
			if revokeErr := repository.RevokeSessionChain(ctx, session.UserID, session.ID, now); revokeErr != nil {
				return revokeErr
			}
			reused = true
			return service.audit(ctx, repository, session.UserID, "session.refresh_reuse_detected", "session", session.ID, now)
		}
		if session.RevokedAt != nil || !session.ExpiresAt.After(now) {
			return ErrInvalidRefreshToken
		}
		newID, idErr := service.ids.Next(ctx)
		if idErr != nil {
			return idErr
		}
		raw, tokenHash, tokenErr := NewBearerToken()
		if tokenErr != nil {
			return tokenErr
		}
		newSession := &Session{ID: newID, UserID: session.UserID, RefreshTokenHash: tokenHash,
			DeviceName: session.DeviceName, CreatedAt: now, LastUsedAt: now, ExpiresAt: now.Add(refreshTokenTTL)}
		if rotateErr := repository.RotateSession(ctx, session.ID, newSession, now); rotateErr != nil {
			return rotateErr
		}
		access, expiresAt, issueErr := service.issueAccess(ctx, session.UserID, newID, now)
		if issueErr != nil {
			return issueErr
		}
		pair = TokenPair{AccessToken: access, RefreshToken: raw, TokenType: "Bearer", ExpiresIn: int64(expiresAt.Sub(now).Seconds())}
		return service.audit(ctx, repository, session.UserID, "session.refreshed", "session", newID, now)
	})
	if err == nil && reused {
		return TokenPair{}, ErrRefreshTokenReuse
	}
	return pair, err
}

func (service *Service) Logout(ctx context.Context, identity AccessIdentity) error {
	now := service.clock.Now().UTC()
	return service.repository.WithinTransaction(ctx, func(repository Repository) error {
		if err := repository.RevokeSession(ctx, identity.UserID, identity.SessionID, now); err != nil {
			return err
		}
		return service.audit(ctx, repository, identity.UserID, "session.revoked", "session", identity.SessionID, now)
	})
}

func (service *Service) LogoutAll(ctx context.Context, userID platform.ID) error {
	now := service.clock.Now().UTC()
	return service.repository.WithinTransaction(ctx, func(repository Repository) error {
		if err := repository.RevokeAllSessions(ctx, userID, now); err != nil {
			return err
		}
		return service.audit(ctx, repository, userID, "session.revoked_all", "user", userID, now)
	})
}

func (service *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	now := service.clock.Now().UTC()
	return service.repository.WithinTransaction(ctx, func(repository Repository) error {
		token, err := repository.FindAccountTokenForUpdate(ctx, HashBearerToken(rawToken), TokenPurposeVerifyEmail)
		if err != nil || token.UsedAt != nil || !token.ExpiresAt.After(now) {
			return ErrInvalidAccountToken
		}
		if err := repository.MarkEmailVerified(ctx, token.UserID, now); err != nil {
			return err
		}
		if err := repository.MarkAccountTokenUsed(ctx, token.ID, now); err != nil {
			return err
		}
		return service.audit(ctx, repository, token.UserID, "account.email_verified", "user", token.UserID, now)
	})
}

func (service *Service) ResendVerification(ctx context.Context, email string) error {
	user, err := service.repository.FindUserByEmail(ctx, CanonicalEmail(email))
	if errors.Is(err, ErrUserNotFound) || (err == nil && user.EmailVerifiedAt != nil) {
		return nil
	}
	if err != nil {
		return err
	}
	now := service.clock.Now().UTC()
	return service.repository.WithinTransaction(ctx, func(repository Repository) error {
		if err := repository.InvalidateAccountTokens(ctx, user.ID, TokenPurposeVerifyEmail, now); err != nil {
			return err
		}
		return service.createAccountEmail(ctx, repository, user, TokenPurposeVerifyEmail, EmailTemplateVerifyEmail, verificationTokenTTL, now)
	})
}

func (service *Service) ForgotPassword(ctx context.Context, email string) error {
	user, err := service.repository.FindUserByEmail(ctx, CanonicalEmail(email))
	if errors.Is(err, ErrUserNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	now := service.clock.Now().UTC()
	return service.repository.WithinTransaction(ctx, func(repository Repository) error {
		if err := repository.InvalidateAccountTokens(ctx, user.ID, TokenPurposeResetPassword, now); err != nil {
			return err
		}
		return service.createAccountEmail(ctx, repository, user, TokenPurposeResetPassword, EmailTemplateResetPassword, passwordResetTTL, now)
	})
}

func (service *Service) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	hash, err := service.passwords.Hash(newPassword)
	if err != nil {
		return err
	}
	now := service.clock.Now().UTC()
	return service.repository.WithinTransaction(ctx, func(repository Repository) error {
		token, findErr := repository.FindAccountTokenForUpdate(ctx, HashBearerToken(rawToken), TokenPurposeResetPassword)
		if findErr != nil || token.UsedAt != nil || !token.ExpiresAt.After(now) {
			return ErrInvalidAccountToken
		}
		if err := repository.UpdatePassword(ctx, token.UserID, hash, now); err != nil {
			return err
		}
		if err := repository.MarkAccountTokenUsed(ctx, token.ID, now); err != nil {
			return err
		}
		if err := repository.RevokeAllSessions(ctx, token.UserID, now); err != nil {
			return err
		}
		return service.audit(ctx, repository, token.UserID, "account.password_reset", "user", token.UserID, now)
	})
}

func (service *Service) ParseAccessToken(raw string) (AccessIdentity, error) {
	return service.jwt.Parse(raw)
}

func (service *Service) newSession(ctx context.Context, repository Repository, userID platform.ID, deviceName *string, now time.Time) (TokenPair, error) {
	sessionID, err := service.ids.Next(ctx)
	if err != nil {
		return TokenPair{}, err
	}
	raw, hash, err := NewBearerToken()
	if err != nil {
		return TokenPair{}, err
	}
	session := &Session{ID: sessionID, UserID: userID, RefreshTokenHash: hash, DeviceName: deviceName,
		CreatedAt: now, LastUsedAt: now, ExpiresAt: now.Add(refreshTokenTTL)}
	if err := repository.CreateSession(ctx, session); err != nil {
		return TokenPair{}, err
	}
	access, expiresAt, err := service.issueAccess(ctx, userID, sessionID, now)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: raw, TokenType: "Bearer", ExpiresIn: int64(expiresAt.Sub(now).Seconds())}, nil
}

func (service *Service) issueAccess(ctx context.Context, userID, sessionID platform.ID, now time.Time) (string, time.Time, error) {
	tokenID, err := service.ids.Next(ctx)
	if err != nil {
		return "", time.Time{}, err
	}
	return service.jwt.Issue(userID, sessionID, tokenID, now)
}

func (service *Service) createAccountEmail(ctx context.Context, repository Repository, user *User, purpose TokenPurpose, template EmailTemplate, ttl time.Duration, now time.Time) error {
	tokenID, err := service.ids.Next(ctx)
	if err != nil {
		return err
	}
	outboxID, err := service.ids.Next(ctx)
	if err != nil {
		return err
	}
	raw, hash, err := NewBearerToken()
	if err != nil {
		return err
	}
	encrypted, err := service.cipher.Encrypt(raw)
	if err != nil {
		return err
	}
	templateData, err := json.Marshal(map[string]string{"display_name": user.DisplayName})
	if err != nil {
		return fmt.Errorf("encode email template data: %w", err)
	}
	token := &AccountToken{ID: tokenID, UserID: user.ID, Purpose: purpose, BearerTokenHash: hash,
		ExpiresAt: now.Add(ttl), CreatedAt: now}
	outbox := &EmailOutbox{ID: outboxID, UserID: user.ID, Template: template, Recipient: user.Email,
		TemplateData: templateData, EncryptedTokenPayload: encrypted, DeliveryState: "pending",
		NextAttemptAt: now, CreatedAt: now, UpdatedAt: now}
	if err := repository.CreateAccountToken(ctx, token); err != nil {
		return err
	}
	return repository.CreateEmailOutbox(ctx, outbox)
}

func (service *Service) audit(ctx context.Context, repository Repository, actorID platform.ID, action, targetType string, targetID platform.ID, now time.Time) error {
	id, err := service.ids.Next(ctx)
	if err != nil {
		return err
	}
	var target *platform.ID
	if targetID > 0 {
		target = &targetID
	}
	event := &AuditEvent{ID: id, ActorUserID: &actorID, ActionCode: action, TargetType: targetType,
		TargetID: target, Metadata: []byte("{}"), RequestID: platform.RequestID(ctx), CreatedAt: now}
	return repository.CreateAuditEvent(ctx, event)
}
