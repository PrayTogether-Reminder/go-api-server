package fcm_token

import (
	"context"
	"fmt"
	"strings"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

// FcmTokenService contains business logic for registering and deleting FCM tokens.
type FcmTokenService struct {
	repo *FcmTokenRepository
}

// NewFcmTokenService returns a new service instance.
func NewFcmTokenService(repo *FcmTokenRepository) *FcmTokenService {
	return &FcmTokenService{repo: repo}
}

// RegisterToken enforces the single-token-per-member policy by deleting previous tokens
// and saving the new token within the provided transaction.
func (s *FcmTokenService) RegisterToken(ctx context.Context, db *gorm.DB, memberID int64, token string) error {
	normalized := strings.TrimSpace(token)
	if normalized == "" {
		return fmt.Errorf("공백 토큰은 등록할 수 없습니다")
	}

	if err := s.repo.DeleteByMemberID(ctx, db, memberID); err != nil {
		return fmt.Errorf("기존 FCM 토큰 삭제 실패: %w", err)
	}

	entity := &model.FcmToken{
		MemberID: memberID,
		Token:    normalized,
		IsActive: true,
	}
	if err := s.repo.Create(ctx, db, entity); err != nil {
		return fmt.Errorf("FCM 토큰 저장 실패: %w", err)
	}
	return nil
}

// DeleteByMemberID removes a token by value scoped to the given member.
func (s *FcmTokenService) DeleteByMemberID(ctx context.Context, db *gorm.DB, token string, memberID int64) error {
	if err := s.repo.DeleteByTokenAndMemberID(ctx, db, strings.TrimSpace(token), memberID); err != nil {
		return fmt.Errorf("FCM 토큰 삭제 실패: %w", err)
	}
	return nil
}
