package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
    JWTSecret          string
    DatabaseURL        string
    AWSAccessKeyID     string
    AWSSecretAccessKey string
    AWSRegion          string
    S3Bucket           string
    SenderEmail        string
    AppBaseURL         string
}

func LoadConfig() (*Config, error) {
    err := godotenv.Load()
    if err != nil && !os.IsNotExist(err) {
        return nil, fmt.Errorf("error loading .env file: %v", err)
    }

    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        return nil, fmt.Errorf("JWT_SECRET environment variable is required")
    }

    awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
    if awsAccessKey == "" {
        return nil, fmt.Errorf("AWS_ACCESS_KEY_ID environment variable is required")
    }

    awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
    if awsSecretKey == "" {
        return nil, fmt.Errorf("AWS_SECRET_ACCESS_KEY environment variable is required")
    }

    awsRegion := os.Getenv("AWS_REGION")
    if awsRegion == "" {
        return nil, fmt.Errorf("AWS_REGION environment variable is required")
    }

    s3Bucket := os.Getenv("S3_BUCKET")
    if s3Bucket == "" {
        return nil, fmt.Errorf("S3_BUCKET environment variable is required")
    }

    senderEmail := os.Getenv("SES_SENDER_EMAIL")

    appBaseURL := os.Getenv("APP_BASE_URL")
    if appBaseURL == "" {
        appBaseURL = "http://localhost:3000"
    }


    return &Config{
        JWTSecret:          jwtSecret,
        AWSAccessKeyID:     awsAccessKey,
        AWSSecretAccessKey: awsSecretKey,
        AWSRegion:          awsRegion,
        S3Bucket:           s3Bucket,
        SenderEmail:        senderEmail,
    }, nil
}