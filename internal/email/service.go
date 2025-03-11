package email

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
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