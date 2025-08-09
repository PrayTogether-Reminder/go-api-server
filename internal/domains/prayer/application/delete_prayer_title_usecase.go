package application

import (
	"context"
	"pray-together/internal/domains/prayer/domain"
)

// DeletePrayerTitleRequest represents the request to delete a prayer title
type DeletePrayerTitleRequest struct {
	PrayerTitleID uint64
	MemberID      uint64
}

// DeletePrayerTitleUseCase handles deleting a prayer title
type DeletePrayerTitleUseCase struct {
	prayerService *domain.Service
}

// NewDeletePrayerTitleUseCase creates a new delete prayer title use case
func NewDeletePrayerTitleUseCase(prayerService *domain.Service) *DeletePrayerTitleUseCase {
	return &DeletePrayerTitleUseCase{
		prayerService: prayerService,
	}
}

// Execute deletes a prayer title
func (u *DeletePrayerTitleUseCase) Execute(ctx context.Context, req *DeletePrayerTitleRequest) error {
	return u.prayerService.DeletePrayerTitle(ctx, req.PrayerTitleID, req.MemberID)
}
