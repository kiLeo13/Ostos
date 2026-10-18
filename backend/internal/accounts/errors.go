package accounts

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailTaken          = errors.New("email is already registered")
	ErrUsernameTaken       = errors.New("username is already registered")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrEmailNotVerified    = errors.New("email is not verified")
	ErrInvalidAccountToken = errors.New("account token is invalid")
	ErrInvalidRefreshToken = errors.New("refresh token is invalid")
	ErrRefreshTokenReuse   = errors.New("refresh token reuse detected")
	ErrOwnsFamilies        = errors.New("user still owns families")
	ErrNoPendingEmail      = errors.New("no pending email")
)
