package application

import (
	"context"
	"pray-together/internal/domains/prayer/domain"
)

// UpdatePrayerRequest represents the request to update a prayer
type UpdatePrayerRequest struct {
	PrayerID uint64
	MemberID uint64
	Content  string
}

// UpdatePrayerResponse represents the response after updating a prayer
type UpdatePrayerResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UpdatePrayerUseCase handles prayer update
type UpdatePrayerUseCase struct {
	prayerService *domain.Service
}

// NewUpdatePrayerUseCase creates a new UpdatePrayerUseCase
func NewUpdatePrayerUseCase(prayerService *domain.Service) *UpdatePrayerUseCase {
	return &UpdatePrayerUseCase{
		prayerService: prayerService,
	}
}

// Execute updates a prayer
func (uc *UpdatePrayerUseCase) Execute(ctx context.Context, req *UpdatePrayerRequest) (*UpdatePrayerResponse, error) {
	_, err := uc.prayerService.UpdatePrayer(
		ctx,
		req.PrayerID,
		req.MemberID,
		req.Content,
	)
	if err != nil {
		return nil, err
	}

	return &UpdatePrayerResponse{
		Success: true,
		Message: "Prayer updated successfully",
	}, nil
}
