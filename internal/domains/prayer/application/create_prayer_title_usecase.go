package application

import (
	"context"
	"pray-together/internal/domains/prayer/domain"
)

// PrayerContentRequest represents prayer content in request
type PrayerContentRequest struct {
	MemberID   uint64
	MemberName string
	Content    string
}

// CreatePrayerTitleRequest represents the request to create a prayer title
type CreatePrayerTitleRequest struct {
	RoomID    uint64
	CreatorID uint64
	Title     string
	Contents  []PrayerContentRequest
}

// CreatePrayerTitleUseCase handles creating a new prayer title
type CreatePrayerTitleUseCase struct {
	prayerService *domain.Service
}

// NewCreatePrayerTitleUseCase creates a new create prayer title use case
func NewCreatePrayerTitleUseCase(prayerService *domain.Service) *CreatePrayerTitleUseCase {
	return &CreatePrayerTitleUseCase{
		prayerService: prayerService,
	}
}

// Execute creates a new prayer title
func (u *CreatePrayerTitleUseCase) Execute(ctx context.Context, req *CreatePrayerTitleRequest) (*domain.PrayerTitle, error) {
	// Validate member exists in room first
	if err := u.prayerService.ValidateRoomAccess(ctx, req.RoomID, req.CreatorID); err != nil {
		return nil, err
	}

	// Convert PrayerContentRequest to strings for now
	// TODO: Update domain to support full content structure
	contents := make([]string, len(req.Contents))
	for i, content := range req.Contents {
		contents[i] = content.Content
	}

	// Create prayer title with contents
	return u.prayerService.CreatePrayerTitle(ctx, req.RoomID, req.CreatorID, req.Title, contents)
}
