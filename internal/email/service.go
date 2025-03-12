// internal/email/service.go
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
	client      *ses.Client
	sender      string
	appBaseURL  string
	environment string
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
	
	environment := "development"
	if cfg.AppBaseURL != "http://localhost:3000" && cfg.AppBaseURL != "" {
		if cfg.AppBaseURL == "https://cadence.chrisabbott.dev" {
			environment = "production"
		} else {
			environment = "staging"
		}
	}
	
	return &Service{
		client:      client,
		sender:      sender,
		appBaseURL:  cfg.AppBaseURL,
		environment: environment,
	}, nil
}

func (s *Service) SendInviteEmail(recipientEmail, inviteToken string, familyName string) error {
	subject := fmt.Sprintf("You've been invited to join %s on Cadence", familyName)
	
	inviteURL := fmt.Sprintf("%s/invite?token=%s", s.appBaseURL, inviteToken)
	
	if s.environment != "production" {
		subject = fmt.Sprintf("[%s] %s", s.environment, subject)
	}
	
	htmlBody := fmt.Sprintf(`
		<html>
		<head>
			<title>Family Invitation</title>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #4a86e8; color: white; padding: 10px 20px; }
				.content { padding: 20px; border: 1px solid #ddd; }
				.button { display: inline-block; background-color: #4a86e8; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px; }
				.footer { margin-top: 20px; font-size: 12px; color: #777; }
				.token-box { background-color: #f5f5f5; padding: 10px; border: 1px solid #ddd; font-family: monospace; margin: 15px 0; }
				.env-notice { background-color: #fff3cd; padding: 10px; border: 1px solid #ffeeba; margin: 15px 0; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Cadence Family Invitation</h1>
				</div>
				<div class="content">
					<h2>You've been invited to join %s</h2>
					<p>You've been invited to join a family on Cadence, the family organisation platform.</p>
					
					%s
					
					<p>To accept this invitation, please click the button below:</p>
					<p><a href="%s" class="button">Accept Invitation</a></p>
					
					<p>If the button doesn't work, you can copy and paste this URL into your browser:</p>
					<p><a href="%s">%s</a></p>
					
					<div class="token-box">
						<p>Your invitation token: <strong>%s</strong></p>
					</div>
					
					<p>If you don't have an account yet, you'll be able to create one.</p>
					<p>This invitation will expire in 7 days.</p>
				</div>
				<div class="footer">
					<p>This email was sent from Cadence - a family organisation platform.</p>
					<p>If you received this by mistake, please ignore this email.</p>
				</div>
			</div>
		</body>
		</html>
	`, 
	familyName, 
	s.getEnvironmentNotice(),
	inviteURL,
	inviteURL, inviteURL,
	inviteToken)
	
	textBody := fmt.Sprintf(`
		You've been invited to join %s on Cadence

		You've been invited to join a family on Cadence, the family organisation platform.
		
		%s
		
		To accept this invitation, copy and paste this link into your browser:
		%s
		
		Your invitation token: %s
		
		If you don't have an account yet, you'll be able to create one.
		
		This invitation will expire in 7 days.
	`, 
	familyName,
	s.getEnvironmentNoticeText(),
	inviteURL,
	inviteToken)

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

func (s *Service) getEnvironmentNotice() string {
	if s.environment == "production" {
		return ""
	}
	
	return fmt.Sprintf(`
		<div class="env-notice">
			<strong>%s Environment</strong>
			<p>This is a %s environment email. In case the links don't work, you can use the invitation token shown below.</p>
		</div>
	`, s.environment, s.environment)
}

func (s *Service) getEnvironmentNoticeText() string {
	if s.environment == "production" {
		return ""
	}
	
	return fmt.Sprintf("[%s ENVIRONMENT] This is a %s environment email.", 
		s.environment, s.environment)
}