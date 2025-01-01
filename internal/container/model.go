package container

import (
	"time"
)

type Container struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	QRCode      string    `json:"qrCode"`
	QRCodeImage string    `json:"qrCodeImage"`
	Number      int       `json:"number"`
	Location    string    `json:"location"`
	UserID      int       `json:"userId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CreateContainerRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}
