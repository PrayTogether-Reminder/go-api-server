package member

import (
	"context"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

type MemberRepository interface {
	IsExist(ctx context.Context, db *gorm.DB, email string) (bool, error)
	Create(ctx context.Context, db *gorm.DB, member *model.Member) error
	FindByEmail(ctx context.Context, db *gorm.DB, email string) (*model.Member, error)
}

type memberRepository struct{}

func NewMemberRepository() MemberRepository {
	return &memberRepository{}
}

func (m *memberRepository) IsExist(ctx context.Context, db *gorm.DB, email string) (bool, error) {
	var count int64
	err := db.Model(&model.Member{}).
		Where("email = ?", email).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (m *memberRepository) Create(ctx context.Context, db *gorm.DB, member *model.Member) error {
	return db.Create(member).Error
}

func (m *memberRepository) FindByEmail(ctx context.Context, db *gorm.DB, email string) (*model.Member, error) {
	var member model.Member
	err := db.Where("email = ?", email).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}
