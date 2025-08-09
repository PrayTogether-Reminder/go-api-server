package application

import (
	"context"
	"pray-together/internal/domains/prayer/domain"
)

// UpdatePrayerContentRequest represents the request to update prayer content
type UpdatePrayerContentRequest struct {
	ContentID uint64
	MemberID  uint64
	Content   string
}

// UpdatePrayerContentUseCase handles updating prayer content
type UpdatePrayerContentUseCase struct {
	prayerService *domain.Service
}

// NewUpdatePrayerContentUseCase creates a new update prayer content use case
func NewUpdatePrayerContentUseCase(prayerService *domain.Service) *UpdatePrayerContentUseCase {
	return &UpdatePrayerContentUseCase{
		prayerService: prayerService,
	}
}

// Execute updates prayer content
func (u *UpdatePrayerContentUseCase) Execute(ctx context.Context, req *UpdatePrayerContentRequest) (*domain.PrayerContent, error) {
	// Get prayer content to find the title ID
	content, err := u.prayerService.Repository.FindPrayerContentByID(ctx, req.ContentID)
	if err != nil {
		return nil, err
	}
	if content == nil {
		return nil, domain.ErrPrayerContentNotFound
	}

	// Get prayer title and check permission
	prayerTitle, err := u.prayerService.GetPrayerTitle(ctx, content.PrayerTitleID)
	if err != nil {
		return nil, err
	}

	// Only the creator of the prayer title can update contents
	if prayerTitle.CreatorID != req.MemberID {
		return nil, domain.ErrUnauthorizedAccess
	}

	// Update content
	content.Content = req.Content
	if err := u.prayerService.Repository.UpdatePrayerContent(ctx, content); err != nil {
		return nil, err
	}

	return content, nil
}
