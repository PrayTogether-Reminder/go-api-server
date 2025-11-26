package otp

import (
	"context"
	"fmt"
	"time"
)

// Service coordinates OTP generation, caching, and delivery.
type Service struct {
	cache     Cache
	sender    *SMTPSender
	generator *NumericGenerator
	ttl       time.Duration
}

func NewService(cache Cache, sender *SMTPSender, generator *NumericGenerator, ttl time.Duration) *Service {
	return &Service{cache: cache, sender: sender, generator: generator, ttl: ttl}
}

func (s *Service) SendToEmail(ctx context.Context, email string) error {
	otp, err := s.generator.Generate()
	if err != nil {
		return fmt.Errorf("OTP 생성 중 오류: %w", err)
	}

	if err := s.sender.SendOTP(ctx, email, otp); err != nil {
		return fmt.Errorf("OTP 발송 실패: %v: %w", err, ErrSendFailed)
	}
	s.cache.Set(email, otp, s.ttl)

	return nil
}

// VerifyOTP verifies the OTP for the given email.
// Returns true if the OTP matches, false otherwise.
// If no OTP exists for the email, returns an error.
func (s *Service) VerifyOTP(ctx context.Context, email, otp string) (bool, error) {
	cachedOTP, exists := s.cache.Get(email)
	if !exists {
		return false, fmt.Errorf("등록되지 않은 OTP 입니다: %w", ErrOTPNotFound)
	}

	if cachedOTP != otp {
		return false, nil
	}

	// OTP 일치 시 캐시에서 즉시 삭제 (재사용 방지)
	s.cache.Delete(email)
	return true, nil
}

// GetCachedOTP는 테스트를 위한 헬퍼입니다.
func (s *Service) GetCachedOTP(email string) (string, bool) {
	return s.cache.Get(email)
}
