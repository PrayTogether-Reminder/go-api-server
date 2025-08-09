package application

import (
	"context"
	"pray-together/internal/domains/prayer/domain"
)

// ListPrayerTitlesRequest represents the request to list prayer titles
type ListPrayerTitlesRequest struct {
	RoomID   uint64
	MemberID uint64
	After    string // For infinite scroll
}

// ListPrayerTitlesUseCase handles listing prayer titles
type ListPrayerTitlesUseCase struct {
	prayerService *domain.Service
}

// NewListPrayerTitlesUseCase creates a new list prayer titles use case
func NewListPrayerTitlesUseCase(prayerService *domain.Service) *ListPrayerTitlesUseCase {
	return &ListPrayerTitlesUseCase{
		prayerService: prayerService,
	}
}

// Execute lists prayer titles for a room with infinite scroll
func (u *ListPrayerTitlesUseCase) Execute(ctx context.Context, req *ListPrayerTitlesRequest) ([]*domain.PrayerTitleInfo, error) {
	// Validate member exists in room (matching Java: memberRoomService.validateMemberExistInRoom)
	if err := u.prayerService.ValidateRoomAccess(ctx, req.RoomID, req.MemberID); err != nil {
		return nil, err
	}

	// Get titles with time-based cursor (matching Java: titleService.fetchTitlesByRoom)
	// Java uses 10 items per page (PRAYER_TITLES_INFINITE_SCROLL_SIZE = 10)
	return u.prayerService.GetRoomPrayerTitleInfos(ctx, req.RoomID, req.After, 10)
}
