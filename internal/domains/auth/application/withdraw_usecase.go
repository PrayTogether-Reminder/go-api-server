package application

import (
	"context"

	memberDomain "pray-together/internal/domains/member/domain"
)

// WithdrawRequest represents the withdrawal request
type WithdrawRequest struct {
	MemberID uint64
}

// WithdrawUseCase handles member withdrawal
type WithdrawUseCase struct {
	memberService *memberDomain.Service
}

// NewWithdrawUseCase creates a new WithdrawUseCase
func NewWithdrawUseCase(memberService *memberDomain.Service) *WithdrawUseCase {
	return &WithdrawUseCase{
		memberService: memberService,
	}
}

// Execute performs member withdrawal
func (uc *WithdrawUseCase) Execute(ctx context.Context, req *WithdrawRequest) error {
	// Delete member (soft delete)
	return uc.memberService.DeleteMember(ctx, req.MemberID)
}
