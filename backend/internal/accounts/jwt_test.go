package accounts

import (
	"bytes"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"ostos/backend/internal/platform"
)

func TestJWTManagerIssuesAndValidatesRestrictedClaims(t *testing.T) {
	manager, err := NewJWTManager(bytes.Repeat([]byte{1}, 32), "ostos-test", "ostos-clients", 30*time.Minute)
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	token, _, err := manager.Issue(101, 202, 303, time.Now().UTC().Add(-time.Second))
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	identity, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if identity != (AccessIdentity{UserID: platform.ID(101), SessionID: platform.ID(202)}) {
		t.Fatalf("identity = %+v", identity)
	}
}

func TestJWTManagerRejectsWrongSignatureAudienceExpiryFutureAndAlgorithm(t *testing.T) {
	now := time.Now().UTC()
	valid, _ := NewJWTManager(bytes.Repeat([]byte{1}, 32), "issuer", "audience", 30*time.Minute)
	wrongSecret, _ := NewJWTManager(bytes.Repeat([]byte{2}, 32), "issuer", "audience", 30*time.Minute)
	wrongAudience, _ := NewJWTManager(bytes.Repeat([]byte{1}, 32), "issuer", "other", 30*time.Minute)
	token, _, _ := valid.Issue(1, 2, 3, now)
	if _, err := wrongSecret.Parse(token); err == nil {
		t.Fatal("wrong signature was accepted")
	}
	if _, err := wrongAudience.Parse(token); err == nil {
		t.Fatal("wrong audience was accepted")
	}
	expired, _ := NewJWTManager(bytes.Repeat([]byte{1}, 32), "issuer", "audience", time.Minute)
	expiredToken, _, _ := expired.Issue(1, 2, 3, now.Add(-2*time.Minute))
	if _, err := expired.Parse(expiredToken); err == nil {
		t.Fatal("expired token was accepted")
	}
	futureToken, _, _ := valid.Issue(1, 2, 3, now.Add(time.Minute))
	if _, err := valid.Parse(futureToken); err == nil {
		t.Fatal("future token was accepted")
	}
	claims := jwt.MapClaims{"sub": "1", "sid": "2", "jti": "3", "iss": "issuer", "aud": "audience", "exp": now.Add(time.Minute).Unix()}
	none, _ := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if _, err := valid.Parse(none); err == nil {
		t.Fatal("none algorithm was accepted")
	}
}
