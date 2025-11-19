package member

import (
	"context"

	"gorm.io/gorm"
)

// MemberUseCase - Member 애플리케이션 유스케이스
type MemberUseCase struct {
	db            *gorm.DB
	memberService *MemberService
}

func NewMemberUseCase(db *gorm.DB, memberService *MemberService) *MemberUseCase {
	return &MemberUseCase{
		db:            db,
		memberService: memberService,
	}
}

// GetProfile - 내 프로필 조회 유스케이스
func (u *MemberUseCase) GetProfile(ctx context.Context, memberID int64) (*Member, error) {
	return u.memberService.GetByID(ctx, u.db, memberID)
}

// GetMemberByEmail - 이메일로 회원 찾기 유스케이스 (로그인 등에서 사용)
func (u *MemberUseCase) GetMemberByEmail(ctx context.Context, email string) (*Member, error) {
	return u.memberService.GetByEmail(ctx, u.db, email)
}

// Signup - 신규 회원 가입 유스케이스
func (u *MemberUseCase) Signup(ctx context.Context, member *Member) error {
	return u.memberService.CreateMember(ctx, u.db, member)
}
