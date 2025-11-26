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

// GetCachedOTP는 테스트를 위한 헬퍼입니다.
func (s *Service) GetCachedOTP(email string) (string, bool) {
	return s.cache.Get(email)
}
