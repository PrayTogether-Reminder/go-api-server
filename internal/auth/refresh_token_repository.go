package auth

import (
	"context"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

// RefreshTokenRepository - RefreshToken 데이터 접근 계층
type RefreshTokenRepository struct {
}

// NewRefreshTokenRepository - RefreshTokenRepository 생성자
func NewRefreshTokenRepository() *RefreshTokenRepository {
	return &RefreshTokenRepository{}
}

// Save - Refresh Token 저장 (기존 토큰이 있으면 삭제 후 저장)
// 회원당 하나의 Refresh Token만 유지
func (r *RefreshTokenRepository) Save(ctx context.Context, db *gorm.DB, memberID int64, token string, expiresAt time.Time) error {
	// 기존 토큰 삭제 (존재하지 않아도 에러 없음)
	_ = r.Delete(ctx, db, memberID)

	// 새로운 토큰 저장
	refreshToken := model.NewRefreshToken(memberID, token, expiresAt)
	return db.WithContext(ctx).Create(refreshToken).Error
}

// FindByMemberID - MemberID로 RefreshToken 조회
func (r *RefreshTokenRepository) FindByMemberID(ctx context.Context, db *gorm.DB, memberID int64) (*model.RefreshToken, error) {
	var refreshToken model.RefreshToken
	err := db.WithContext(ctx).
		Where("member_id = ?", memberID).
		First(&refreshToken).Error

	if err != nil {
		return nil, err
	}

	return &refreshToken, nil
}

// Delete - MemberID로 RefreshToken 삭제
func (r *RefreshTokenRepository) Delete(ctx context.Context, db *gorm.DB, memberID int64) error {
	return db.WithContext(ctx).
		Where("member_id = ?", memberID).
		Delete(&model.RefreshToken{}).Error
}

// Exists - MemberID로 RefreshToken 존재 여부 확인
func (r *RefreshTokenRepository) Exists(ctx context.Context, db *gorm.DB, memberID int64, token string) (bool, error) {
	var count int64
	err := db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("member_id = ? AND token = ?", memberID, token).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
