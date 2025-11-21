package prayer

import (
	"context"

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
