package otp

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"path/filepath"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/config"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/logger"
)

// SMTPSender sends OTP via email using SMTP.
type SMTPSender struct {
	config config.SMTPConfig
}

// NewSMTPSender creates a new SMTP-based OTP sender.
func NewSMTPSender(cfg config.SMTPConfig) *SMTPSender {
	return &SMTPSender{
		config: cfg,
	}
}

// SendOTP sends an OTP to the given email address.
func (s *SMTPSender) SendOTP(ctx context.Context, email, otp string) error {
	log := logger.FromContext(ctx)

	// Load and parse HTML template
	htmlBody, err := s.renderTemplate(otp)
	if err != nil {
		return fmt.Errorf("OTP 템플릿 로드 실패: %w", err)
	}

	// Prepare email message
	subject := "기도함께 이메일 인증번호"
	message := s.buildMessage(s.config.From, email, subject, htmlBody)

	// SMTP authentication
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// Send email
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	err = smtp.SendMail(addr, auth, s.config.From, []string{email}, []byte(message))
	if err != nil {
		return fmt.Errorf("OTP 이메일 발송 실패: email=%s %w", logger.MaskEmail(email), err)
	}

	log.Info("OTP 발송 성공", "email", logger.MaskEmail(email))
	return nil
}

// renderTemplate loads and renders the OTP HTML template.
func (s *SMTPSender) renderTemplate(otp string) (string, error) {
	tmpl, err := template.ParseFiles(GetDefaultTemplatePath())
	if err != nil {
		return "", fmt.Errorf("템플릿 파일 읽기 실패: %w", err)
	}

	var buf bytes.Buffer
	data := map[string]string{"OTP": otp}
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("템플릿 렌더링 실패: %w", err)
	}

	return buf.String(), nil
}

// buildMessage constructs the email message with headers.
func (s *SMTPSender) buildMessage(from, to, subject, htmlBody string) string {
	headers := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n",
		from, to, subject)

	return headers + htmlBody
}

// GetDefaultTemplatePath returns the default template path.
func GetDefaultTemplatePath() string {
	return filepath.Join("templates", "email", "otp.html")
}
