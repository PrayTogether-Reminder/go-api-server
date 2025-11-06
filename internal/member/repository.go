package member

import (
	"context"
	"errors"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

type MemberRepository interface {
	Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error
	IsExist(ctx context.Context, db *gorm.DB, email string) (bool, error)
	Create(ctx context.Context, db *gorm.DB, member *model.Member) error
}

type memberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) MemberRepository {
	return &memberRepository{
		db: db,
	}
}

// Transaction wraps gorm.Transaction and ensures the provided context flows into the session.
func (m *memberRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	if fn == nil {
		return errors.New("member repository: transaction function is nil")
	}

	if ctx == nil {
		ctx = context.Background()
	}
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx.WithContext(ctx))
	})
}

func (m *memberRepository) IsExist(ctx context.Context, db *gorm.DB, email string) (bool, error) {
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

func (m *memberRepository) Create(ctx context.Context, db *gorm.DB, member *model.Member) error {
	return db.WithContext(ctx).Create(member).Error
}
