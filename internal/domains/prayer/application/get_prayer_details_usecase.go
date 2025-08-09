package application

import (
	"context"
	"pray-together/internal/domains/prayer/domain"
)

// GetPrayerDetailsRequest represents the request to get prayer details
type GetPrayerDetailsRequest struct {
	PrayerTitleID uint64
	MemberID      uint64
}

// GetPrayerDetailsResponse represents the prayer details response
type GetPrayerDetailsResponse struct {
	*domain.PrayerTitle
	Completions []*domain.PrayerCompletion `json:"completions,omitempty"`
}

// GetPrayerDetailsUseCase handles getting prayer details
type GetPrayerDetailsUseCase struct {
	prayerService *domain.Service
}

// NewGetPrayerDetailsUseCase creates a new get prayer details use case
func NewGetPrayerDetailsUseCase(prayerService *domain.Service) *GetPrayerDetailsUseCase {
	return &GetPrayerDetailsUseCase{
		prayerService: prayerService,
	}
}

// Execute gets prayer details with contents
func (u *GetPrayerDetailsUseCase) Execute(ctx context.Context, req *GetPrayerDetailsRequest) (*GetPrayerDetailsResponse, error) {
	// Get prayer title first to get room ID
	prayerTitle, err := u.prayerService.GetPrayerTitle(ctx, req.PrayerTitleID)
	if err != nil {
		return nil, err
	}

	// Validate member exists in room (matching Java: validateMemberExistInRoomByTitleId)
	if err := u.prayerService.ValidateRoomAccess(ctx, prayerTitle.RoomID, req.MemberID); err != nil {
		return nil, err
	}

	// Now get prayer title with contents
	prayerTitle, err = u.prayerService.GetPrayerTitleWithContents(ctx, req.PrayerTitleID)
	if err != nil {
		return nil, err
	}

	// Get completions
	completions, err := u.prayerService.GetPrayerCompletions(ctx, req.PrayerTitleID)
	if err != nil {
		// Don't fail if we can't get completions
		completions = []*domain.PrayerCompletion{}
	}

	return &GetPrayerDetailsResponse{
		PrayerTitle: prayerTitle,
		Completions: completions,
	}, nil
}
