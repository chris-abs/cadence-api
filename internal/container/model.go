package container

import (
	"time"

	"github.com/chrisabs/storage/internal/item"
)

type Container struct {
	ID          int         `json:"id"`
	Name        string      `json:"name"`
	QRCode      string      `json:"qrCode"`
	QRCodeImage string      `json:"qrCodeImage"`
	Number      int         `json:"number"`
	Location    string      `json:"location"`
	UserID      int         `json:"userId"`
	Items       []item.Item `json:"items"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type CreateContainerRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	ItemIDs  []int  `json:"itemIds"`
}

type UpdateContainerRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	ItemIDs  []int  `json:"itemIds"`
}
