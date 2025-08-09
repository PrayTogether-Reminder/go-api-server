package application

import (
	"context"
	"pray-together/internal/domains/member/domain"
)

// DeleteMemberRequest represents the request to delete member
type DeleteMemberRequest struct {
	MemberID uint64
	Password string
}

// DeleteMemberResponse represents the response after deleting member
type DeleteMemberResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// DeleteMemberUseCase handles deleting member account
type DeleteMemberUseCase struct {
	memberService    *domain.Service
	validatePassword func(hashedPassword, password string) error
}

// NewDeleteMemberUseCase creates a new DeleteMemberUseCase
func NewDeleteMemberUseCase(
	memberService *domain.Service,
	validatePassword func(hashedPassword, password string) error,
) *DeleteMemberUseCase {
	return &DeleteMemberUseCase{
		memberService:    memberService,
		validatePassword: validatePassword,
	}
}

// Execute deletes member account
func (uc *DeleteMemberUseCase) Execute(ctx context.Context, req *DeleteMemberRequest) (*DeleteMemberResponse, error) {
	// Get member
	member, err := uc.memberService.GetMember(ctx, req.MemberID)
	if err != nil {
		return nil, err
	}

	// Validate password
	if uc.validatePassword != nil {
		if err := uc.validatePassword(member.Password, req.Password); err != nil {
			return nil, domain.ErrInvalidPassword
		}
	}

	// Delete member
	if err := uc.memberService.DeleteMember(ctx, req.MemberID); err != nil {
		return nil, err
	}

	return &DeleteMemberResponse{
		Success: true,
		Message: "Member account deleted successfully",
	}, nil
}
