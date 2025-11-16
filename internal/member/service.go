package member

import (
	"context"
	"errors"
	"fmt"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/database"
	"gorm.io/gorm"
)

type MemberService struct {
	db               *gorm.DB
	memberRepository *MemberRepository
}

func NewMemberService(db *gorm.DB, memberRepository *MemberRepository) *MemberService {
	return &MemberService{
		db:               db,
		memberRepository: memberRepository,
	}
}

func (s *MemberService) GetProfile(ctx context.Context, memberID int64) (*GetProfileResponse, error) {
	var response *GetProfileResponse

	err := database.WithTransaction(ctx, s.db, func(tx *gorm.DB) error {
		member, err := s.memberRepository.FindByID(ctx, tx, memberID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("회원을 찾을 수 없습니다: %w", ErrMemberNotFound)
			}
			return fmt.Errorf("회원 조회 실패: %w", err)
		}

		response = &GetProfileResponse{
			ID:          member.ID,
			Name:        member.Name,
			Email:       member.Email,
			PhoneNumber: member.PhoneNumber,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

// ValidateMemberExists checks if a member exists by ID
// Returns ErrMemberNotFound if the member does not exist
func (s *MemberService) ValidateMemberExists(ctx context.Context, tx *gorm.DB, memberID int64) error {
	_, err := s.memberRepository.IsExistByID(ctx, tx, memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMemberNotFound
		}
		return fmt.Errorf("회원 존재 확인 실패: %w", err)
	}
	return nil
}
