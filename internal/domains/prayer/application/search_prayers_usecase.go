package application

import (
	"context"
	"pray-together/internal/domains/prayer/domain"
	"strings"
)

// SearchPrayersRequest represents the request to search prayers
type SearchPrayersRequest struct {
	MemberID   uint64
	RoomID     uint64
	Query      string
	IsAnswered *bool
	Limit      int
	Offset     int
}

// SearchPrayersResponse represents search prayers response
type SearchPrayersResponse struct {
	Prayers []*domain.PrayerTitle `json:"prayers"`
	Total   int                   `json:"total"`
	HasMore bool                  `json:"hasMore"`
}

// SearchPrayersUseCase handles searching prayers
type SearchPrayersUseCase struct {
	prayerService *domain.Service
}

// NewSearchPrayersUseCase creates a new search prayers use case
func NewSearchPrayersUseCase(prayerService *domain.Service) *SearchPrayersUseCase {
	return &SearchPrayersUseCase{
		prayerService: prayerService,
	}
}

// Execute searches for prayers
func (u *SearchPrayersUseCase) Execute(ctx context.Context, req *SearchPrayersRequest) (*SearchPrayersResponse, error) {
	// Default limit
	if req.Limit == 0 {
		req.Limit = 20
	}

	// Get all prayers for the room
	prayers, err := u.prayerService.Repository.FindPrayerTitlesByRoomID(ctx, req.RoomID, 0, 1000)
	if err != nil {
		return nil, err
	}

	// Filter prayers based on search criteria
	var filtered []*domain.PrayerTitle
	for _, prayer := range prayers {
		// Search in title and description
		if req.Query != "" {
			query := strings.ToLower(req.Query)
			title := strings.ToLower(prayer.Title)
			description := strings.ToLower(prayer.Description)

			if !strings.Contains(title, query) && !strings.Contains(description, query) {
				// Also search in contents
				contents, err := u.prayerService.Repository.FindPrayerContentsByTitleID(ctx, prayer.ID)
				if err == nil {
					found := false
					for _, content := range contents {
						if strings.Contains(strings.ToLower(content.Content), query) {
							found = true
							break
						}
					}
					if !found {
						continue
					}
				} else {
					continue
				}
			}
		}

		// Filter by answered status
		if req.IsAnswered != nil {
			if *req.IsAnswered != prayer.IsAnswered {
				continue
			}
		}

		filtered = append(filtered, prayer)
	}

	// Apply pagination
	total := len(filtered)
	start := req.Offset
	end := req.Offset + req.Limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginated := filtered[start:end]

	return &SearchPrayersResponse{
		Prayers: paginated,
		Total:   total,
		HasMore: end < total,
	}, nil
}
