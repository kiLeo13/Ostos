package accounts

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonMemory      = 19 * 1024
	argonIterations  = 2
	argonParallelism = 1
	argonSaltLength  = 16
	argonKeyLength   = 32
	tokenBytes       = 32
)

type PasswordHasher interface {
	Hash(string) (string, error)
	Compare(string, string) bool
}

type Argon2Hasher struct{}

func (Argon2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, argonSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallelism, argonKeyLength)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, argonMemory,
		argonIterations, argonParallelism, base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash)), nil
}

func (Argon2Hasher) Compare(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var version int
	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false
	}
	if memory > 256*1024 || iterations > 10 || parallelism > 16 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 8 {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expected) < 16 {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expected)))
	return subtle.ConstantTimeCompare(expected, actual) == 1
}

func NewBearerToken() (string, []byte, error) {
	raw := make([]byte, tokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, fmt.Errorf("generate bearer token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, HashBearerToken(token), nil
}

func HashBearerToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}

type TokenCipher struct{ aead cipher.AEAD }

func NewTokenCipher(key []byte) (*TokenCipher, error) {
	if len(key) != 32 {
		return nil, errors.New("email token encryption key must contain exactly 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create email token cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create email token AEAD: %w", err)
	}
	return &TokenCipher{aead: aead}, nil
}

func (cipher *TokenCipher) Encrypt(plaintext string) ([]byte, error) {
	nonce := make([]byte, cipher.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generate token nonce: %w", err)
	}
	result := make([]byte, 2, 2+len(nonce)+len(plaintext)+cipher.aead.Overhead())
	binary.BigEndian.PutUint16(result, uint16(len(nonce)))
	result = append(result, nonce...)
	return cipher.aead.Seal(result, nonce, []byte(plaintext), nil), nil
}

func (cipher *TokenCipher) Decrypt(payload []byte) (string, error) {
	if len(payload) < 2 {
		return "", errors.New("encrypted token payload is truncated")
	}
	nonceLength := int(binary.BigEndian.Uint16(payload[:2]))
	if nonceLength != cipher.aead.NonceSize() || len(payload) < 2+nonceLength+cipher.aead.Overhead() {
		return "", errors.New("encrypted token payload is malformed")
	}
	nonce := payload[2 : 2+nonceLength]
	plaintext, err := cipher.aead.Open(nil, nonce, payload[2+nonceLength:], nil)
	if err != nil {
		return "", errors.New("decrypt email token payload")
	}
	return string(plaintext), nil
}
