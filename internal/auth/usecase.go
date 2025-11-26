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
	db                  *gorm.DB
	memberService       *member.MemberService
	tokenManager        token.Manager
	authService         *AuthService
	otpService          *otp.Service
	refreshTokenService *RefreshTokenService
}

func NewAuthUseCase(db *gorm.DB, memberService *member.MemberService, tokenManager token.Manager, authService *AuthService, otpService *otp.Service, refreshTokenService *RefreshTokenService) *AuthUseCase {
	return &AuthUseCase{
		db:                  db,
		memberService:       memberService,
		tokenManager:        tokenManager,
		authService:         authService,
		otpService:          otpService,
		refreshTokenService: refreshTokenService,
	}
}

func (u *AuthUseCase) Login(ctx context.Context, request *LoginRequest) (*LoginResponse, error) {
	foundMember, err := u.memberService.GetByEmail(ctx, u.db, request.Email)
	if err != nil {
		return nil, err
	}

	if err := u.authService.ComparePassword(foundMember.Password, request.Password); err != nil {
		return nil, err
	}

	memberID := strconv.FormatInt(foundMember.ID, 10)
	accessToken, err := u.tokenManager.GenerateAccessToken(memberID, foundMember.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := u.tokenManager.GenerateRefreshToken(memberID, foundMember.Email)
	if err != nil {
		return nil, err
	}

	// Refresh Token의 만료 시각 추출 및 저장
	expiresAt, err := u.tokenManager.ExtractExpiration(refreshToken)
	if err != nil {
		return nil, err
	}

	if err := u.refreshTokenService.Save(ctx, u.db, foundMember.ID, refreshToken, expiresAt); err != nil {
		return nil, fmt.Errorf("RefreshToken 저장 실패: %w", err)
	}

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
	log := logger.FromContext(ctx)

	exists, err := u.memberService.ExistsByEmail(ctx, u.db, request.Email)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, fmt.Errorf("이미 존재하는 이메일입니다: %w", member.ErrMemberAlreadyExists)
	}

	if err := u.otpService.SendToEmail(ctx, request.Email); err != nil {
		log.Error("OTP 발송 실패", "email", logger.MaskEmail(request.Email), "error", err)
		return nil, err
	}

	log.Info("OTP 발송 성공", "email", logger.MaskEmail(request.Email))
	return &sharedHttp.MessageResponse{Message: "인증 번호를 요청했습니다."}, nil
}

func (u *AuthUseCase) VerifyEmailOTP(ctx context.Context, request *VerifyOtpRequest) (*sharedHttp.MessageResponse, error) {
	log := logger.FromContext(ctx)

	isValid, err := u.otpService.VerifyOTP(ctx, request.Email, request.Otp)
	if err != nil {
		log.Error("OTP 검증 실패", "email", logger.MaskEmail(request.Email), "error", err)
		return nil, err
	}

	if !isValid {
		log.Warn("OTP 불일치", "email", logger.MaskEmail(request.Email))
		return &sharedHttp.MessageResponse{Message: "인증 번호가 일치하지 않습니다."}, nil
	}

	log.Info("OTP 검증 성공", "email", logger.MaskEmail(request.Email))
	return &sharedHttp.MessageResponse{Message: "인증에 성공했습니다."}, nil
}

func (u *AuthUseCase) Withdraw(ctx context.Context, memberID int64) (*sharedHttp.MessageResponse, error) {
	log := logger.FromContext(ctx)

	err := database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		if err := u.memberService.DeleteMember(ctx, tx, memberID); err != nil {
			log.Error("회원 탈퇴 실패", "memberID", memberID, "error", err)
			return err
		}

		log.Info("회원 탈퇴 성공", "memberID", memberID)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &sharedHttp.MessageResponse{Message: "회원 탈퇴를 완료했습니다.\n 함께 기도해 주셔 감사합니다."}, nil
}

// ReissueAuthToken - Refresh Token을 사용하여 Access Token과 Refresh Token 재발급
func (u *AuthUseCase) ReissueAuthToken(ctx context.Context, request *AuthTokenReissueRequest) (*AuthTokenReissueResponse, error) {
	var newAccessToken string
	var newRefreshToken string

	err := database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		// 1. Refresh Token에서 memberID 추출
		memberID, err := u.tokenManager.ExtractMemberID(request.RefreshToken)
		if err != nil {
			return err
		}

		// 2. Member 존재 확인
		foundMember, err := u.memberService.GetByID(ctx, tx, memberID)
		if err != nil {
			return err
		}

		// 3. DB에 저장된 RefreshToken 검증
		if err := u.refreshTokenService.ValidateExist(ctx, tx, memberID, request.RefreshToken); err != nil {
			return err
		}

		// 4. JWT 토큰 검증
		if _, err := u.tokenManager.ValidateToken(request.RefreshToken); err != nil {
			return err
		}

		// 5. 새로운 토큰 생성
		memberIDStr := strconv.FormatInt(memberID, 10)
		generatedAccess, err := u.tokenManager.GenerateAccessToken(memberIDStr, foundMember.Email)
		if err != nil {
			return err
		}
		newAccessToken = generatedAccess

		generatedRefresh, err := u.tokenManager.GenerateRefreshToken(memberIDStr, foundMember.Email)
		if err != nil {
			return err
		}
		newRefreshToken = generatedRefresh

		expiresAt, err := u.tokenManager.ExtractExpiration(generatedRefresh)
		if err != nil {
			return err
		}

		// 6. 기존 토큰 삭제 및 새 토큰 저장
		if err := u.refreshTokenService.Delete(ctx, tx, memberID); err != nil {
			return err
		}

		if err := u.refreshTokenService.Save(ctx, tx, memberID, generatedRefresh, expiresAt); err != nil {
			return fmt.Errorf("RefreshToken 저장 실패: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &AuthTokenReissueResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
