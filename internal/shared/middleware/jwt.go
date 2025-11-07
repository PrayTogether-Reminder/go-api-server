package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/config"
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	AuthorizationHeader = "Authorization"
	BearerScheme        = "Bearer"
	UserIDKey           = "user_id"
	UserEmailKey        = "user_email"
)

// JWT error constants (errInfo)
const (
	missingToken  = "MISSING_TOKEN"
	invalidToken  = "INVALID_TOKEN"
	expiredToken  = "EXPIRED_TOKEN"
	invalidClaims = "INVALID_CLAIMS"
)

// Domain errors
var (
	ErrMissingToken  = sharedError.NewDomainError(missingToken)
	ErrInvalidToken  = sharedError.NewDomainError(invalidToken)
	ErrExpiredToken  = sharedError.NewDomainError(expiredToken)
	ErrInvalidClaims = sharedError.NewDomainError(invalidClaims)
)

// Register JWT error responses
func init() {
	sharedError.RegisterDomainErrorResponse(missingToken, sharedError.ErrorResponse{
		Status:  http.StatusUnauthorized,
		Code:    "AUTH-001",
		Message: "로그인을 해주세요.",
	})

	sharedError.RegisterDomainErrorResponse(invalidToken, sharedError.ErrorResponse{
		Status:  http.StatusUnauthorized,
		Code:    "AUTH-002",
		Message: "로그인을 해주세요.",
	})

	sharedError.RegisterDomainErrorResponse(expiredToken, sharedError.ErrorResponse{
		Status:  http.StatusUnauthorized,
		Code:    "AUTH-003",
		Message: "로그인을 해주세요.",
	})

	sharedError.RegisterDomainErrorResponse(invalidClaims, sharedError.ErrorResponse{
		Status:  http.StatusUnauthorized,
		Code:    "AUTH-004",
		Message: "로그인을 해주세요.",
	})
}

type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	jwt.RegisteredClaims
}

func JWT(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 요청 정보 (로깅용)
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		userAgent := c.Request.UserAgent()

		// Step 1: 토큰 추출
		token, err := extractToken(c)
		if err != nil {
			// 에러 발생 지점에서 바로 로깅
			slog.Warn("JWT 토큰 추출 실패",
				"step", "extract_token",
				"error", err.Error(),
				"client_ip", clientIP,
				"method", method,
				"path", path,
				"user_agent", userAgent,
			)
			handleJWTError(c, err)
			return
		}

		// Step 2: 토큰 검증
		claims, err := ValidateToken(token, cfg.JWT.Secret)
		if err != nil {
			// 에러 발생 지점에서 바로 로깅
			slog.Warn("JWT 토큰 검증 실패",
				"step", "validate_token",
				"error", err.Error(),
				"client_ip", clientIP,
				"method", method,
				"path", path,
				"user_agent", userAgent,
			)
			handleJWTError(c, err)
			return
		}

		// 인증 성공 - Context에 사용자 정보 저장
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Next()
	}
}

// handleJWTError handles JWT errors using the standardized error response format
// Note: Logging is done at the point of error detection in JWT() function
func handleJWTError(c *gin.Context, err error) {
	if resp, ok := sharedError.ResolveDomainError(err); ok {
		c.JSON(resp.Status, resp)
	} else {
		// 예상치 못한 에러 → Fallback 응답
		c.JSON(http.StatusUnauthorized, sharedError.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Code:    "AUTH-999",
			Message: "인증에 실패했습니다.",
		})
	}
	c.Abort()
}

func GenerateToken(userID, email string, cfg *config.Config) (string, error) {
	now := time.Now()
	expiresAt := now.Add(cfg.JWT.Expiry)

	claims := Claims{
		UserID:    userID,
		Email:     email,
		ExpiresAt: expiresAt.Unix(),
		IssuedAt:  now.Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    cfg.App.Name,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

func GenerateRefreshToken(userID string, cfg *config.Config) (string, error) {
	now := time.Now()
	expiresAt := now.Add(cfg.JWT.RefreshExpiry)

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(now),
		Issuer:    cfg.App.Name,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

func ValidateToken(tokenString, secret string) (*Claims, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

	token, err := parser.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader(AuthorizationHeader)
	if authHeader == "" {
		return "", ErrMissingToken
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], BearerScheme) {
		return "", ErrInvalidToken
	}

	return parts[1], nil
}

func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return "", false
	}

	id, ok := userID.(string)
	return id, ok
}

func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(UserEmailKey)
	if !exists {
		return "", false
	}

	e, ok := email.(string)
	return e, ok
}
