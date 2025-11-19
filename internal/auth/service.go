package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"strconv"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/member"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/database"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/logger"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/token"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db            *gorm.DB
	memberService *member.MemberService
	tokenManager  token.Manager
}

func NewAuthService(db *gorm.DB, memberService *member.MemberService, tokenManager token.Manager) *AuthService {
	return &AuthService{
		db:            db,
		memberService: memberService,
		tokenManager:  tokenManager,
	}
}

func (a *AuthService) Login(ctx context.Context, request *LoginRequest) (*LoginResponse, error) {
	log := logger.FromContext(ctx)

	// 1. Find member by email (Domain Service 사용)
	foundMember, err := a.memberService.GetByEmail(ctx, a.db, request.Email)
	if err != nil {
		if errors.Is(err, member.ErrMemberNotFound) {
			return nil, fmt.Errorf("이메일을 찾을 수 없습니다: email=%s %w", logger.MaskEmail(request.Email), ErrInCorrectEmailPassword) // Security: don't reveal if email exists
		}
		return nil, fmt.Errorf("로그인 실패: email=%s %w", logger.MaskEmail(request.Email), err)
	}

	// 2. Validate password
	if err := bcrypt.CompareHashAndPassword([]byte(foundMember.Password), []byte(request.Password)); err != nil {
		return nil, fmt.Errorf("로그인 실패: email=%s %w", logger.MaskEmail(request.Email), ErrInCorrectEmailPassword)
	}

	// 3. Generate JWT tokens
	memberID := strconv.FormatInt(foundMember.ID, 10)
	accessToken, err := a.tokenManager.GenerateAccessToken(memberID, foundMember.Email)
	if err != nil {
		return nil, fmt.Errorf("AccessToken 생성 실패: memberID=%s %w", memberID, err)
	}

	refreshToken, err := a.tokenManager.GenerateRefreshToken(memberID, foundMember.Email)
	if err != nil {
		return nil, fmt.Errorf("RefreshToken 생성 실패: memberID=%s %w", memberID, err)
	}

	log.Info("로그인 성공", "email", logger.MaskEmail(request.Email))

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *AuthService) Signup(ctx context.Context, request *SignupRequest) error {
	log := logger.FromContext(ctx)
	return database.WithTransaction(ctx, a.db, func(tx *gorm.DB) error {
		// 비밀번호 해싱
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("비밀번호 해싱 실패: %w", err)
		}

		// 회원 생성 (Domain Service 사용 - 중복 체크 포함)
		newMember := model.NewMember(request.Name, request.Email, request.PhoneNumber, string(hashedPassword))
		if err := a.memberService.CreateMember(ctx, tx, newMember); err != nil {
			return fmt.Errorf("회원 계정 생성 실패: %w", err)
		}

		log.Info("Member created successfully", "email", logger.MaskEmail(request.Email))
		return nil
	})
}
