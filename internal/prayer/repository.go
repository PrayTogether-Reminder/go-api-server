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
