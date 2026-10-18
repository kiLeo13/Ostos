package accounts

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"ostos/backend/internal/platform"
)

type outboxClock struct{ now time.Time }

func (clock outboxClock) Now() time.Time { return clock.now }

type outboxRepository struct {
	email  *EmailOutbox
	sent   bool
	failed bool
	dead   bool
}

func (repository *outboxRepository) ClaimEmail(context.Context, time.Time, time.Time) (*EmailOutbox, error) {
	if repository.email == nil {
		return nil, ErrNoPendingEmail
	}
	return repository.email, nil
}

func (repository *outboxRepository) MarkEmailSent(context.Context, platform.ID, time.Time) error {
	repository.sent = true
	return nil
}

func (repository *outboxRepository) MarkEmailFailed(_ context.Context, _ platform.ID, _ uint16, _ time.Time, dead bool) error {
	repository.failed, repository.dead = true, dead
	return nil
}

type emailSender struct {
	message EmailMessage
	err     error
}

func (sender *emailSender) Send(_ context.Context, message EmailMessage) error {
	sender.message = message
	return sender.err
}

func TestOutboxWorkerDecryptsAndDeliversWithoutExposingTokenOutsideMessage(t *testing.T) {
	cipher, _ := NewTokenCipher(bytes.Repeat([]byte{8}, 32))
	encrypted, _ := cipher.Encrypt("secret-token")
	repository := &outboxRepository{email: &EmailOutbox{
		ID: 1, Template: EmailTemplateVerifyEmail, Recipient: "user@example.com",
		TemplateData: []byte("{\"display_name\":\"A <B>\"}"), EncryptedTokenPayload: encrypted,
	}}
	sender := &emailSender{}
	worker, err := NewOutboxWorker(repository, cipher, sender, outboxClock{now: time.Now()}, "https://app.example.test")
	if err != nil {
		t.Fatalf("worker: %v", err)
	}
	if err := worker.DispatchOne(context.Background()); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if !repository.sent || repository.failed {
		t.Fatalf("delivery state: %+v", repository)
	}
	if !strings.Contains(sender.message.TextBody, "secret-token") || strings.Contains(sender.message.HTMLBody, "A <B>") {
		t.Fatalf("message was not built or escaped correctly: %+v", sender.message)
	}
}

func TestOutboxWorkerRetriesAndEventuallyMarksDead(t *testing.T) {
	cipher, _ := NewTokenCipher(bytes.Repeat([]byte{8}, 32))
	encrypted, _ := cipher.Encrypt("token")
	repository := &outboxRepository{email: &EmailOutbox{ID: 1, AttemptCount: 4, TemplateData: []byte("{}"), EncryptedTokenPayload: encrypted}}
	worker, _ := NewOutboxWorker(repository, cipher, &emailSender{err: errors.New("SES unavailable")}, outboxClock{now: time.Now()}, "https://app.example.test")
	if err := worker.DispatchOne(context.Background()); err == nil {
		t.Fatal("delivery failure was hidden")
	}
	if !repository.failed || !repository.dead || repository.sent {
		t.Fatalf("delivery state: %+v", repository)
	}
}
