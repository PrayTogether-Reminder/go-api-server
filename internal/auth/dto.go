package auth

type SignupRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=20"`
	Email       string `json:"email" binding:"required,email,max=50"`
	PhoneNumber string `json:"phoneNumber" binding:"required,phone"`
	Password    string `json:"password" binding:"required,min=8,max=15"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=15"`
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type EmailOtpRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyOtpRequest struct {
	Email string `json:"email" binding:"required,email"`
	Otp   string `json:"otp" binding:"required"`
}

// AuthTokenReissueRequest - 토큰 재발급 요청 DTO
type AuthTokenReissueRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// AuthTokenReissueResponse - 토큰 재발급 응답 DTO
type AuthTokenReissueResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
