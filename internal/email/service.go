package email

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	appconfig "github.com/chrisabs/cadence/internal/config"
)

type Service struct {
	client *ses.Client
	sender string
}

func NewService() (*Service, error) {
	cfg, err := appconfig.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to load app config: %v", err)
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.AWSRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %v", err)
	}

	client := ses.NewFromConfig(awsCfg)
	
	sender := cfg.SenderEmail
	if sender == "" {
		sender = "no-reply@chrisabbott.dev"
	}
	
	return &Service{
		client: client,
		sender: sender,
	}, nil
}

func (s *Service) SendInviteEmail(recipientEmail, inviteToken string, familyName string) error {
	subject := fmt.Sprintf("You've been invited to join %s on Cadence", familyName)
	
	htmlBody := fmt.Sprintf(`
		<html>
		<head>
			<title>Family Invitation</title>
		</head>
		<body>
			<h1>You've been invited to join %s on Cadence</h1>
			<p>You've been invited to join a family on Cadence, the family organisation platform.</p>
			<p>To accept this invitation, please click the link below:</p>
			<p><a href="https://app.chrisabbott.dev/invite?token=%s">Accept Invitation</a></p>
			<p>If you don't have an account yet, you'll be able to create one.</p>
			<p>This invitation will expire in 7 days.</p>
		</body>
		</html>
	`, familyName, inviteToken)
	
	textBody := fmt.Sprintf(`
		You've been invited to join %s on Cadence

		You've been invited to join a family on Cadence, the family organisation platform.
		
		To accept this invitation, copy and paste this link into your browser:
		https://app.chrisabbott.dev/invite?token=%s
		
		If you don't have an account yet, you'll be able to create one.
		
		This invitation will expire in 7 days.
	`, familyName, inviteToken)

	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{recipientEmail},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(htmlBody),
				},
				Text: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(textBody),
				},
			},
			Subject: &types.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(s.sender),
	}

	_, err := s.client.SendEmail(context.Background(), input)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}