package prayer

import "context"

type PrayerUseCase struct {
	prayerService *PrayerService
}

func NewPrayerUseCase(prayerService *PrayerService) *PrayerUseCase {
	return &PrayerUseCase{
		prayerService: prayerService,
	}
}

func (u *PrayerUseCase) CreatePrayerTitle(ctx context.Context, memberID int64, request *CreatePrayerTitleRequest) (*CreatePrayerTitleResponse, error) {
	return nil, nil
}
