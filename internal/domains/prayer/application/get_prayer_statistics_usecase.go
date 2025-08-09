package application

import (
	"context"
	"pray-together/internal/domains/prayer/domain"
)

// GetPrayerStatisticsRequest represents the request to get prayer statistics
type GetPrayerStatisticsRequest struct {
	MemberID uint64
	RoomID   uint64
}

// GetPrayerStatisticsResponse represents prayer statistics response
type GetPrayerStatisticsResponse struct {
	TotalPrayers     int     `json:"totalPrayers"`
	CompletedPrayers int     `json:"completedPrayers"`
	PendingPrayers   int     `json:"pendingPrayers"`
	CompletionRate   float64 `json:"completionRate"`
	MemberPrayers    int     `json:"memberPrayers,omitempty"`
	RoomPrayers      int     `json:"roomPrayers,omitempty"`
}

// GetPrayerStatisticsUseCase handles getting prayer statistics
type GetPrayerStatisticsUseCase struct {
	prayerService *domain.Service
}

// NewGetPrayerStatisticsUseCase creates a new get prayer statistics use case
func NewGetPrayerStatisticsUseCase(prayerService *domain.Service) *GetPrayerStatisticsUseCase {
	return &GetPrayerStatisticsUseCase{
		prayerService: prayerService,
	}
}

// Execute gets prayer statistics
func (u *GetPrayerStatisticsUseCase) Execute(ctx context.Context, req *GetPrayerStatisticsRequest) (*GetPrayerStatisticsResponse, error) {
	stats := &GetPrayerStatisticsResponse{}

	if req.RoomID > 0 {
		// Get room-specific statistics
		titles, err := u.prayerService.Repository.FindPrayerTitlesByRoomID(ctx, req.RoomID, 0, 1000)
		if err != nil {
			return nil, err
		}

		stats.RoomPrayers = len(titles)
		stats.TotalPrayers = len(titles)

		// Count completed prayers
		for _, title := range titles {
			completions, err := u.prayerService.Repository.FindPrayerCompletionsByTitleID(ctx, title.ID)
			if err != nil {
				continue
			}

			// Check if this member completed this prayer
			for _, completion := range completions {
				if req.MemberID > 0 && completion.MemberID == req.MemberID {
					stats.CompletedPrayers++
					break
				}
			}
		}
	} else if req.MemberID > 0 {
		// Get member-specific statistics
		// Count prayers created by member
		count, err := u.prayerService.Repository.CountByMemberID(ctx, req.MemberID)
		if err == nil {
			stats.MemberPrayers = count
			stats.TotalPrayers = count
		}

		// Count answered prayers
		answeredCount, err := u.prayerService.Repository.CountAnsweredByMemberID(ctx, req.MemberID)
		if err == nil {
			stats.CompletedPrayers = answeredCount
		}
	}

	// Calculate pending prayers and completion rate
	stats.PendingPrayers = stats.TotalPrayers - stats.CompletedPrayers
	if stats.TotalPrayers > 0 {
		stats.CompletionRate = float64(stats.CompletedPrayers) / float64(stats.TotalPrayers) * 100
	}

	return stats, nil
}
