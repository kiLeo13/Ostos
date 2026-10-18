package accounts

import (
	"bytes"
	"testing"
)

func TestArgon2HasherRoundTripAndRejectsInvalidInputs(t *testing.T) {
	hasher := Argon2Hasher{}
	encoded, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !hasher.Compare(encoded, "correct horse battery staple") {
		t.Fatal("correct password was rejected")
	}
	if hasher.Compare(encoded, "wrong") {
		t.Fatal("wrong password was accepted")
	}
	for _, malformed := range []string{"", "$argon2id$v=19$m=999999,t=2,p=1$c2FsdA$aGFzaA", "$argon2i$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA"} {
		if hasher.Compare(malformed, "password") {
			t.Fatalf("malformed hash was accepted: %q", malformed)
		}
	}
}

func TestBearerTokenAndCipher(t *testing.T) {
	raw, hash, err := NewBearerToken()
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	if raw == "" || len(hash) != 32 || !bytes.Equal(hash, HashBearerToken(raw)) {
		t.Fatal("token or hash has unexpected shape")
	}
	cipher, err := NewTokenCipher(bytes.Repeat([]byte{7}, 32))
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	encrypted, err := cipher.Encrypt(raw)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	decrypted, err := cipher.Decrypt(encrypted)
	if err != nil || decrypted != raw {
		t.Fatalf("decrypt = %q, %v", decrypted, err)
	}
	encrypted[len(encrypted)-1] ^= 1
	if _, err := cipher.Decrypt(encrypted); err == nil {
		t.Fatal("tampered ciphertext was accepted")
	}
}

func TestTokenCipherRejectsWrongKeyLength(t *testing.T) {
	if _, err := NewTokenCipher(make([]byte, 31)); err == nil {
		t.Fatal("31-byte key was accepted")
	}
}
