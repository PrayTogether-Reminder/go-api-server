package domain

import (
	"errors"
	"time"
)

// BaseEntity contains common fields for all entities in auth domain
type BaseEntity struct {
	CreatedAt time.Time  `gorm:"column:created_at;not null" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index" json:"deletedAt,omitempty"`
}

// RefreshToken represents a refresh token entity
type RefreshToken struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	MemberID  uint64    `gorm:"column:member_id;not null;index" json:"memberId"`
	Token     string    `gorm:"column:token;uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null" json:"expiresAt"`
	BaseEntity
}

// TableName specifies the table name for RefreshToken
func (RefreshToken) TableName() string {
	return "refresh_token"
}

// NewRefreshToken creates a new refresh token
func NewRefreshToken(memberID uint64, token string, expiresAt time.Time) *RefreshToken {
	return &RefreshToken{
		MemberID:  memberID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
}

// IsExpired checks if the refresh token is expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// OTP represents a one-time password entity
type OTP struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Email     string    `gorm:"column:email;index;not null" json:"email"`
	Code      string    `gorm:"column:code;not null" json:"code"`
	Purpose   string    `gorm:"column:purpose;not null" json:"purpose"` // signup, password-reset
	ExpiresAt time.Time `gorm:"column:expires_at;not null" json:"expiresAt"`
	Verified  bool      `gorm:"column:verified;default:false" json:"verified"`
	BaseEntity
}

// TableName specifies the table name for OTP
func (OTP) TableName() string {
	return "otp"
}

// NewOTP creates a new OTP
func NewOTP(email, code, purpose string, expiresAt time.Time) *OTP {
	return &OTP{
		Email:     email,
		Code:      code,
		Purpose:   purpose,
		ExpiresAt: expiresAt,
		Verified:  false,
	}
}

// IsExpired checks if the OTP is expired
func (o *OTP) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

// IsValid checks if the OTP is valid (not expired and not verified)
func (o *OTP) IsValid() bool {
	return !o.IsExpired() && !o.Verified
}

// MarkAsVerified marks the OTP as verified
func (o *OTP) MarkAsVerified() {
	o.Verified = true
}

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"` // in seconds
}

// AuthCredentials represents authentication credentials
type AuthCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthClaims represents JWT claims
type AuthClaims struct {
	MemberID uint64 `json:"memberId"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

// Domain errors
var (
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrTokenExpired         = errors.New("token has expired")
	ErrTokenInvalid         = errors.New("invalid token")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrOTPNotFound          = errors.New("OTP not found")
	ErrOTPExpired           = errors.New("OTP has expired")
	ErrOTPAlreadyVerified   = errors.New("OTP already verified")
	ErrOTPInvalid           = errors.New("invalid OTP")
	ErrMemberNotFound       = errors.New("member not found")
	ErrMemberAlreadyExists  = errors.New("member already exists")
	ErrInvalidEmail         = errors.New("invalid email format")
	ErrInvalidPassword      = errors.New("invalid password format")
	ErrPasswordMismatch     = errors.New("password mismatch")
)
