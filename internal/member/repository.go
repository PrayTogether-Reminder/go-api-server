package member

import (
	"context"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"

	"gorm.io/gorm"
)

// MemberRepository - Member 데이터 접근 계층
type MemberRepository struct {
}

// NewMemberRepository - MemberRepository 생성자
func NewMemberRepository() *MemberRepository {
	return &MemberRepository{}
}

// IsExistByEmail - 이메일 존재 여부 확인
func (m *MemberRepository) IsExistByEmail(ctx context.Context, db *gorm.DB, email string) (bool, error) {
	var count int64
	err := db.WithContext(ctx).
		Model(&model.Member{}).
		Where("email = ?", email).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// IsExistByID - ID 존재 여부 확인
func (m *MemberRepository) IsExistByID(ctx context.Context, db *gorm.DB, memberID int64) (bool, error) {
	var count int64
	err := db.WithContext(ctx).
		Model(&model.Member{}).
		Where("id = ?", memberID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Create - 회원 생성
func (m *MemberRepository) Create(ctx context.Context, db *gorm.DB, member *model.Member) error {
	return db.WithContext(ctx).Create(member).Error
}

// FindByEmail - 이메일로 회원 조회
func (m *MemberRepository) FindByEmail(ctx context.Context, db *gorm.DB, email string) (*model.Member, error) {
	var member model.Member
	err := db.WithContext(ctx).Where("email = ?", email).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// FindByID - ID로 회원 조회
func (m *MemberRepository) FindByID(ctx context.Context, db *gorm.DB, ID int64) (*model.Member, error) {
	var member model.Member
	err := db.WithContext(ctx).Where("id = ?", ID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// Delete - 회원 삭제
func (m *MemberRepository) Delete(ctx context.Context, db *gorm.DB, memberID int64) error {
	result := db.WithContext(ctx).Delete(&model.Member{}, memberID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
