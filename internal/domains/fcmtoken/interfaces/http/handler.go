package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pray-together/internal/domains/fcmtoken/application"
)

// Handler handles HTTP requests for FCM token operations
type Handler struct {
	registerTokenUseCase *application.RegisterTokenUseCase
	removeTokenUseCase   *application.RemoveTokenUseCase
}

// NewHandler creates a new FCM token handler
func NewHandler(
	registerTokenUseCase *application.RegisterTokenUseCase,
	removeTokenUseCase *application.RemoveTokenUseCase,
) *Handler {
	return &Handler{
		registerTokenUseCase: registerTokenUseCase,
		removeTokenUseCase:   removeTokenUseCase,
	}
}

// RegisterRoutes registers FCM token routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	fcm := router.Group("/fcm-token")
	{
		// Match Java API exactly:
		fcm.POST("", h.RegisterToken) // POST /fcm-token
		fcm.DELETE("", h.RemoveToken) // DELETE /fcm-token
		// Removed: PUT (not in Java API)
	}
}

// RegisterTokenRequest represents the request to register an FCM token (matching Java)
type RegisterTokenRequest struct {
	FcmToken string `json:"fcmToken" binding:"required"`
}

// RegisterToken handles POST /fcm-token
func (h *Handler) RegisterToken(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req RegisterTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.registerTokenUseCase.Execute(c.Request.Context(), &application.RegisterTokenRequest{
		MemberID:   memberID.(uint64),
		Token:      req.FcmToken,
		DeviceType: "unknown", // Java doesn't specify device type
		DeviceID:   "",        // Java doesn't specify device ID
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register FCM token"})
		return
	}

	// Java returns HTTP 200 with no body
	c.Status(http.StatusOK)
}

// RemoveTokenRequest represents the request to remove an FCM token (matching Java)
type RemoveTokenRequest struct {
	FcmToken string `json:"fcmToken" binding:"required"`
}

// RemoveToken handles DELETE /fcm-token
func (h *Handler) RemoveToken(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req RemoveTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Remove token by token value (matching Java)
	err := h.removeTokenUseCase.Execute(c.Request.Context(), &application.RemoveTokenRequest{
		MemberID: memberID.(uint64),
		Token:    req.FcmToken,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove FCM token"})
		return
	}

	// Java returns HTTP 200 with no body
	c.Status(http.StatusOK)
}

// UpdateTokenRequest represents the request to update an FCM token
type UpdateTokenRequest struct {
	OldToken string `json:"oldToken" binding:"required"`
	NewToken string `json:"newToken" binding:"required"`
	DeviceID string `json:"deviceId" binding:"required"`
}

// UpdateToken handles PUT /fcm-token
func (h *Handler) UpdateToken(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req UpdateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Remove old token
	_ = h.removeTokenUseCase.Execute(c.Request.Context(), &application.RemoveTokenRequest{
		MemberID: memberID.(uint64),
		Token:    req.OldToken,
	})

	// Register new token
	err := h.registerTokenUseCase.Execute(c.Request.Context(), &application.RegisterTokenRequest{
		MemberID:   memberID.(uint64),
		Token:      req.NewToken,
		DeviceType: "unknown", // Would need to get from request or storage
		DeviceID:   req.DeviceID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update FCM token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "FCM token updated successfully",
	})
}
