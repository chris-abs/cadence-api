package cloud

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	appconfig "github.com/chrisabs/storage/internal/config"
)

type S3Handler struct {
    client *s3.Client
    bucket string
    region string
}

func NewS3Handler() (*S3Handler, error) {
    cfg, err := appconfig.LoadConfig()
    if err != nil {
        return nil, fmt.Errorf("unable to load app config: %v", err)
    }

    awsCfg, err := config.LoadDefaultConfig(context.Background(),
        config.WithRegion(cfg.AWSRegion),
    )
    if err != nil {
        return nil, fmt.Errorf("unable to load SDK config: %v", err)
    }

    client := s3.NewFromConfig(awsCfg)
    return &S3Handler{
        client: client,
        bucket: cfg.S3Bucket,
        region: cfg.AWSRegion,
    }, nil
}

func (h *S3Handler) UploadFile(file *multipart.FileHeader, prefix string) (string, error) {
    src, err := file.Open()
    if err != nil {
        return "", fmt.Errorf("error opening file: %v", err)
    }
    defer src.Close()

    filename := generateFilename(prefix, file.Filename)

    contentType := file.Header.Get("Content-Type")
    _, err = h.client.PutObject(context.Background(), &s3.PutObjectInput{
        Bucket:      &h.bucket,
        Key:         &filename,
        Body:        src,
        ContentType: &contentType,
    })

    if err != nil {
        return "", fmt.Errorf("error uploading to S3: %v", err)
    }

    return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", h.bucket, h.region, filename), nil
}

func generateFilename(prefix, originalName string) string {
    ext := filepath.Ext(originalName)
    timestamp := time.Now().UnixNano()
    return fmt.Sprintf("%s/%d%s", prefix, timestamp, ext)
}