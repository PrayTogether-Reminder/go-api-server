package prayer

import (
	"context"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

type PrayerRepository struct{}

func NewPrayerRepository() *PrayerRepository {
	return &PrayerRepository{}
}

// Create creates a new prayer title in the database
func (r *PrayerRepository) Create(ctx context.Context, db *gorm.DB, prayerTitle *model.PrayerTitle) error {
	return db.WithContext(ctx).Create(prayerTitle).Error
}

// FindTitleInfosByRoomIDInitial fetches the most recent prayer titles in a room
func (r *PrayerRepository) FindTitleInfosByRoomIDInitial(ctx context.Context, db *gorm.DB, roomID int64, limit int) ([]PrayerTitleInfo, error) {
	var results []PrayerTitleInfo

	err := db.WithContext(ctx).
		Table((&model.PrayerTitle{}).TableName()).
		Select("id, title, created_time").
		Where("room_id = ?", roomID).
		Order("created_time DESC, id DESC").
		Limit(limit).
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

// FindTitleInfosByRoomIDAfter fetches prayer titles older than the cursor in a room
func (r *PrayerRepository) FindTitleInfosByRoomIDAfter(ctx context.Context, db *gorm.DB, roomID int64, cursorTime time.Time, limit int) ([]PrayerTitleInfo, error) {
	var results []PrayerTitleInfo

	err := db.WithContext(ctx).
		Table((&model.PrayerTitle{}).TableName()).
		Select("id, title, created_time").
		Where("room_id = ? AND created_time < ?", roomID, cursorTime).
		Order("created_time DESC, id DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// FindTitleByID finds a prayer title by its ID
func (r *PrayerRepository) FindTitleByID(ctx context.Context, db *gorm.DB, titleID int64) (*model.PrayerTitle, error) {
	var prayerTitle model.PrayerTitle
	if err := db.WithContext(ctx).Where("id = ?", titleID).First(&prayerTitle).Error; err != nil {
		return nil, err
	}
	return &prayerTitle, nil
}

// FindTitleByIDAndRoomID finds a prayer title by its ID and room ID
func (r *PrayerRepository) FindTitleByIDAndRoomID(ctx context.Context, db *gorm.DB, titleID int64, roomID int64) (*model.PrayerTitle, error) {
	var prayerTitle model.PrayerTitle
	err := db.WithContext(ctx).
		Where("id = ? AND room_id = ?", titleID, roomID).
		First(&prayerTitle).Error
	if err != nil {
		return nil, err
	}
	return &prayerTitle, nil
}

// CreateContent creates a new prayer content in the database
func (r *PrayerRepository) CreateContent(ctx context.Context, db *gorm.DB, prayerContent *model.PrayerContent) error {
	return db.WithContext(ctx).Create(prayerContent).Error
}

// FindContentsByTitleID finds all prayer contents by prayer title ID
func (r *PrayerRepository) FindContentsByTitleID(ctx context.Context, db *gorm.DB, titleID int64) ([]PrayerContentInfo, error) {
	var results []PrayerContentInfo

	err := db.WithContext(ctx).
		Table((&model.PrayerContent{}).TableName()).
		Select("id, writer_id, writer_name, member_id, member_name, content").
		Where("prayer_title_id = ?", titleID).
		Order("created_time ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// Update updates a prayer title in the database
func (r *PrayerRepository) Update(ctx context.Context, db *gorm.DB, prayerTitle *model.PrayerTitle) error {
	return db.WithContext(ctx).
		Model(&model.PrayerTitle{}).
		Where("id = ?", prayerTitle.ID).
		Updates(map[string]interface{}{
			"title": prayerTitle.Title,
		}).Error
}

// FindContentByID finds a prayer content by its ID
func (r *PrayerRepository) FindContentByID(ctx context.Context, db *gorm.DB, contentID int64) (*model.PrayerContent, error) {
	var prayerContent model.PrayerContent
	if err := db.WithContext(ctx).Where("id = ?", contentID).First(&prayerContent).Error; err != nil {
		return nil, err
	}
	return &prayerContent, nil
}

// ExistsContentInTitle checks if a prayer content exists in a specific title
func (r *PrayerRepository) ExistsContentInTitle(ctx context.Context, db *gorm.DB, contentID int64, titleID int64) (bool, error) {
	var count int64
	err := db.WithContext(ctx).
		Model(&model.PrayerContent{}).
		Where("id = ? AND prayer_title_id = ?", contentID, titleID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateContent updates a prayer content in the database
func (r *PrayerRepository) UpdateContent(ctx context.Context, db *gorm.DB, prayerContent *model.PrayerContent) error {
	return db.WithContext(ctx).
		Model(&model.PrayerContent{}).
		Where("id = ?", prayerContent.ID).
		Updates(map[string]interface{}{
			"content": prayerContent.Content,
		}).Error
}

// Delete deletes a prayer title by ID
// CASCADE 설정으로 연관된 PrayerContent도 자동 삭제됨
func (r *PrayerRepository) Delete(ctx context.Context, db *gorm.DB, title *model.PrayerTitle) error {
	return db.WithContext(ctx).
		Select("PrayerContents").
		Delete(&title).Error
}

// DeleteContent deletes a prayer content by ID
func (r *PrayerRepository) DeleteContent(ctx context.Context, db *gorm.DB, contentID int64) error {
	return db.WithContext(ctx).
		Where("id = ?", contentID).
		Delete(&model.PrayerContent{}).Error
}
