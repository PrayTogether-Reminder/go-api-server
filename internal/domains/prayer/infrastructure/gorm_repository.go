package infrastructure

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"pray-together/internal/domains/prayer/domain"
)

// GormRepository implements prayer domain repository using GORM
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM repository
func NewGormRepository(db *gorm.DB) domain.Repository {
	return &GormRepository{
		db: db,
	}
}

// Create creates a new prayer
func (r *GormRepository) Create(ctx context.Context, prayer *domain.Prayer) error {
	return r.db.WithContext(ctx).Create(prayer).Error
}

// FindByID finds a prayer by ID
func (r *GormRepository) FindByID(ctx context.Context, id uint64) (*domain.Prayer, error) {
	var prayer domain.Prayer
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&prayer).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &prayer, err
}

// Update updates a prayer
func (r *GormRepository) Update(ctx context.Context, prayer *domain.Prayer) error {
	return r.db.WithContext(ctx).Save(prayer).Error
}

// Delete deletes a prayer
func (r *GormRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Prayer{}, id).Error
}

// FindByMemberID finds prayers by member ID
func (r *GormRepository) FindByMemberID(ctx context.Context, memberID uint64, limit, offset int) ([]*domain.Prayer, error) {
	var prayers []*domain.Prayer

	query := r.db.WithContext(ctx).Where("member_id = ?", memberID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&prayers).Error
	return prayers, err
}

// FindByRoomID finds prayers by room ID
func (r *GormRepository) FindByRoomID(ctx context.Context, roomID uint64, limit, offset int) ([]*domain.Prayer, error) {
	var prayers []*domain.Prayer

	query := r.db.WithContext(ctx).Where("room_id = ?", roomID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&prayers).Error
	return prayers, err
}

// FindByMemberAndRoom finds prayers by member and room
func (r *GormRepository) FindByMemberAndRoom(ctx context.Context, memberID, roomID uint64, limit, offset int) ([]*domain.Prayer, error) {
	var prayers []*domain.Prayer

	query := r.db.WithContext(ctx).
		Where("member_id = ? AND room_id = ?", memberID, roomID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&prayers).Error
	return prayers, err
}

// FindSharedPrayersByRoom finds shared prayers in a room
func (r *GormRepository) FindSharedPrayersByRoom(ctx context.Context, roomID uint64, limit, offset int) ([]*domain.Prayer, error) {
	var prayers []*domain.Prayer

	query := r.db.WithContext(ctx).
		Where("room_id = ? AND type = ?", roomID, domain.PrayerTypeShared).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&prayers).Error
	return prayers, err
}

// FindAnsweredPrayers finds answered prayers
func (r *GormRepository) FindAnsweredPrayers(ctx context.Context, memberID uint64, limit, offset int) ([]*domain.Prayer, error) {
	var prayers []*domain.Prayer

	query := r.db.WithContext(ctx).
		Where("member_id = ? AND is_answered = ?", memberID, true).
		Order("answered_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&prayers).Error
	return prayers, err
}

// FindUnansweredPrayers finds unanswered prayers
func (r *GormRepository) FindUnansweredPrayers(ctx context.Context, memberID uint64, limit, offset int) ([]*domain.Prayer, error) {
	var prayers []*domain.Prayer

	query := r.db.WithContext(ctx).
		Where("member_id = ? AND is_answered = ?", memberID, false).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&prayers).Error
	return prayers, err
}

// FindPrayersByDateRange finds prayers in a date range
func (r *GormRepository) FindPrayersByDateRange(ctx context.Context, memberID uint64, startDate, endDate time.Time) ([]*domain.Prayer, error) {
	var prayers []*domain.Prayer

	err := r.db.WithContext(ctx).
		Where("member_id = ? AND created_at BETWEEN ? AND ?", memberID, startDate, endDate).
		Order("created_at DESC").
		Find(&prayers).Error

	return prayers, err
}

// CountByMemberID counts prayers by member ID
func (r *GormRepository) CountByMemberID(ctx context.Context, memberID uint64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Prayer{}).
		Where("member_id = ?", memberID).
		Count(&count).Error

	return int(count), err
}

// CountByRoomID counts prayers by room ID
func (r *GormRepository) CountByRoomID(ctx context.Context, roomID uint64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Prayer{}).
		Where("room_id = ?", roomID).
		Count(&count).Error

	return int(count), err
}

// CountAnsweredByMemberID counts answered prayers by member ID
func (r *GormRepository) CountAnsweredByMemberID(ctx context.Context, memberID uint64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Prayer{}).
		Where("member_id = ? AND is_answered = ?", memberID, true).
		Count(&count).Error

	return int(count), err
}

// CountSharedByRoomID counts shared prayers by room ID
func (r *GormRepository) CountSharedByRoomID(ctx context.Context, roomID uint64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Prayer{}).
		Where("room_id = ? AND type = ?", roomID, domain.PrayerTypeShared).
		Count(&count).Error

	return int(count), err
}

// ExistsByID checks if prayer exists by ID
func (r *GormRepository) ExistsByID(ctx context.Context, id uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Prayer{}).
		Where("id = ?", id).
		Count(&count).Error

	return count > 0, err
}

// PrayerTitle operations

// CreatePrayerTitle creates a new prayer title
func (r *GormRepository) CreatePrayerTitle(ctx context.Context, title *domain.PrayerTitle) error {
	return r.db.WithContext(ctx).Create(title).Error
}

// FindPrayerTitleByID finds a prayer title by ID
func (r *GormRepository) FindPrayerTitleByID(ctx context.Context, id uint64) (*domain.PrayerTitle, error) {
	var title domain.PrayerTitle
	err := r.db.WithContext(ctx).First(&title, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPrayerNotFound
		}
		return nil, err
	}
	return &title, nil
}

// FindPrayerTitleWithContents finds a prayer title with its contents
func (r *GormRepository) FindPrayerTitleWithContents(ctx context.Context, id uint64) (*domain.PrayerTitle, error) {
	var title domain.PrayerTitle
	err := r.db.WithContext(ctx).Preload("Contents").First(&title, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPrayerNotFound
		}
		return nil, err
	}
	return &title, nil
}

// FindPrayerTitlesByRoomID finds prayer titles by room ID with pagination
func (r *GormRepository) FindPrayerTitlesByRoomID(ctx context.Context, roomID uint64, after uint64, limit int) ([]*domain.PrayerTitle, error) {
	var titles []*domain.PrayerTitle
	query := r.db.WithContext(ctx).Where("room_id = ?", roomID)

	if after > 0 {
		query = query.Where("id < ?", after)
	}

	err := query.Order("id DESC").Limit(limit).Find(&titles).Error
	return titles, err
}

// UpdatePrayerTitle updates a prayer title
func (r *GormRepository) UpdatePrayerTitle(ctx context.Context, title *domain.PrayerTitle) error {
	return r.db.WithContext(ctx).Save(title).Error
}

// DeletePrayerTitle deletes a prayer title (soft delete)
func (r *GormRepository) DeletePrayerTitle(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.PrayerTitle{}, id).Error
}

// PrayerContent operations

// CreatePrayerContent creates a new prayer content
func (r *GormRepository) CreatePrayerContent(ctx context.Context, content *domain.PrayerContent) error {
	return r.db.WithContext(ctx).Create(content).Error
}

// CreatePrayerContents creates multiple prayer contents
func (r *GormRepository) CreatePrayerContents(ctx context.Context, contents []*domain.PrayerContent) error {
	if len(contents) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&contents).Error
}

// FindPrayerContentByID finds a prayer content by ID
func (r *GormRepository) FindPrayerContentByID(ctx context.Context, id uint64) (*domain.PrayerContent, error) {
	var content domain.PrayerContent
	err := r.db.WithContext(ctx).First(&content, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &content, err
}

// FindPrayerContentsByTitleID finds prayer contents by title ID
func (r *GormRepository) FindPrayerContentsByTitleID(ctx context.Context, titleID uint64) ([]*domain.PrayerContent, error) {
	var contents []*domain.PrayerContent
	err := r.db.WithContext(ctx).Where("prayer_title_id = ?", titleID).Find(&contents).Error
	return contents, err
}

// UpdatePrayerContent updates a prayer content
func (r *GormRepository) UpdatePrayerContent(ctx context.Context, content *domain.PrayerContent) error {
	return r.db.WithContext(ctx).Save(content).Error
}

// DeletePrayerContent deletes a specific prayer content
func (r *GormRepository) DeletePrayerContent(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.PrayerContent{}, id).Error
}

// DeletePrayerContentsByTitleID deletes prayer contents by title ID
func (r *GormRepository) DeletePrayerContentsByTitleID(ctx context.Context, titleID uint64) error {
	return r.db.WithContext(ctx).Where("prayer_title_id = ?", titleID).Delete(&domain.PrayerContent{}).Error
}

// PrayerCompletion operations

// CreatePrayerCompletion creates a new prayer completion
func (r *GormRepository) CreatePrayerCompletion(ctx context.Context, completion *domain.PrayerCompletion) error {
	return r.db.WithContext(ctx).Create(completion).Error
}

// FindPrayerCompletionsByTitleID finds prayer completions by title ID
func (r *GormRepository) FindPrayerCompletionsByTitleID(ctx context.Context, titleID uint64) ([]*domain.PrayerCompletion, error) {
	var completions []*domain.PrayerCompletion
	err := r.db.WithContext(ctx).Where("prayer_title_id = ?", titleID).Find(&completions).Error
	return completions, err
}

// ExistsPrayerCompletionByMemberAndTitle checks if a member has completed a prayer
func (r *GormRepository) ExistsPrayerCompletionByMemberAndTitle(ctx context.Context, memberID, titleID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.PrayerCompletion{}).
		Where("member_id = ? AND prayer_title_id = ?", memberID, titleID).
		Count(&count).Error
	return count > 0, err
}

// CountByRoomIDSince counts prayers in a room since a given time
func (r *GormRepository) CountByRoomIDSince(ctx context.Context, roomID uint64, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Prayer{}).
		Where("room_id = ? AND created_at >= ?", roomID, since).
		Count(&count).Error
	return count, err
}

// FindFirstPrayerTitleInfosByRoomID finds first page of prayer title infos (matching Java)
func (r *GormRepository) FindFirstPrayerTitleInfosByRoomID(ctx context.Context, roomID uint64, limit int) ([]*domain.PrayerTitleInfo, error) {
	var titles []*domain.PrayerTitle
	err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at DESC").
		Limit(limit).
		Find(&titles).Error

	if err != nil {
		return nil, err
	}

	// Convert to PrayerTitleInfo
	infos := make([]*domain.PrayerTitleInfo, len(titles))
	for i, title := range titles {
		infos[i] = title.ToPrayerTitleInfo()
	}

	return infos, nil
}

// FindPrayerTitleInfosByRoomIDAfterTime finds prayer title infos after a specific time (matching Java)
func (r *GormRepository) FindPrayerTitleInfosByRoomIDAfterTime(ctx context.Context, roomID uint64, afterTime time.Time, limit int) ([]*domain.PrayerTitleInfo, error) {
	var titles []*domain.PrayerTitle
	err := r.db.WithContext(ctx).
		Where("room_id = ? AND created_at < ?", roomID, afterTime).
		Order("created_at DESC").
		Limit(limit).
		Find(&titles).Error

	if err != nil {
		return nil, err
	}

	// Convert to PrayerTitleInfo
	infos := make([]*domain.PrayerTitleInfo, len(titles))
	for i, title := range titles {
		infos[i] = title.ToPrayerTitleInfo()
	}

	return infos, nil
}
