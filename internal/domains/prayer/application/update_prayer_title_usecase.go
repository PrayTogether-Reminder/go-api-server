package application

import (
	"context"
	"pray-together/internal/domains/prayer/domain"
)

// UpdatePrayerTitleRequest represents the request to update a prayer title
type UpdatePrayerTitleRequest struct {
	PrayerTitleID uint64
	MemberID      uint64
	Title         string
	Contents      []PrayerContentRequest
}

// UpdatePrayerTitleUseCase handles updating a prayer title
type UpdatePrayerTitleUseCase struct {
	prayerService *domain.Service
}

// NewUpdatePrayerTitleUseCase creates a new update prayer title use case
func NewUpdatePrayerTitleUseCase(prayerService *domain.Service) *UpdatePrayerTitleUseCase {
	return &UpdatePrayerTitleUseCase{
		prayerService: prayerService,
	}
}

// Execute updates a prayer title and its contents
func (u *UpdatePrayerTitleUseCase) Execute(ctx context.Context, req *UpdatePrayerTitleRequest) (*domain.PrayerTitle, error) {
	// Validate member exists in room first (matching Java)
	prayerTitle, err := u.prayerService.GetPrayerTitle(ctx, req.PrayerTitleID)
	if err != nil {
		return nil, err
	}

	// Validate member is in the room
	if err := u.prayerService.ValidateRoomAccess(ctx, prayerTitle.RoomID, req.MemberID); err != nil {
		return nil, err
	}

	// Update title
	if req.Title != "" {
		prayerTitle.Title = req.Title
	}

	// Update title
	if err := u.prayerService.Repository.UpdatePrayerTitle(ctx, prayerTitle); err != nil {
		return nil, err
	}

	// Update contents (replace all contents - matching Java behavior)
	// First delete existing contents
	existingContents, _ := u.prayerService.Repository.FindPrayerContentsByTitleID(ctx, req.PrayerTitleID)
	for _, content := range existingContents {
		_ = u.prayerService.Repository.DeletePrayerContent(ctx, content.ID)
	}

	// Add new contents
	for _, content := range req.Contents {
		newContent := &domain.PrayerContent{
			PrayerTitleID: req.PrayerTitleID,
			AuthorID:      req.MemberID,
			Content:       content.Content,
		}
		_ = u.prayerService.Repository.CreatePrayerContent(ctx, newContent)
	}

	return prayerTitle, nil
}
