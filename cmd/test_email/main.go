package main

import (
	"fmt"
	"os"

	"github.com/chrisabs/cadence/internal/email"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/test_email/main.go recipient@example.com")
		os.Exit(1)
	}
	
	recipientEmail := os.Args[1]
	
	emailService, err := email.NewService()
	if err != nil {
		fmt.Printf("Failed to create email service: %v\n", err)
		os.Exit(1)
	}
	
	err = emailService.SendInviteEmail(
		recipientEmail,
		"TEST_TOKEN_1234567890",
		"Test Family",
	)
	
	if err != nil {
		fmt.Printf("Failed to send test email: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Test email sent successfully to %s\n", recipientEmail)
}