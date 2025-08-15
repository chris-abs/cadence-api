package cloud

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	appconfig "github.com/chrisabs/cadence/internal/config"
	"github.com/chrisabs/cadence/internal/models"
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

func (h *S3Handler) UploadFile(file *multipart.FileHeader, familyID models.FamilyID, prefix string) (string, error) {
    src, err := file.Open()
    if err != nil {
        return "", fmt.Errorf("error opening file: %v", err)
    }
    defer src.Close()

    filename := generateFilename(familyID, prefix, file.Filename)

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

// For family-shared data (storage items, etc.)
func (h *S3Handler) UploadFamilySharedFile(file *multipart.FileHeader, familyID models.FamilyID, category string) (string, error) {
    prefix := fmt.Sprintf("shared/%s", category)
    return h.UploadFile(file, familyID, prefix)
}

// For profile-specific media (private to profile)
func (h *S3Handler) UploadProfileMediaFile(file *multipart.FileHeader, familyID models.FamilyID, profileID models.ProfileID, mediaType string) (string, error) {
    prefix := fmt.Sprintf("profiles/%s/media/%s", profileID, mediaType)
    return h.UploadFile(file, familyID, prefix)
}

// For profile avatars (family-visible)
func (h *S3Handler) UploadProfileAvatar(file *multipart.FileHeader, familyID models.FamilyID, profileID models.ProfileID) (string, error) {
    prefix := fmt.Sprintf("profiles/%s/avatar", profileID)
    return h.UploadFile(file, familyID, prefix)
}

func generateFilename(familyID models.FamilyID, prefix, originalName string) string {
    ext := filepath.Ext(originalName)
    timestamp := time.Now().UnixNano()
    return fmt.Sprintf("families/%s/%s/%d%s", familyID, prefix, timestamp, ext)
}