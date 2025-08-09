package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"pray-together/internal/domains/member/domain"
)

// GormRepository implements member domain repository using GORM
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM repository
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{
		db: db,
	}
}

// Create creates a new member
func (r *GormRepository) Create(ctx context.Context, member *domain.Member) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// FindByID finds a member by ID
func (r *GormRepository) FindByID(ctx context.Context, id uint64) (*domain.Member, error) {
	var member domain.Member
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&member).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &member, err
}

// FindByEmail finds a member by email
func (r *GormRepository) FindByEmail(ctx context.Context, email string) (*domain.Member, error) {
	var member domain.Member
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&member).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &member, err
}

// FindMemberProfileByID finds member profile by ID
func (r *GormRepository) FindMemberProfileByID(ctx context.Context, id uint64) (*domain.MemberProfile, error) {
	var member domain.Member
	err := r.db.WithContext(ctx).
		Select("id", "email", "name").
		Where("id = ?", id).
		First(&member).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return member.ToProfile(), nil
}

// Update updates a member
func (r *GormRepository) Update(ctx context.Context, member *domain.Member) error {
	return r.db.WithContext(ctx).Save(member).Error
}

// Delete deletes a member (soft delete)
func (r *GormRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Member{}, id).Error
}

// ExistsByEmail checks if email exists
func (r *GormRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Member{}).
		Where("email = ?", email).
		Count(&count).Error

	return count > 0, err
}

// ExistsByID checks if member exists by ID
func (r *GormRepository) ExistsByID(ctx context.Context, id uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Member{}).
		Where("id = ?", id).
		Count(&count).Error

	return count > 0, err
}

// SearchByName searches members by name (partial match)
func (r *GormRepository) SearchByName(ctx context.Context, name string) ([]*domain.Member, error) {
	var members []*domain.Member
	err := r.db.WithContext(ctx).
		Where("name LIKE ?", "%"+name+"%").
		Limit(20). // Limit results for performance
		Find(&members).Error
	return members, err
}
