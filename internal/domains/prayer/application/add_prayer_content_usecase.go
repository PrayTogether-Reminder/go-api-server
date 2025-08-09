package application

import (
	"context"
	"pray-together/internal/domains/prayer/domain"
)

// AddPrayerContentRequest represents the request to add prayer content
type AddPrayerContentRequest struct {
	PrayerTitleID uint64
	AuthorID      uint64
	Content       string
}

// AddPrayerContentUseCase handles adding content to a prayer title
type AddPrayerContentUseCase struct {
	prayerService *domain.Service
}

// NewAddPrayerContentUseCase creates a new add prayer content use case
func NewAddPrayerContentUseCase(prayerService *domain.Service) *AddPrayerContentUseCase {
	return &AddPrayerContentUseCase{
		prayerService: prayerService,
	}
}

// Execute adds content to a prayer title
func (u *AddPrayerContentUseCase) Execute(ctx context.Context, req *AddPrayerContentRequest) (*domain.PrayerContent, error) {
	// Get prayer title to verify it exists
	prayerTitle, err := u.prayerService.GetPrayerTitle(ctx, req.PrayerTitleID)
	if err != nil {
		return nil, err
	}
	if prayerTitle == nil {
		return nil, domain.ErrPrayerNotFound
	}

	// Create new content
	content := domain.NewPrayerContent(req.PrayerTitleID, req.AuthorID, req.Content)

	// Save to repository
	if err := u.prayerService.Repository.CreatePrayerContent(ctx, content); err != nil {
		return nil, err
	}

	return content, nil
}
