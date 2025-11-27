package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// RefreshTokenService - RefreshToken 비즈니스 로직 계층
type RefreshTokenService struct {
	repo *RefreshTokenRepository
}

// NewRefreshTokenService - RefreshTokenService 생성자
func NewRefreshTokenService(repo *RefreshTokenRepository) *RefreshTokenService {
	return &RefreshTokenService{
		repo: repo,
	}
}

// Save - Refresh Token 저장
func (s *RefreshTokenService) Save(ctx context.Context, db *gorm.DB, memberID int64, token string, expiresAt time.Time) error {
	return s.repo.Save(ctx, db, memberID, token, expiresAt)
}

// ValidateExist - Refresh Token 존재 및 유효성 검증
func (s *RefreshTokenService) ValidateExist(ctx context.Context, db *gorm.DB, memberID int64, requestToken string) error {
	// 저장된 토큰 조회
	storedToken, err := s.repo.FindByMemberID(ctx, db, memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("RefreshToken 조회 실패: %w", ErrRefreshTokenNotFound)
		}
		return fmt.Errorf("RefreshToken 조회 중 오류 발생: %w", err)
	}

	// 만료 여부 확인
	if storedToken.IsExpired() {
		// 만료된 토큰 삭제
		_ = s.repo.Delete(ctx, db, memberID)
		return fmt.Errorf("RefreshToken 만료: %w", ErrRefreshTokenExpired)
	}

	// 토큰 값 일치 여부 확인
	if storedToken.Token != requestToken {
		return fmt.Errorf("RefreshToken 불일치: %w", ErrRefreshTokenMismatch)
	}

	return nil
}

// Delete - Refresh Token 삭제
func (s *RefreshTokenService) Delete(ctx context.Context, db *gorm.DB, memberID int64) error {
	if err := s.repo.Delete(ctx, db, memberID); err != nil {
		return fmt.Errorf("RefreshToken 삭제 실패: %w", err)
	}
	return nil
}
