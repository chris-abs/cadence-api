package entities

import "time"

type Container struct {
    ID           int        `json:"id"`
    Name         string     `json:"name"`
    Description string      `json:"description"`
    QRCode       string     `json:"qrCode"`
    QRCodeImage  string     `json:"qrCodeImage"`
    Number       int        `json:"number"`
    Location     string     `json:"location"`
    UserID       int        `json:"userId"`
    FamilyID    int        `json:"familyId"`
    WorkspaceID  *int       `json:"workspaceId,omitempty"`
    Workspace    *Workspace `json:"workspace,omitempty"`
    Items        []Item     `json:"items"`
    CreatedAt    time.Time  `json:"createdAt"`
    UpdatedAt    time.Time  `json:"updatedAt"`
}
