package entities

import (
	"time"

	"github.com/chrisabs/cadence/internal/models"
)

type Container struct {
    ID           models.ContainerID  `json:"id"`
    Name         string              `json:"name"`
    Description  string              `json:"description"`
    QRCode       string              `json:"qrCode"`
    QRCodeImage  string              `json:"qrCodeImage"`
    Number       int                 `json:"number"`
    Location     string              `json:"location"`
    ProfileID    models.ProfileID    `json:"profileId"`
    FamilyID     models.FamilyID     `json:"familyId"`
    WorkspaceID  *models.WorkspaceID `json:"workspaceId,omitempty"`
    Workspace    *Workspace          `json:"workspace,omitempty"`
    Items        []Item              `json:"items"`
    CreatedAt    time.Time           `json:"createdAt"`
    UpdatedAt    time.Time           `json:"updatedAt"`
}