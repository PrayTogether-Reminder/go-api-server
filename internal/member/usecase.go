package member

import (
	"context"
	"strings"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/database"
	sharedDomain "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/domain"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/http"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/logger"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
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
func (u *MemberUseCase) GetProfile(ctx context.Context, memberID int64) (*model.Member, error) {
	return u.memberService.GetByID(ctx, u.db, memberID)
}

// GetMemberByEmail - 이메일로 회원 찾기 유스케이스 (로그인 등에서 사용)
func (u *MemberUseCase) GetMemberByEmail(ctx context.Context, email string) (*model.Member, error) {
	return u.memberService.GetByEmail(ctx, u.db, email)
}

// Signup - 신규 회원 가입 유스케이스
func (u *MemberUseCase) Signup(ctx context.Context, member *model.Member) error {
	return u.memberService.CreateMember(ctx, u.db, member)
}

// UpdateProfile - 내 프로필 수정 유스케이스
func (u *MemberUseCase) UpdateProfile(ctx context.Context, memberID int64, request *UpdateProfileRequest) (*sharedHttp.MessageResponse, error) {
	var normalizedPhone *string
	if request.PhoneNumber != nil {
		phone, err := sharedDomain.NewPhoneNumber(*request.PhoneNumber)
		if err != nil {
			return nil, err
		}
		formatted := phone.String()
		normalizedPhone = &formatted
	}

	if err := database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		return u.memberService.UpdateProfile(ctx, tx, memberID, request.Name, normalizedPhone)
	}); err != nil {
		return nil, err
	}

	return &sharedHttp.MessageResponse{Message: "프로필을 변경했습니다."}, nil
}

// SearchMembers - 이름으로 회원 검색
func (u *MemberUseCase) SearchMembers(ctx context.Context, memberID int64, name string) (*SearchMemberResponse, error) {
	log := logger.FromContext(ctx)
	log.Info("회원 검색 요청", "memberID", memberID, "keyword", name)

	results, err := u.memberService.SearchMembers(ctx, u.db, name)
	if err != nil {
		return nil, err
	}

	members := make([]SearchMemberDTO, 0, len(results))
	for _, result := range results {
		var suffix *string
		if strings.TrimSpace(result.PhoneNumber) != "" {
			phone, err := sharedDomain.NewPhoneNumber(result.PhoneNumber)
			if err != nil {
				return nil, err
			}
			s := phone.GetSuffix()
			suffix = &s
		}

		members = append(members, SearchMemberDTO{
			ID:                result.ID,
			Name:              result.Name,
			PhoneNumberSuffix: suffix,
		})
	}

	return &SearchMemberResponse{Members: members}, nil
}
