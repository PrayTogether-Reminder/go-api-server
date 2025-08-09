package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"pray-together/internal/domains/member/application"
	"pray-together/internal/domains/member/domain"
)

// Handler represents member HTTP handler
type Handler struct {
	getMemberUseCase    *application.GetMemberUseCase
	updateMemberUseCase *application.UpdateMemberUseCase
	deleteMemberUseCase *application.DeleteMemberUseCase
}

// NewHandler creates a new member handler
func NewHandler(
	getMemberUseCase *application.GetMemberUseCase,
	updateMemberUseCase *application.UpdateMemberUseCase,
	deleteMemberUseCase *application.DeleteMemberUseCase,
) *Handler {
	return &Handler{
		getMemberUseCase:    getMemberUseCase,
		updateMemberUseCase: updateMemberUseCase,
		deleteMemberUseCase: deleteMemberUseCase,
	}
}

// RegisterRoutes registers member routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	members := router.Group("/members")
	{
		members.GET("/me", h.GetMyProfile)
		// Java API only has GET /me endpoint
		// Removed: GET /:id, PUT /me, DELETE /me, GET /search
	}
}

// GetMyProfile handles GET /members/me
func (h *Handler) GetMyProfile(c *gin.Context) {
	// Get member ID from context (set by auth middleware)
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	req := &application.GetMemberRequest{
		MemberID: memberID.(uint64),
	}

	resp, err := h.getMemberUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		if err == domain.ErrMemberNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMemberProfile handles GET /members/:id
func (h *Handler) GetMemberProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member ID"})
		return
	}

	req := &application.GetMemberRequest{
		MemberID: id,
	}

	resp, err := h.getMemberUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		if err == domain.ErrMemberNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateMyProfileRequest represents update profile request
type UpdateMyProfileRequest struct {
	Name string `json:"name" binding:"required,min=2,max=30"`
}

// UpdateMyProfile handles PUT /members/me
func (h *Handler) UpdateMyProfile(c *gin.Context) {
	// Get member ID from context
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req UpdateMyProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateReq := &application.UpdateMemberRequest{
		MemberID: memberID.(uint64),
		Name:     req.Name,
	}

	resp, err := h.updateMemberUseCase.Execute(c.Request.Context(), updateReq)
	if err != nil {
		if err == domain.ErrMemberNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
			return
		}
		if err == domain.ErrInvalidMemberName {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteAccountRequest represents delete account request
type DeleteAccountRequest struct {
	Password string `json:"password" binding:"required"`
}

// DeleteMyAccount handles DELETE /members/me
func (h *Handler) DeleteMyAccount(c *gin.Context) {
	// Get member ID from context
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deleteReq := &application.DeleteMemberRequest{
		MemberID: memberID.(uint64),
		Password: req.Password,
	}

	resp, err := h.deleteMemberUseCase.Execute(c.Request.Context(), deleteReq)
	if err != nil {
		if err == domain.ErrMemberNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
			return
		}
		if err == domain.ErrInvalidPassword {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// SearchMembers handles GET /members/search
func (h *Handler) SearchMembers(c *gin.Context) {
	// Get search parameters
	email := c.Query("email")
	name := c.Query("name")

	if email == "" && name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email or name parameter is required"})
		return
	}

	// Search members
	var members []*domain.MemberInfo
	var err error

	if email != "" {
		// Search by email (exact match for privacy)
		member, searchErr := h.getMemberUseCase.GetMemberByEmail(c.Request.Context(), email)
		if searchErr == nil && member != nil {
			members = append(members, member)
		}
		err = searchErr
	} else if name != "" {
		// Search by name (partial match)
		members, err = h.getMemberUseCase.SearchMembersByName(c.Request.Context(), name)
	}

	if err != nil && err != domain.ErrMemberNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search members"})
		return
	}

	// Filter out sensitive information for search results
	results := make([]gin.H, 0, len(members))
	for _, member := range members {
		results = append(results, gin.H{
			"id":    member.ID,
			"name":  member.Name,
			"email": member.Email,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"members": results,
		"count":   len(results),
	})
}
