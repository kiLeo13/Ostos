package ses

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsses "github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"ostos/backend/internal/accounts"
)

type Client interface {
	SendEmail(context.Context, *awsses.SendEmailInput, ...func(*awsses.Options)) (*awsses.SendEmailOutput, error)
}

type Sender struct {
	client Client
	from   string
}

func New(client Client, from string) (*Sender, error) {
	if client == nil || from == "" {
		return nil, errors.New("SES client and sender address are required")
	}
	return &Sender{client: client, from: from}, nil
}

func (sender *Sender) Send(ctx context.Context, message accounts.EmailMessage) error {
	_, err := sender.client.SendEmail(ctx, &awsses.SendEmailInput{
		FromEmailAddress: aws.String(sender.from),
		Destination:      &types.Destination{ToAddresses: []string{message.Recipient}},
		Content: &types.EmailContent{Simple: &types.Message{
			Subject: &types.Content{Data: aws.String(message.Subject), Charset: aws.String("UTF-8")},
			Body: &types.Body{
				Text: &types.Content{Data: aws.String(message.TextBody), Charset: aws.String("UTF-8")},
				Html: &types.Content{Data: aws.String(message.HTMLBody), Charset: aws.String("UTF-8")},
			},
		}},
	})
	if err != nil {
		return fmt.Errorf("SES SendEmail: %w", err)
	}
	return nil
}
