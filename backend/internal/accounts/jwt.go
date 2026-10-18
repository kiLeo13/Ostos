package accounts

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"ostos/backend/internal/platform"
)

type AccessIdentity struct {
	UserID    platform.ID
	SessionID platform.ID
}

type accessClaims struct {
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
}

func NewJWTManager(secret []byte, issuer, audience string, ttl time.Duration) (*JWTManager, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT secret must contain at least 32 bytes")
	}
	if issuer == "" || audience == "" || ttl <= 0 {
		return nil, errors.New("JWT issuer, audience, and positive lifetime are required")
	}
	return &JWTManager{secret: append([]byte(nil), secret...), issuer: issuer, audience: audience, ttl: ttl}, nil
}

func (manager *JWTManager) Issue(userID, sessionID, tokenID platform.ID, now time.Time) (string, time.Time, error) {
	expiresAt := now.UTC().Add(manager.ttl)
	claims := accessClaims{
		SessionID: sessionID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: manager.issuer, Subject: userID.String(), Audience: jwt.ClaimStrings{manager.audience},
			ExpiresAt: jwt.NewNumericDate(expiresAt), IssuedAt: jwt.NewNumericDate(now.UTC()),
			NotBefore: jwt.NewNumericDate(now.UTC()), ID: tokenID.String(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(manager.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return token, expiresAt, nil
}

func (manager *JWTManager) Parse(raw string) (AccessIdentity, error) {
	claims := &accessClaims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		return manager.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(manager.issuer),
		jwt.WithAudience(manager.audience), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil {
		return AccessIdentity{}, ErrInvalidCredentials
	}
	userID, err := parsePositiveID(claims.Subject)
	if err != nil {
		return AccessIdentity{}, ErrInvalidCredentials
	}
	sessionID, err := parsePositiveID(claims.SessionID)
	if err != nil || claims.ID == "" {
		return AccessIdentity{}, ErrInvalidCredentials
	}
	return AccessIdentity{UserID: userID, SessionID: sessionID}, nil
}

func parsePositiveID(value string) (platform.ID, error) {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, errors.New("invalid ID")
	}
	return platform.ID(parsed), nil
}
