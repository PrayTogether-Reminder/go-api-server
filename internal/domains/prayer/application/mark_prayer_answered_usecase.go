package application

import (
	"context"
	"pray-together/internal/domains/prayer/domain"
)

// MarkPrayerAnsweredRequest represents the request to mark a prayer as answered
type MarkPrayerAnsweredRequest struct {
	PrayerID uint64
	MemberID uint64
}

// MarkPrayerAnsweredResponse represents the response after marking prayer as answered
type MarkPrayerAnsweredResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// MarkPrayerAnsweredUseCase handles marking prayer as answered
type MarkPrayerAnsweredUseCase struct {
	prayerService *domain.Service
}

// NewMarkPrayerAnsweredUseCase creates a new MarkPrayerAnsweredUseCase
func NewMarkPrayerAnsweredUseCase(prayerService *domain.Service) *MarkPrayerAnsweredUseCase {
	return &MarkPrayerAnsweredUseCase{
		prayerService: prayerService,
	}
}

// Execute marks a prayer as answered
func (uc *MarkPrayerAnsweredUseCase) Execute(ctx context.Context, req *MarkPrayerAnsweredRequest) (*MarkPrayerAnsweredResponse, error) {
	err := uc.prayerService.MarkPrayerAsAnswered(
		ctx,
		req.PrayerID,
		req.MemberID,
	)
	if err != nil {
		return nil, err
	}

	return &MarkPrayerAnsweredResponse{
		Success: true,
		Message: "Prayer marked as answered",
	}, nil
}
