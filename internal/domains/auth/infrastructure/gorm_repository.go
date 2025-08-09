package infrastructure

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"pray-together/internal/domains/auth/domain"
)

// GormRepository implements auth domain repository using GORM
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM repository
func NewGormRepository(db *gorm.DB) domain.Repository {
	return &GormRepository{
		db: db,
	}
}

// RefreshToken operations

// CreateRefreshToken creates a new refresh token
func (r *GormRepository) CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// FindRefreshTokenByToken finds a refresh token by token string
func (r *GormRepository) FindRefreshTokenByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var refreshToken domain.RefreshToken
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&refreshToken).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &refreshToken, err
}

// FindRefreshTokensByMemberID finds refresh tokens by member ID
func (r *GormRepository) FindRefreshTokensByMemberID(ctx context.Context, memberID uint64) ([]*domain.RefreshToken, error) {
	var tokens []*domain.RefreshToken
	err := r.db.WithContext(ctx).Where("member_id = ?", memberID).Find(&tokens).Error
	return tokens, err
}

// DeleteRefreshToken deletes a refresh token
func (r *GormRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&domain.RefreshToken{}).Error
}

// DeleteRefreshTokensByMemberID deletes all refresh tokens for a member
func (r *GormRepository) DeleteRefreshTokensByMemberID(ctx context.Context, memberID uint64) error {
	return r.db.WithContext(ctx).Where("member_id = ?", memberID).Delete(&domain.RefreshToken{}).Error
}

// DeleteExpiredRefreshTokens deletes expired refresh tokens
func (r *GormRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&domain.RefreshToken{}).Error
}

// OTP operations

// CreateOTP creates a new OTP
func (r *GormRepository) CreateOTP(ctx context.Context, otp *domain.OTP) error {
	return r.db.WithContext(ctx).Create(otp).Error
}

// FindOTPByEmailAndCode finds an OTP by email and code
func (r *GormRepository) FindOTPByEmailAndCode(ctx context.Context, email, code, purpose string) (*domain.OTP, error) {
	var otp domain.OTP
	err := r.db.WithContext(ctx).
		Where("email = ? AND code = ? AND purpose = ?", email, code, purpose).
		First(&otp).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &otp, err
}

// FindLatestOTPByEmail finds the latest OTP for an email
func (r *GormRepository) FindLatestOTPByEmail(ctx context.Context, email, purpose string) (*domain.OTP, error) {
	var otp domain.OTP
	err := r.db.WithContext(ctx).
		Where("email = ? AND purpose = ?", email, purpose).
		Order("created_at DESC").
		First(&otp).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &otp, err
}

// UpdateOTP updates an OTP
func (r *GormRepository) UpdateOTP(ctx context.Context, otp *domain.OTP) error {
	return r.db.WithContext(ctx).Save(otp).Error
}

// DeleteOTP deletes an OTP
func (r *GormRepository) DeleteOTP(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.OTP{}, id).Error
}

// DeleteExpiredOTPs deletes expired OTPs
func (r *GormRepository) DeleteExpiredOTPs(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&domain.OTP{}).Error
}

// DeleteOTPsByEmail deletes all OTPs for an email
func (r *GormRepository) DeleteOTPsByEmail(ctx context.Context, email string) error {
	return r.db.WithContext(ctx).Where("email = ?", email).Delete(&domain.OTP{}).Error
}

// CleanupExpiredTokens cleans up expired tokens
func (r *GormRepository) CleanupExpiredTokens(ctx context.Context, before time.Time) error {
	// Clean up expired refresh tokens
	if err := r.db.WithContext(ctx).
		Where("expires_at < ?", before).
		Delete(&domain.RefreshToken{}).Error; err != nil {
		return err
	}

	// Clean up expired OTPs
	if err := r.db.WithContext(ctx).
		Where("expires_at < ?", before).
		Delete(&domain.OTP{}).Error; err != nil {
		return err
	}

	return nil
}
