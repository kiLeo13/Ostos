package accounts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/url"
	"strings"
	"time"

	"ostos/backend/internal/platform"
)

type EmailMessage struct {
	Recipient string
	Subject   string
	TextBody  string
	HTMLBody  string
}

type EmailSender interface {
	Send(context.Context, EmailMessage) error
}

type OutboxRepository interface {
	ClaimEmail(context.Context, time.Time, time.Time) (*EmailOutbox, error)
	MarkEmailSent(context.Context, platform.ID, time.Time) error
	MarkEmailFailed(context.Context, platform.ID, uint16, time.Time, bool) error
}

type OutboxWorker struct {
	repository   OutboxRepository
	cipher       *TokenCipher
	sender       EmailSender
	clock        platform.Clock
	publicAppURL string
}

func NewOutboxWorker(repository OutboxRepository, cipher *TokenCipher, sender EmailSender, clock platform.Clock, publicAppURL string) (*OutboxWorker, error) {
	if repository == nil || cipher == nil || sender == nil || clock == nil || publicAppURL == "" {
		return nil, errors.New("all outbox worker dependencies are required")
	}
	return &OutboxWorker{repository: repository, cipher: cipher, sender: sender, clock: clock, publicAppURL: strings.TrimRight(publicAppURL, "/")}, nil
}

func (worker *OutboxWorker) DispatchOne(ctx context.Context) error {
	now := worker.clock.Now().UTC()
	email, err := worker.repository.ClaimEmail(ctx, now, now.Add(2*time.Minute))
	if err != nil {
		return err
	}
	token, err := worker.cipher.Decrypt(email.EncryptedTokenPayload)
	var data map[string]string
	if err == nil {
		err = json.Unmarshal(email.TemplateData, &data)
	}
	if err == nil {
		err = worker.sender.Send(ctx, worker.message(*email, token, data["display_name"]))
	}
	if err == nil {
		return worker.repository.MarkEmailSent(ctx, email.ID, worker.clock.Now().UTC())
	}
	attempts := email.AttemptCount + 1
	dead := attempts >= 5
	backoff := time.Minute * time.Duration(1<<min(int(attempts), 6))
	if markErr := worker.repository.MarkEmailFailed(ctx, email.ID, attempts, now.Add(backoff), dead); markErr != nil {
		return fmt.Errorf("send email: %v; mark failed: %w", err, markErr)
	}
	return fmt.Errorf("send account email: %w", err)
}

func (worker *OutboxWorker) message(email EmailOutbox, token, displayName string) EmailMessage {
	path, subject, action := "/verify-email", "Verify your Ostos email", "Verify email"
	if email.Template == EmailTemplateResetPassword {
		path, subject, action = "/reset-password", "Reset your Ostos password", "Reset password"
	}
	link := worker.publicAppURL + path + "?token=" + url.QueryEscape(token)
	safeName, safeLink := html.EscapeString(displayName), html.EscapeString(link)
	return EmailMessage{
		Recipient: email.Recipient,
		Subject:   subject,
		TextBody:  fmt.Sprintf("Hello %s,\n\n%s: %s\n", displayName, action, link),
		HTMLBody:  fmt.Sprintf("<p>Hello %s,</p><p><a href=\"%s\">%s</a></p>", safeName, safeLink, action),
	}
}
