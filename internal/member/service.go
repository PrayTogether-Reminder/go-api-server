package member

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// MemberService - Member 도메인 서비스
// 도메인 로직 + Repository 호출 + 에러 변환 담당
type MemberService struct {
	memberRepository *MemberRepository
}

// NewMemberService - MemberService 생성자
func NewMemberService(memberRepository *MemberRepository) *MemberService {
	return &MemberService{
		memberRepository: memberRepository,
	}
}

// GetByID - ID로 회원 조회
func (s *MemberService) GetByID(ctx context.Context, db *gorm.DB, memberID int64) (*Member, error) {
	member, err := s.memberRepository.FindByID(ctx, db, memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("회원을 찾을 수 없습니다: %w", ErrMemberNotFound)
		}
		return nil, fmt.Errorf("회원 조회 실패: %w", err)
	}

	return member, nil
}

// GetByEmail - 이메일로 회원 조회
func (s *MemberService) GetByEmail(ctx context.Context, db *gorm.DB, email string) (*Member, error) {
	member, err := s.memberRepository.FindByEmail(ctx, db, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("회원을 찾을 수 없습니다: %w", ErrMemberNotFound)
		}
		return nil, fmt.Errorf("회원 조회 실패: %w", err)
	}

	return member, nil
}

// ExistsByEmail - 이메일 중복 확인
func (s *MemberService) ExistsByEmail(ctx context.Context, db *gorm.DB, email string) (bool, error) {
	exists, err := s.memberRepository.IsExistByEmail(ctx, db, email)
	if err != nil {
		return false, fmt.Errorf("이메일 중복 확인 실패: %w", err)
	}

	return exists, nil
}

// ValidateMemberExists - 회원 존재 여부 검증 (도메인 규칙)
// 회원이 존재하지 않으면 ErrMemberNotFound 반환
func (s *MemberService) ValidateMemberExists(ctx context.Context, db *gorm.DB, memberID int64) error {
	exists, err := s.memberRepository.IsExistByID(ctx, db, memberID)
	if err != nil {
		return fmt.Errorf("회원 존재 확인 실패: %w", err)
	}

	if !exists {
		return ErrMemberNotFound
	}

	return nil
}

// CreateMember - 회원 생성
func (s *MemberService) CreateMember(ctx context.Context, db *gorm.DB, member *Member) error {
	// 이메일 중복 확인 (도메인 규칙)
	exists, err := s.ExistsByEmail(ctx, db, member.Email)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("이미 존재하는 이메일입니다: %w", ErrMemberAlreadyExists)
	}

	// 회원 생성
	if err := s.memberRepository.Create(ctx, db, member); err != nil {
		return fmt.Errorf("회원 생성 실패: %w", err)
	}

	return nil
}
