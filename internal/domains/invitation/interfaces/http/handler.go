package http

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"pray-together/internal/domains/invitation/application"
	"pray-together/internal/domains/invitation/domain"
)

// Handler handles HTTP requests for invitation operations
type Handler struct {
	sendInvitationUseCase   *application.SendInvitationUseCase
	acceptInvitationUseCase *application.AcceptInvitationUseCase
	rejectInvitationUseCase *application.RejectInvitationUseCase
	listInvitationsUseCase  *application.ListInvitationsUseCase
}

// NewHandler creates a new invitation handler
func NewHandler(
	sendInvitationUseCase *application.SendInvitationUseCase,
	acceptInvitationUseCase *application.AcceptInvitationUseCase,
	rejectInvitationUseCase *application.RejectInvitationUseCase,
	listInvitationsUseCase *application.ListInvitationsUseCase,
) *Handler {
	return &Handler{
		sendInvitationUseCase:   sendInvitationUseCase,
		acceptInvitationUseCase: acceptInvitationUseCase,
		rejectInvitationUseCase: rejectInvitationUseCase,
		listInvitationsUseCase:  listInvitationsUseCase,
	}
}

// RegisterRoutes registers invitation routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	invitations := router.Group("/invitations")
	{
		// Match Java API exactly:
		invitations.POST("", h.SendInvitation)           // POST /invitations
		invitations.GET("", h.GetMyInvitations)          // GET /invitations
		invitations.PATCH("/:id", h.RespondToInvitation) // PATCH /invitations/:id
		// Removed: DELETE /:id (not in Java API)
	}
}

// SendInvitationRequest represents the request to send an invitation (matching Java)
type SendInvitationRequest struct {
	RoomID uint64 `json:"roomId" binding:"required"`
	Email  string `json:"email" binding:"required,email"`
}

// SendInvitation handles POST /invitations
func (h *Handler) SendInvitation(c *gin.Context) {
	inviterID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req SendInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set expiration to 7 days from now
	expiresAt := time.Now().AddDate(0, 0, 7)

	_, err := h.sendInvitationUseCase.Execute(c.Request.Context(), &application.SendInvitationRequest{
		RoomID:       req.RoomID,
		InviterID:    inviterID.(uint64),
		InviteeEmail: req.Email,
		Message:      "", // Java doesn't have message field
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		// Log the actual error for debugging
		log.Printf("[DEBUG] Invitation creation error: %v", err)

		switch err {
		case domain.ErrNotAuthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to send invitations for this room"})
		case domain.ErrAlreadyInvited:
			c.JSON(http.StatusConflict, gin.H{"error": "user already invited to this room"})
		default:
			// For now, return 400 for member not found (matching Java behavior)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "초대를 완료했습니다.", // Matching Java message
	})
}

// GetMyInvitations handles GET /invitations
func (h *Handler) GetMyInvitations(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get query parameters for filtering
	status := c.Query("status") // PENDING, ACCEPTED, REJECTED, EXPIRED
	includeExpired := c.Query("includeExpired") == "true"

	invitations, err := h.listInvitationsUseCase.Execute(c.Request.Context(), &application.ListInvitationsRequest{
		InviteeID:      memberID.(uint64),
		Status:         status,
		IncludeExpired: includeExpired,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get invitations"})
		return
	}

	// Match Java response format exactly
	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
	})
}

// RespondToInvitationRequest represents the request to respond to an invitation (matching Java)
type RespondToInvitationRequest struct {
	Status string `json:"status" binding:"required,oneof=ACCEPTED REJECTED"`
}

// RespondToInvitation handles PATCH /invitations/:id
func (h *Handler) RespondToInvitation(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	invitationIDStr := c.Param("id")
	invitationID, err := strconv.ParseUint(invitationIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invitation ID"})
		return
	}

	// First try to parse as raw JSON to check for invalid types
	var rawReq map[string]interface{}
	if err := c.ShouldBindJSON(&rawReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"code":    "VALIDATION-001",
			"message": "Invalid request format",
		})
		return
	}

	// Check if status field exists and is a string
	statusVal, exists := rawReq["status"]
	if !exists || statusVal == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"code":    "VALIDATION-001",
			"message": "status field is required",
		})
		return
	}

	statusStr, ok := statusVal.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"code":    "VALIDATION-001",
			"message": "status must be a string",
		})
		return
	}

	// Validate status value
	if statusStr != "ACCEPTED" && statusStr != "REJECTED" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"code":    "VALIDATION-001",
			"message": "status must be either ACCEPTED or REJECTED",
		})
		return
	}

	req := RespondToInvitationRequest{
		Status: statusStr,
	}

	var respErr error
	if req.Status == "ACCEPTED" {
		_, respErr = h.acceptInvitationUseCase.Execute(c.Request.Context(), &application.AcceptInvitationRequest{
			InvitationID: invitationID,
			InviteeID:    memberID.(uint64),
		})
		if respErr != nil {
			log.Printf("[DEBUG] Accept invitation error: %v", respErr)
		}
	} else {
		_, respErr = h.rejectInvitationUseCase.Execute(c.Request.Context(), &application.RejectInvitationRequest{
			InvitationID: invitationID,
			InviteeID:    memberID.(uint64),
		})
		if respErr != nil {
			log.Printf("[DEBUG] Reject invitation error: %v", respErr)
		}
	}

	if respErr != nil {
		switch respErr {
		case domain.ErrInvitationNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "invitation not found"})
		case domain.ErrInvitationExpired:
			c.JSON(http.StatusGone, gin.H{"error": "invitation has expired"})
		case domain.ErrAlreadyResponded:
			// Match Java: returns 400 for already responded invitations
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"code":    "INVITATION-002",
				"message": "이미 응답한 초대장입니다.",
			})
		case domain.ErrNotAuthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to respond to this invitation"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to respond to invitation"})
		}
		return
	}

	var statusKorean string
	if req.Status == "ACCEPTED" {
		statusKorean = "수락"
	} else {
		statusKorean = "거절"
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "기도방 초대를 " + statusKorean + "했습니다.", // Matching Java message
	})
}

// CancelInvitation handles DELETE /invitations/:id
func (h *Handler) CancelInvitation(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	invitationIDStr := c.Param("id")
	invitationID, err := strconv.ParseUint(invitationIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invitation ID"})
		return
	}

	// This would need a CancelInvitationUseCase implementation
	// For now, return a placeholder response
	_ = memberID
	_ = invitationID

	c.JSON(http.StatusOK, gin.H{
		"message": "invitation cancelled successfully",
	})
}
