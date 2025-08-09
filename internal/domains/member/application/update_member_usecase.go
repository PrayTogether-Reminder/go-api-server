package application

import (
	"context"
	"pray-together/internal/domains/member/domain"
)

// UpdateMemberRequest represents the request to update member
type UpdateMemberRequest struct {
	MemberID uint64
	Name     string
}

// UpdateMemberResponse represents the response after updating member
type UpdateMemberResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UpdateMemberUseCase handles updating member information
type UpdateMemberUseCase struct {
	memberService *domain.Service
}

// NewUpdateMemberUseCase creates a new UpdateMemberUseCase
func NewUpdateMemberUseCase(memberService *domain.Service) *UpdateMemberUseCase {
	return &UpdateMemberUseCase{
		memberService: memberService,
	}
}

// Execute updates member information
func (uc *UpdateMemberUseCase) Execute(ctx context.Context, req *UpdateMemberRequest) (*UpdateMemberResponse, error) {
	// Get member
	member, err := uc.memberService.GetMember(ctx, req.MemberID)
	if err != nil {
		return nil, err
	}

	// Update name
	if err := member.UpdateName(req.Name); err != nil {
		return nil, err
	}

	// Save changes
	if err := uc.memberService.UpdateMember(ctx, member); err != nil {
		return nil, err
	}

	return &UpdateMemberResponse{
		Success: true,
		Message: "Member name updated successfully",
	}, nil
}
