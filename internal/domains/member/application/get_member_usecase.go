package application

import (
	"context"
	"pray-together/internal/domains/member/domain"
)

// GetMemberRequest represents the request to get member info
type GetMemberRequest struct {
	MemberID uint64
}

// GetMemberResponse represents the response with member info
type GetMemberResponse struct {
	ID    uint64 `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// GetMemberUseCase handles getting member information
type GetMemberUseCase struct {
	memberService *domain.Service
}

// NewGetMemberUseCase creates a new GetMemberUseCase
func NewGetMemberUseCase(memberService *domain.Service) *GetMemberUseCase {
	return &GetMemberUseCase{
		memberService: memberService,
	}
}

// Execute gets member information
func (uc *GetMemberUseCase) Execute(ctx context.Context, req *GetMemberRequest) (*GetMemberResponse, error) {
	member, err := uc.memberService.GetMember(ctx, req.MemberID)
	if err != nil {
		return nil, err
	}

	return &GetMemberResponse{
		ID:    member.ID,
		Email: member.Email,
		Name:  member.Name,
	}, nil
}

// GetMemberByEmail gets member information by email
func (uc *GetMemberUseCase) GetMemberByEmail(ctx context.Context, email string) (*domain.MemberInfo, error) {
	member, err := uc.memberService.GetMemberByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return member.ToInfo(), nil
}

// SearchMembersByName searches members by name
func (uc *GetMemberUseCase) SearchMembersByName(ctx context.Context, name string) ([]*domain.MemberInfo, error) {
	members, err := uc.memberService.SearchMembersByName(ctx, name)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.MemberInfo, len(members))
	for i, member := range members {
		result[i] = member.ToInfo()
	}

	return result, nil
}
