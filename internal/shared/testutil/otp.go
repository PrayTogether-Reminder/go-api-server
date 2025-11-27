package testutil

import (
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/auth/otp"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/config"
)

// NewTestOTPService creates an OTP service for testing with mock SMTP sender
func NewTestOTPService() (*otp.Service, *otp.SMTPSender) {
	cache := otp.NewInMemoryCache()
	mockSender := NewRealSMTPSender()
	generator := otp.NewNumericGenerator(6)
	service := otp.NewService(cache, nil, generator, 3*time.Minute)
	return service, mockSender
}

// NewRealSMTPSender creates a real SMTP sender for integration tests
// Only use this if you have a real SMTP server configured
func NewRealSMTPSender() *otp.SMTPSender {
	cfg := config.SMTPConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
	}
	return otp.NewSMTPSender(cfg)
}
