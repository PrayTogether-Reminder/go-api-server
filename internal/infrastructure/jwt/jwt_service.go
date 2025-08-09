package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService handles JWT token operations
type JWTService struct {
	secretKey          string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, accessTokenExpiry, refreshTokenExpiry time.Duration) *JWTService {
	return &JWTService{
		secretKey:          secretKey,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// Claims represents JWT claims
type Claims struct {
	MemberID uint64 `json:"member_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a new access token
func (s *JWTService) GenerateAccessToken(memberID uint64, email, name string) (string, error) {
	claims := &Claims{
		MemberID: memberID,
		Email:    email,
		Name:     name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "pray-together",
			Subject:   fmt.Sprintf("%d", memberID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a new refresh token
func (s *JWTService) GenerateRefreshToken() (string, error) {
	// Generate a random 32-byte token
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode to base64 URL-safe string
	return base64.URLEncoding.EncodeToString(b), nil
}

// ValidateAccessToken validates an access token and returns the claims
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

// GetAccessTokenExpiry returns the access token expiry duration
func (s *JWTService) GetAccessTokenExpiry() time.Duration {
	return s.accessTokenExpiry
}

// GetRefreshTokenExpiry returns the refresh token expiry duration
func (s *JWTService) GetRefreshTokenExpiry() time.Duration {
	return s.refreshTokenExpiry
}

// GenerateTokenPair generates both access and refresh tokens
func (s *JWTService) GenerateTokenPair(memberID uint64, email, name string) (accessToken string, refreshToken string, err error) {
	accessToken, err = s.GenerateAccessToken(memberID, email, name)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = s.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ExtractMemberID extracts member ID from token without full validation
func (s *JWTService) ExtractMemberID(tokenString string) (uint64, error) {
	claims, err := s.ValidateAccessToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.MemberID, nil
}
