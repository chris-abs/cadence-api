package cloud

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
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

// (shared across families)
func (h *S3Handler) UploadFamilySharedFile(file *multipart.FileHeader, familyID models.FamilyID, category string) (string, error) {
    prefix := fmt.Sprintf("shared/%s", category)
    return h.UploadFile(file, familyID, prefix)
}

// (private to profile)
func (h *S3Handler) UploadProfileMediaFile(file *multipart.FileHeader, familyID models.FamilyID, profileID models.ProfileID, mediaType string) (string, error) {
    prefix := fmt.Sprintf("profiles/%s/media/%s", profileID, mediaType)
    return h.UploadFile(file, familyID, prefix)
}

// (family-visible)
func (h *S3Handler) UploadProfileAvatar(file *multipart.FileHeader, familyID models.FamilyID, profileID models.ProfileID) (string, error) {
    prefix := fmt.Sprintf("profiles/%s/avatar", profileID)
    return h.UploadFile(file, familyID, prefix)
}

func (h *S3Handler) DeleteFile(familyID models.FamilyID, prefix, filename string) error {
    key := fmt.Sprintf("families/%s/%s/%s", familyID, prefix, filename)
    
    _, err := h.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
        Bucket: &h.bucket,
        Key:    &key,
    })
    
    if err != nil {
        return fmt.Errorf("error deleting file from S3: %v", err)
    }
    
    return nil
}

func (h *S3Handler) DeleteFileByURL(url string) error {
    // URL format: https://bucket.s3.region.amazonaws.com/families/familyID/prefix/timestamp.ext
    if !strings.Contains(url, ".amazonaws.com/") {
        return fmt.Errorf("invalid S3 URL format")
    }
    
    key := strings.Split(url, ".amazonaws.com/")[1]
    if key == "" {
        return fmt.Errorf("invalid S3 URL format: empty key extracted from URL")
    }
    
    _, err := h.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
        Bucket: &h.bucket,
        Key:    &key,
    })
    
    if err != nil {
        return fmt.Errorf("error deleting file from S3: %v", err)
    }
    
    return nil
}

func generateFilename(familyID models.FamilyID, prefix, originalName string) string {
    ext := filepath.Ext(originalName)
    timestamp := time.Now().UnixNano()
    return fmt.Sprintf("families/%s/%s/%d%s", familyID, prefix, timestamp, ext)
}

func (h *S3Handler) GetFamilyPath(familyID models.FamilyID) string {
    return fmt.Sprintf("families/%s", familyID)
}

func (h *S3Handler) GetProfileMediaPath(familyID models.FamilyID, profileID models.ProfileID, mediaType string) string {
    return fmt.Sprintf("families/%s/profiles/%s/media/%s", familyID, profileID, mediaType)
}

func (h *S3Handler) GetProfileAvatarPath(familyID models.FamilyID, profileID models.ProfileID) string {
    return fmt.Sprintf("families/%s/profiles/%s/avatar", familyID, profileID)
}

func (h *S3Handler) GetFamilySharedPath(familyID models.FamilyID, category string) string {
    return fmt.Sprintf("families/%s/shared/%s", familyID, category)
}