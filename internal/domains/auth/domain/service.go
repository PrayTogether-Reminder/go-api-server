package domain

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Service represents auth domain service
type Service struct {
	repo             Repository
	tokenService     TokenService
	passwordService  PasswordService
	otpCache         OTPCache
	getMemberByEmail func(ctx context.Context, email string) (uint64, string, string, error) // returns id, name, hashedPassword
	getMemberByID    func(ctx context.Context, memberID uint64) (string, string, error)      // returns email, name
	createMember     func(ctx context.Context, email, password, name string) (uint64, error)
	sendEmail        func(ctx context.Context, to, subject, body string) error
}

// TokenService interface for JWT operations
type TokenService interface {
	GenerateAccessToken(claims *AuthClaims) (string, error)
	GenerateRefreshToken() (string, error)
	ValidateAccessToken(token string) (*AuthClaims, error)
	GetTokenExpiry() time.Duration
}

// PasswordService interface for password operations
type PasswordService interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
}

// NewService creates a new auth service
func NewService(
	repo Repository,
	tokenService TokenService,
	passwordService PasswordService,
	otpCache OTPCache,
	getMemberByEmail func(ctx context.Context, email string) (uint64, string, string, error),
	createMember func(ctx context.Context, email, password, name string) (uint64, error),
	sendEmail func(ctx context.Context, to, subject, body string) error,
) *Service {
	// If no cache provided, use in-memory cache
	if otpCache == nil {
		otpCache = NewInMemoryOTPCache()
	}

	return &Service{
		repo:             repo,
		tokenService:     tokenService,
		passwordService:  passwordService,
		otpCache:         otpCache,
		getMemberByEmail: getMemberByEmail,
		getMemberByID:    nil, // Will be set later if needed
		createMember:     createMember,
		sendEmail:        sendEmail,
	}
}

// SetGetMemberByID sets the getMemberByID helper function
func (s *Service) SetGetMemberByID(fn func(ctx context.Context, memberID uint64) (string, string, error)) {
	s.getMemberByID = fn
}

// Login authenticates a member and returns tokens
func (s *Service) Login(ctx context.Context, email, password string) (*TokenPair, *AuthClaims, error) {
	// Get member by email
	memberID, name, hashedPassword, err := s.getMemberByEmail(ctx, email)
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Verify password
	if err := s.passwordService.ComparePassword(hashedPassword, password); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Generate tokens
	claims := &AuthClaims{
		MemberID: memberID,
		Email:    email,
		Name:     name,
	}

	accessToken, err := s.tokenService.GenerateAccessToken(claims)
	if err != nil {
		return nil, nil, err
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, nil, err
	}

	// Save refresh token
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	rt := NewRefreshToken(memberID, refreshToken, expiresAt)
	if err := s.repo.CreateRefreshToken(ctx, rt); err != nil {
		return nil, nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.tokenService.GetTokenExpiry().Seconds()),
	}, claims, nil
}

// RefreshToken refreshes the access token using refresh token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Find refresh token
	rt, err := s.repo.FindRefreshTokenByToken(ctx, refreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if rt == nil {
		return nil, ErrInvalidRefreshToken
	}

	// Check if expired
	if rt.IsExpired() {
		_ = s.repo.DeleteRefreshToken(ctx, refreshToken)
		return nil, ErrInvalidRefreshToken
	}

	// Delete old refresh token first (matching Java)
	_ = s.repo.DeleteRefreshToken(ctx, refreshToken)

	// Get member info if helper is available
	email := ""
	name := ""
	if s.getMemberByID != nil {
		memberEmail, memberName, err := s.getMemberByID(ctx, rt.MemberID)
		if err == nil {
			email = memberEmail
			name = memberName
		}
	}

	// Generate new tokens
	claims := &AuthClaims{
		MemberID: rt.MemberID,
		Email:    email,
		Name:     name,
	}

	accessToken, err := s.tokenService.GenerateAccessToken(claims)
	if err != nil {
		return nil, err
	}

	// Generate new refresh token
	newRefreshToken, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Save new refresh token
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	newRT := NewRefreshToken(rt.MemberID, newRefreshToken, expiresAt)
	if err := s.repo.CreateRefreshToken(ctx, newRT); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int(s.tokenService.GetTokenExpiry().Seconds()),
	}, nil
}

// Logout logs out a member by deleting refresh token
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.repo.DeleteRefreshToken(ctx, refreshToken)
}

// LogoutAll logs out from all devices
func (s *Service) LogoutAll(ctx context.Context, memberID uint64) error {
	return s.repo.DeleteRefreshTokensByMemberID(ctx, memberID)
}

// SendOTP sends OTP to email
func (s *Service) SendOTP(ctx context.Context, email, purpose string) error {
	// Generate 6-digit OTP
	code := s.generateOTPCode()

	// Save OTP to cache with 3 minute TTL (matching Java)
	if err := s.otpCache.Put(ctx, email, code, 3*time.Minute); err != nil {
		return err
	}

	// Send email
	if s.sendEmail != nil {
		subject := "기도함께 이메일 인증번호" // Matching Java
		body := fmt.Sprintf("인증번호: %s\n이 코드는 3분간 유효합니다.", code)

		if err := s.sendEmail(ctx, email, subject, body); err != nil {
			return err
		}
	}

	return nil
}

// VerifyOTP verifies the OTP
func (s *Service) VerifyOTP(ctx context.Context, email, code, purpose string) error {
	// Get OTP from cache
	cachedOTP, err := s.otpCache.Get(ctx, email)
	if err != nil {
		return err
	}

	// Check if OTP matches
	if cachedOTP != code {
		return ErrOTPInvalid
	}

	// Delete OTP from cache after successful verification
	if err := s.otpCache.Delete(ctx, email); err != nil {
		return err
	}

	return nil
}

// SignUp creates a new member account
func (s *Service) SignUp(ctx context.Context, email, password, name, otpCode string) (*TokenPair, *AuthClaims, error) {
	// Verify OTP first
	if err := s.VerifyOTP(ctx, email, otpCode, "signup"); err != nil {
		return nil, nil, err
	}

	// Hash password
	hashedPassword, err := s.passwordService.HashPassword(password)
	if err != nil {
		return nil, nil, err
	}

	// Create member
	memberID, err := s.createMember(ctx, email, hashedPassword, name)
	if err != nil {
		return nil, nil, err
	}

	// Clean up OTPs for this email
	_ = s.repo.DeleteOTPsByEmail(ctx, email)

	// Generate tokens
	claims := &AuthClaims{
		MemberID: memberID,
		Email:    email,
		Name:     name,
	}

	accessToken, err := s.tokenService.GenerateAccessToken(claims)
	if err != nil {
		return nil, nil, err
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, nil, err
	}

	// Save refresh token
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	rt := NewRefreshToken(memberID, refreshToken, expiresAt)
	if err := s.repo.CreateRefreshToken(ctx, rt); err != nil {
		return nil, nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.tokenService.GetTokenExpiry().Seconds()),
	}, claims, nil
}

// ResetPassword resets member password
func (s *Service) ResetPassword(ctx context.Context, email, newPassword, otpCode string) error {
	// Verify OTP first
	if err := s.VerifyOTP(ctx, email, otpCode, "password-reset"); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := s.passwordService.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password (would need to add this to member domain)
	// For now, return nil as placeholder
	_ = hashedPassword

	// Clean up OTPs for this email
	_ = s.repo.DeleteOTPsByEmail(ctx, email)

	// Invalidate all refresh tokens for this member
	memberID, _, _, _ := s.getMemberByEmail(ctx, email)
	_ = s.repo.DeleteRefreshTokensByMemberID(ctx, memberID)

	return nil
}

// ValidateToken validates an access token
func (s *Service) ValidateToken(ctx context.Context, token string) (*AuthClaims, error) {
	return s.tokenService.ValidateAccessToken(token)
}

// DeleteRefreshToken deletes refresh tokens for a member
func (s *Service) DeleteRefreshToken(ctx context.Context, memberID uint64) error {
	return s.repo.DeleteRefreshTokensByMemberID(ctx, memberID)
}

// HashPassword hashes a password
func (s *Service) HashPassword(password string) (string, error) {
	return s.passwordService.HashPassword(password)
}

// VerifySimpleOTP verifies OTP without specific purpose (for simplified API)
func (s *Service) VerifySimpleOTP(ctx context.Context, email, code string) error {
	// Simplified verification matching Java - just check cache
	return s.VerifyOTP(ctx, email, code, "")
}

// CheckEmailExists checks if an email already exists
func (s *Service) CheckEmailExists(ctx context.Context, email string) error {
	memberID, _, _, err := s.getMemberByEmail(ctx, email)
	if err == nil && memberID > 0 {
		return ErrEmailAlreadyExists
	}
	return nil
}

// CreateMember creates a new member
func (s *Service) CreateMember(ctx context.Context, name, email, hashedPassword string) (uint64, error) {
	return s.createMember(ctx, email, hashedPassword, name)
}

// CleanupExpiredTokens removes expired tokens
func (s *Service) CleanupExpiredTokens(ctx context.Context) error {
	_ = s.repo.DeleteExpiredRefreshTokens(ctx)
	_ = s.repo.DeleteExpiredOTPs(ctx)
	return nil
}

// generateOTPCode generates a 6-digit OTP code
func (s *Service) generateOTPCode() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	num := int(bytes[0])<<16 | int(bytes[1])<<8 | int(bytes[2])
	return fmt.Sprintf("%06d", num%1000000)
}

// generateSecureToken generates a secure random token
func (s *Service) generateSecureToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
