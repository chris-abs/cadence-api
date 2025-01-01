package main

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"

	"github.com/skip2/go-qrcode"
)

type User struct {
	ID         uint        `json:"id"`
	Email      string      `json:"email"`
	Password   string      `json:"-"`
	FirstName  string      `json:"firstName"`
	LastName   string      `json:"lastName"`
	Containers []Container `json:"containers"`
	CreatedAt  time.Time   `json:"createdAt"`
	UpdatedAt  time.Time   `json:"updatedAt"`
}

type CreateContainerRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type Container struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	QRCode      string `json:"qrCode"`
	QRCodeImage string `json:"qrCodeImage"`
	Number      int    `json:"number"`
	Location    string `json:"location"`
	// Items       []Item    `json:"items"`
	// image
	UserId    int       `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func generateQRCode(containerID int) (string, string, error) {
	qrString := fmt.Sprintf("STQRAGE-CONTAINER-%d-%d", containerID, time.Now().Unix())

	qr, err := qrcode.Encode(qrString, qrcode.Medium, 256)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate QR code: %v", err)
	}

	qrBase64 := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(qr))

	return qrString, qrBase64, nil
}

func NewContainer(name, location string) *Container {
	containerID := rand.Intn(10000)
	qrString, qrImage, err := generateQRCode(containerID)
	if err != nil {
		qrString = fmt.Sprintf("STQRAGE-CONTAINER-%d", containerID)
		qrImage = ""
	}

	return &Container{
		ID:          containerID,
		Name:        name,
		QRCode:      qrString,
		QRCodeImage: qrImage,
		Number:      rand.Intn(1000),
		Location:    location,
		UserId:      rand.Intn(10000),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
}
