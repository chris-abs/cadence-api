package utils

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/skip2/go-qrcode"
)

func GenerateQRCode(containerID int) (string, string, error) {
	qrString := fmt.Sprintf("STQRAGE-CONTAINER-%d-%d", containerID, time.Now().Unix())

	qr, err := qrcode.Encode(qrString, qrcode.Medium, 256)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate QR code: %v", err)
	}

	qrBase64 := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(qr))

	return qrString, qrBase64, nil
}
