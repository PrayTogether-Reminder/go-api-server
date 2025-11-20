package prayer

import (
	"context"
	"fmt"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

type PrayerService struct {
	prayerRepository *PrayerRepository
}

func NewPrayerService(repo *PrayerRepository) *PrayerService {
	return &PrayerService{prayerRepository: repo}
}

// CreatePrayerTitle creates a new prayer title
func (s *PrayerService) CreatePrayerTitle(ctx context.Context, tx *gorm.DB, room *model.Room, title string) (*model.PrayerTitle, error) {
	prayerTitle := model.NewPrayerTitle(room, title)

	// Repository로 저장
	if err := s.prayerRepository.Create(ctx, tx, prayerTitle); err != nil {
		return nil, fmt.Errorf("기도제목 생성 실패: %w", err)
	}

	return prayerTitle, nil
}
