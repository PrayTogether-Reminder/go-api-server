package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pray-together/internal/domains/auth/application"
	"pray-together/internal/domains/auth/domain"
)

// Handler handles HTTP requests for authentication
type Handler struct {
	signupUseCase       *application.SignupUseCase
	loginUseCase        *application.LoginUseCase
	logoutUseCase       *application.LogoutUseCase
	refreshTokenUseCase *application.RefreshTokenUseCase
	withdrawUseCase     *application.WithdrawUseCase
	sendOTPUseCase      *application.SendOTPUseCase
	verifyOTPUseCase    *application.VerifyOTPUseCase
}

// NewHandler creates a new auth handler
func NewHandler(
	signupUseCase *application.SignupUseCase,
	loginUseCase *application.LoginUseCase,
	logoutUseCase *application.LogoutUseCase,
	refreshTokenUseCase *application.RefreshTokenUseCase,
	withdrawUseCase *application.WithdrawUseCase,
	sendOTPUseCase *application.SendOTPUseCase,
	verifyOTPUseCase *application.VerifyOTPUseCase,
) *Handler {
	return &Handler{
		signupUseCase:       signupUseCase,
		loginUseCase:        loginUseCase,
		logoutUseCase:       logoutUseCase,
		refreshTokenUseCase: refreshTokenUseCase,
		withdrawUseCase:     withdrawUseCase,
		sendOTPUseCase:      sendOTPUseCase,
		verifyOTPUseCase:    verifyOTPUseCase,
	}
}

// SignupRequest represents signup request
type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// RefreshTokenResponse represents refresh token response
type RefreshTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// SendOTPRequest represents send OTP request
type SendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyOTPRequest represents verify OTP request
type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

// Signup handles member signup
func (h *Handler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	signupReq := &application.SignupRequest{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	if err := h.signupUseCase.Execute(c.Request.Context(), signupReq); err != nil {
		if err == domain.ErrEmailAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to signup"})
		return
	}

	c.JSON(http.StatusCreated, MessageResponse{Message: "회원 가입을 완료했습니다."})
}

// Login handles member login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loginReq := &application.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.loginUseCase.Execute(c.Request.Context(), loginReq)
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "이메일 또는 비밀번호가 일치하지 않습니다."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	})
}

// Logout handles member logout
func (h *Handler) Logout(c *gin.Context) {
	// Get member ID from context (set by auth middleware)
	memberIDVal, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	memberID := memberIDVal.(uint64)

	logoutReq := &application.LogoutRequest{
		MemberID: memberID,
	}

	if err := h.logoutUseCase.Execute(c.Request.Context(), logoutReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "로그아웃 되었습니다."})
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refreshReq := &application.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	resp, err := h.refreshTokenUseCase.Execute(c.Request.Context(), refreshReq)
	if err != nil {
		if err == domain.ErrInvalidRefreshToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh token"})
		return
	}

	c.JSON(http.StatusOK, RefreshTokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	})
}

// Withdraw handles member withdrawal
func (h *Handler) Withdraw(c *gin.Context) {
	// Get member ID from context (set by auth middleware)
	memberIDVal, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	memberID := memberIDVal.(uint64)

	withdrawReq := &application.WithdrawRequest{
		MemberID: memberID,
	}

	if err := h.withdrawUseCase.Execute(c.Request.Context(), withdrawReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to withdraw"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "회원 탈퇴를 완료했습니다.\n함께 기도해 주셔 감사합니다."})
}

// SendOTP handles sending OTP
func (h *Handler) SendOTP(c *gin.Context) {
	var req SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	otpReq := &application.SendOTPRequest{
		Email: req.Email,
	}

	if err := h.sendOTPUseCase.Execute(c.Request.Context(), otpReq); err != nil {
		if err == domain.ErrEmailAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "인증 번호를 요청했습니다."})
}

// VerifyOTP handles OTP verification
func (h *Handler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	verifyReq := &application.VerifyOTPRequest{
		Email: req.Email,
		OTP:   req.OTP,
	}

	resp, err := h.verifyOTPUseCase.Execute(c.Request.Context(), verifyReq)
	if err != nil || !resp.IsValid {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "인증 번호가 일치하지 않습니다."})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "인증에 성공했습니다."})
}

// RegisterRoutes registers auth routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/signup", h.Signup)
		// Removed login/logout - not in Java API
		auth.POST("/reissue-token", h.RefreshToken)
		auth.DELETE("/withdraw", h.Withdraw)
		auth.POST("/otp/email", h.SendOTP)
		auth.POST("/otp/email/verification", h.VerifyOTP)
	}
}
