package application

import (
	"context"
	"pray-together/internal/domains/prayer/domain"
)

// CreatePrayerRequest represents the request to create a prayer
type CreatePrayerRequest struct {
	MemberID uint64
	RoomID   uint64
	Content  string
	Type     domain.PrayerType
}

// CreatePrayerResponse represents the response after creating a prayer
type CreatePrayerResponse struct {
	ID        uint64            `json:"id"`
	MemberID  uint64            `json:"memberId"`
	RoomID    uint64            `json:"roomId"`
	Content   string            `json:"content"`
	Type      domain.PrayerType `json:"type"`
	CreatedAt string            `json:"createdAt"`
}

// CreatePrayerUseCase handles prayer creation
type CreatePrayerUseCase struct {
	prayerService *domain.Service
}

// NewCreatePrayerUseCase creates a new CreatePrayerUseCase
func NewCreatePrayerUseCase(prayerService *domain.Service) *CreatePrayerUseCase {
	return &CreatePrayerUseCase{
		prayerService: prayerService,
	}
}

// Execute creates a new prayer
func (uc *CreatePrayerUseCase) Execute(ctx context.Context, req *CreatePrayerRequest) (*CreatePrayerResponse, error) {
	prayer, err := uc.prayerService.CreatePrayer(
		ctx,
		req.MemberID,
		req.RoomID,
		req.Content,
		req.Type,
	)
	if err != nil {
		return nil, err
	}

	return &CreatePrayerResponse{
		ID:        prayer.ID,
		MemberID:  prayer.MemberID,
		RoomID:    prayer.RoomID,
		Content:   prayer.Content,
		Type:      prayer.Type,
		CreatedAt: prayer.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
