package application

import (
	"context"
	"pray-together/internal/domains/prayer/domain"
)

// DeletePrayerContentRequest represents the request to delete prayer content
type DeletePrayerContentRequest struct {
	ContentID uint64
	MemberID  uint64
}

// DeletePrayerContentUseCase handles deleting prayer content
type DeletePrayerContentUseCase struct {
	prayerService *domain.Service
}

// NewDeletePrayerContentUseCase creates a new delete prayer content use case
func NewDeletePrayerContentUseCase(prayerService *domain.Service) *DeletePrayerContentUseCase {
	return &DeletePrayerContentUseCase{
		prayerService: prayerService,
	}
}

// Execute deletes prayer content
func (u *DeletePrayerContentUseCase) Execute(ctx context.Context, req *DeletePrayerContentRequest) error {
	// Get prayer content to find the title ID
	content, err := u.prayerService.Repository.FindPrayerContentByID(ctx, req.ContentID)
	if err != nil {
		return err
	}
	if content == nil {
		return domain.ErrPrayerContentNotFound
	}

	// Get prayer title and check permission
	prayerTitle, err := u.prayerService.GetPrayerTitle(ctx, content.PrayerTitleID)
	if err != nil {
		return err
	}

	// Only the creator of the prayer title can delete contents
	if prayerTitle.CreatorID != req.MemberID {
		return domain.ErrUnauthorizedAccess
	}

	// Delete the specific content
	return u.prayerService.Repository.DeletePrayerContent(ctx, req.ContentID)
}
