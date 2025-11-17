package http

import (
	"net/http"
	"strconv"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/logger"
	"github.com/gin-gonic/gin"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// Context keys for storing user authentication information in context.Context
const (
	memberIDContextKey    contextKey = "member_id"
	memberEmailContextKey contextKey = "member_email"
)

// Gin context keys (string type for Gin's c.Set/c.Get)
const (
	MemberIDKey    = "member_id"
	MemberEmailKey = "member_email"
)

// ContextKey returns the context.Context key for member ID (used in middleware)
func ContextKey() contextKey {
	return memberIDContextKey
}

// EmailContextKey returns the context.Context key for member email (used in middleware)
func EmailContextKey() contextKey {
	return memberEmailContextKey
}

func GetMemberID(c *gin.Context) (int64, bool) {
	memberID, exists := c.Get(MemberIDKey)
	if !exists {
		return 0, false
	}

	// JWT claims는 memberID를 string으로 저장
	// 여기서 int64로 변환
	memberIDStr, ok := memberID.(string)
	if !ok {
		return 0, false
	}

	memberIDInt, err := strconv.ParseInt(memberIDStr, 10, 64)
	if err != nil {
		return 0, false
	}

	return memberIDInt, true
}

// RequireMemberID retrieves the authenticated user's ID from the Gin context.
// If the user ID is not found, automatically sends an authentication error response.
// Returns the user ID and true if found, empty string and false if not found (error already sent).
// Use this in most handlers to reduce boilerplate.
func RequireMemberID(c *gin.Context) (int64, bool) {
	memberID, ok := GetMemberID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, sharedError.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Code:    "AUTH-000",
			Message: "로그인을 해주세요.",
		})
		c.Abort()
		logger.FromContext(c.Request.Context()).Error("[API] context에 회원 ID가 존재하지 않습니다.")
		return 0, false
	}
	return memberID, true
}
