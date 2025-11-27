package member

import (
	"context"
	"strings"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"

	"gorm.io/gorm"
)

// MemberRepository - Member 데이터 접근 계층
type MemberRepository struct {
}

// SearchMemberResult - 회원 검색 결과 프로젝션
type SearchMemberResult struct {
	ID          int64
	Name        string
	PhoneNumber string
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

// FindByIDs - ID 리스트로 회원 일괄 조회
func (m *MemberRepository) FindByIDs(ctx context.Context, db *gorm.DB, memberIDs []int64) ([]*model.Member, error) {
	if len(memberIDs) == 0 {
		return []*model.Member{}, nil
	}

	var members []*model.Member
	err := db.WithContext(ctx).Where("id IN ?", memberIDs).Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
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

// UpdateFields - 특정 필드 업데이트
func (m *MemberRepository) UpdateFields(ctx context.Context, db *gorm.DB, memberID int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	result := db.WithContext(ctx).
		Model(&model.Member{}).
		Where("id = ?", memberID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// SearchByName - 이름으로 회원 검색 (부분 일치)
func (m *MemberRepository) SearchByName(ctx context.Context, db *gorm.DB, name string) ([]SearchMemberResult, error) {
	var results []SearchMemberResult

	query := db.WithContext(ctx).
		Model(&model.Member{}).
		Select("id, name, phone_number")

	if trimmed := strings.TrimSpace(name); trimmed != "" {
		query = query.Where("name LIKE ?", "%"+trimmed+"%")
	}

	if err := query.Order("id ASC").Find(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}
