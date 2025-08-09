package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"pray-together/internal/domains/prayer/application"
	"pray-together/internal/domains/prayer/domain"
)

// Handler represents prayer HTTP handler
type Handler struct {
	createPrayerUseCase       *application.CreatePrayerUseCase
	updatePrayerUseCase       *application.UpdatePrayerUseCase
	markPrayerAnsweredUseCase *application.MarkPrayerAnsweredUseCase
	prayerService             *domain.Service
}

// NewHandler creates a new prayer handler
func NewHandler(
	createPrayerUseCase *application.CreatePrayerUseCase,
	updatePrayerUseCase *application.UpdatePrayerUseCase,
	markPrayerAnsweredUseCase *application.MarkPrayerAnsweredUseCase,
	prayerService *domain.Service,
) *Handler {
	return &Handler{
		createPrayerUseCase:       createPrayerUseCase,
		updatePrayerUseCase:       updatePrayerUseCase,
		markPrayerAnsweredUseCase: markPrayerAnsweredUseCase,
		prayerService:             prayerService,
	}
}

// RegisterRoutes registers prayer routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	prayers := router.Group("/prayers")
	{
		prayers.POST("", h.CreatePrayer)
		prayers.GET("", h.GetPrayers)
		prayers.GET("/:id", h.GetPrayer)
		prayers.PUT("/:id", h.UpdatePrayer)
		prayers.DELETE("/:id", h.DeletePrayer)
		prayers.POST("/:id/complete", h.CompletePrayer)
		prayers.GET("/titles", h.GetPrayerTitles)
	}
}

// CreatePrayerRequest represents create prayer request
type CreatePrayerRequest struct {
	RoomID  uint64 `json:"roomId" binding:"required"`
	Content string `json:"content" binding:"required,min=1,max=1000"`
	Type    string `json:"type" binding:"required,oneof=PERSONAL SHARED"`
}

// CreatePrayer handles POST /prayers
func (h *Handler) CreatePrayer(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreatePrayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createReq := &application.CreatePrayerRequest{
		MemberID: memberID.(uint64),
		RoomID:   req.RoomID,
		Content:  req.Content,
		Type:     domain.PrayerType(req.Type),
	}

	resp, err := h.createPrayerUseCase.Execute(c.Request.Context(), createReq)
	if err != nil {
		if err == domain.ErrUnauthorizedAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetPrayers handles GET /prayers
func (h *Handler) GetPrayers(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get query parameters
	roomIDStr := c.Query("roomId")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	var prayers []*domain.Prayer
	var err error

	if roomIDStr != "" {
		roomID, _ := strconv.ParseUint(roomIDStr, 10, 64)
		prayers, err = h.prayerService.GetRoomPrayers(c.Request.Context(), roomID, memberID.(uint64), limit, offset)
	} else {
		prayers, err = h.prayerService.GetMemberPrayers(c.Request.Context(), memberID.(uint64), limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Convert to response format
	var prayerInfos []gin.H
	for _, prayer := range prayers {
		prayerInfos = append(prayerInfos, gin.H{
			"id":         prayer.ID,
			"memberId":   prayer.MemberID,
			"roomId":     prayer.RoomID,
			"content":    prayer.Content,
			"type":       prayer.Type,
			"isAnswered": prayer.IsAnswered,
			"answeredAt": prayer.AnsweredAt,
			"createdAt":  prayer.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"prayers": prayerInfos,
		"hasMore": len(prayers) == limit,
	})
}

// GetPrayer handles GET /prayers/:id
func (h *Handler) GetPrayer(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	prayerIDStr := c.Param("id")
	prayerID, err := strconv.ParseUint(prayerIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prayer ID"})
		return
	}

	prayer, err := h.prayerService.GetPrayer(c.Request.Context(), prayerID, memberID.(uint64))
	if err != nil {
		if err == domain.ErrPrayerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "prayer not found"})
			return
		}
		if err == domain.ErrUnauthorizedAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, prayer.ToInfo())
}

// UpdatePrayerRequest represents update prayer request
type UpdatePrayerRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}

// UpdatePrayer handles PUT /prayers/:id
func (h *Handler) UpdatePrayer(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	prayerIDStr := c.Param("id")
	prayerID, err := strconv.ParseUint(prayerIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prayer ID"})
		return
	}

	var req UpdatePrayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateReq := &application.UpdatePrayerRequest{
		PrayerID: prayerID,
		MemberID: memberID.(uint64),
		Content:  req.Content,
	}

	resp, err := h.updatePrayerUseCase.Execute(c.Request.Context(), updateReq)
	if err != nil {
		if err == domain.ErrPrayerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "prayer not found"})
			return
		}
		if err == domain.ErrUnauthorizedAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeletePrayer handles DELETE /prayers/:id
func (h *Handler) DeletePrayer(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	prayerIDStr := c.Param("id")
	prayerID, err := strconv.ParseUint(prayerIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prayer ID"})
		return
	}

	// Validate positive ID (matching Java's @Positive)
	if prayerID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prayer ID"})
		return
	}

	err = h.prayerService.DeletePrayer(c.Request.Context(), prayerID, memberID.(uint64))
	if err != nil {
		if err == domain.ErrPrayerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "prayer not found"})
			return
		}
		if err == domain.ErrUnauthorizedAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "prayer deleted successfully"})
}

// CompletePrayerRequest represents complete prayer request
type CompletePrayerRequest struct {
	Note string `json:"note"`
}

// CompletePrayer handles POST /prayers/:id/complete
func (h *Handler) CompletePrayer(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	prayerIDStr := c.Param("id")
	prayerID, err := strconv.ParseUint(prayerIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prayer ID"})
		return
	}

	var req CompletePrayerRequest
	_ = c.ShouldBindJSON(&req)

	markReq := &application.MarkPrayerAnsweredRequest{
		PrayerID: prayerID,
		MemberID: memberID.(uint64),
	}

	resp, err := h.markPrayerAnsweredUseCase.Execute(c.Request.Context(), markReq)
	if err != nil {
		if err == domain.ErrPrayerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "prayer not found"})
			return
		}
		if err == domain.ErrUnauthorizedAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetPrayerTitles handles GET /prayers/titles
func (h *Handler) GetPrayerTitles(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	roomIDStr := c.Query("roomId")
	if roomIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roomId is required"})
		return
	}

	roomID, _ := strconv.ParseUint(roomIDStr, 10, 64)

	// Get shared prayers from room as "titles"
	prayers, err := h.prayerService.GetRoomPrayers(c.Request.Context(), roomID, memberID.(uint64), 50, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Convert to title format
	var titles []gin.H
	for _, prayer := range prayers {
		// Use first 50 characters of content as title
		title := prayer.Content
		if len(title) > 50 {
			title = title[:50] + "..."
		}

		titles = append(titles, gin.H{
			"id":        prayer.ID,
			"title":     title,
			"memberId":  prayer.MemberID,
			"createdAt": prayer.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"titles": titles})
}
