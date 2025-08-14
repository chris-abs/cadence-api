package utils

import (
	"encoding/base64"
	"fmt"

	"github.com/skip2/go-qrcode"
)

func GenerateQRCode(containerID string) (string, string, error) {
	qrString := containerID

	qr, err := qrcode.Encode(qrString, qrcode.Medium, 256)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate QR code: %v", err)
	}

	qrBase64 := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(qr))

	return qrString, qrBase64, nil
}