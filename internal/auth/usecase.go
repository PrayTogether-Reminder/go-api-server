package auth

import (
	"context"
	"fmt"
	"strconv"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/auth/otp"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/member"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/database"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/logger"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/token"
	"gorm.io/gorm"
)

// AuthUseCase orchestrates auth-specific application logic.
type AuthUseCase struct {
	db            *gorm.DB
	memberService *member.MemberService
	tokenManager  token.Manager
	authService   *AuthService
	otpService    *otp.Service
}

func NewAuthUseCase(db *gorm.DB, memberService *member.MemberService, tokenManager token.Manager, authService *AuthService, otpService *otp.Service) *AuthUseCase {
	return &AuthUseCase{
		db:            db,
		memberService: memberService,
		tokenManager:  tokenManager,
		authService:   authService,
		otpService:    otpService,
	}
}

func (u *AuthUseCase) Login(ctx context.Context, request *LoginRequest) (*LoginResponse, error) {
	log := logger.FromContext(ctx)

	foundMember, err := u.memberService.GetByEmail(ctx, u.db, request.Email)
	if err != nil {
		return nil, err
	}

	if err := u.authService.ComparePassword(foundMember.Password, request.Password); err != nil {
		return nil, err
	}

	memberID := strconv.FormatInt(foundMember.ID, 10)
	accessToken, err := u.tokenManager.GenerateAccessToken(memberID, foundMember.Email) // todo : token manager err
	if err != nil {
		return nil, err
	}

	refreshToken, err := u.tokenManager.GenerateRefreshToken(memberID, foundMember.Email)
	if err != nil {
		return nil, err
	}

	log.Info("로그인 성공", "email", logger.MaskEmail(request.Email))

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *AuthUseCase) Signup(ctx context.Context, request *SignupRequest) error {
	log := logger.FromContext(ctx)

	return database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		hashedPassword, err := u.authService.HashPassword(request.Password)
		if err != nil {
			return err
		}

		newMember := model.NewMember(request.Name, request.Email, request.PhoneNumber, hashedPassword)
		if err := u.memberService.CreateMember(ctx, tx, newMember); err != nil {
			return err
		}

		log.Info("Member created successfully", "email", logger.MaskEmail(request.Email))
		return nil
	})
}

func (u *AuthUseCase) RequestEmailOTP(ctx context.Context, request *EmailOtpRequest) (*sharedHttp.MessageResponse, error) {
	exists, err := u.memberService.ExistsByEmail(ctx, u.db, request.Email)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, fmt.Errorf("이미 존재하는 이메일입니다: %w", member.ErrMemberAlreadyExists)
	}

	if err := u.otpService.SendToEmail(ctx, request.Email); err != nil {
		return nil, err
	}

	return &sharedHttp.MessageResponse{Message: "인증 번호를 요청했습니다."}, nil
}
