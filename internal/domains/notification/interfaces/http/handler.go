package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pray-together/internal/domains/notification/application"
)

// Handler handles HTTP requests for notification operations
type Handler struct {
	sendNotificationUseCase           *application.SendNotificationUseCase
	sendBulkNotificationUseCase       *application.SendBulkNotificationUseCase
	listNotificationsUseCase          *application.ListNotificationsUseCase
	markNotificationReadUseCase       *application.MarkNotificationReadUseCase
	updateNotificationSettingsUseCase *application.UpdateNotificationSettingsUseCase
}

// NewHandler creates a new notification handler
func NewHandler(
	sendNotificationUseCase *application.SendNotificationUseCase,
	sendBulkNotificationUseCase *application.SendBulkNotificationUseCase,
	listNotificationsUseCase *application.ListNotificationsUseCase,
	markNotificationReadUseCase *application.MarkNotificationReadUseCase,
	updateNotificationSettingsUseCase *application.UpdateNotificationSettingsUseCase,
) *Handler {
	return &Handler{
		sendNotificationUseCase:           sendNotificationUseCase,
		sendBulkNotificationUseCase:       sendBulkNotificationUseCase,
		listNotificationsUseCase:          listNotificationsUseCase,
		markNotificationReadUseCase:       markNotificationReadUseCase,
		updateNotificationSettingsUseCase: updateNotificationSettingsUseCase,
	}
}

// RegisterRoutes registers notification routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	notifications := router.Group("/notifications")
	{
		notifications.GET("", h.GetMyNotifications)
		notifications.POST("/read/:id", h.MarkAsRead)
		notifications.POST("/read-all", h.MarkAllAsRead)
		notifications.GET("/unread-count", h.GetUnreadCount)
		notifications.PUT("/settings", h.UpdateNotificationSettings)
	}

	// Admin routes for sending notifications
	admin := router.Group("/admin/notifications")
	{
		admin.POST("", h.SendNotification)
		admin.POST("/bulk", h.SendBulkNotification)
	}
}

// GetMyNotifications handles GET /notifications
func (h *Handler) GetMyNotifications(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	unreadOnly := c.Query("unreadOnly") == "true"

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	notifications, err := h.listNotificationsUseCase.Execute(c.Request.Context(), &application.ListNotificationsRequest{
		MemberID:   memberID.(uint64),
		Limit:      limit,
		Offset:     offset,
		UnreadOnly: unreadOnly,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"count":         len(notifications),
	})
}

// MarkAsRead handles POST /notifications/read/:id
func (h *Handler) MarkAsRead(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	err = h.markNotificationReadUseCase.Execute(c.Request.Context(), &application.MarkNotificationReadRequest{
		NotificationID: notificationID,
		MemberID:       memberID.(uint64),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "notification marked as read",
	})
}

// MarkAllAsRead handles POST /notifications/read-all
func (h *Handler) MarkAllAsRead(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := h.markNotificationReadUseCase.MarkAllAsRead(c.Request.Context(), memberID.(uint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark all notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "all notifications marked as read",
	})
}

// GetUnreadCount handles GET /notifications/unread-count
func (h *Handler) GetUnreadCount(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	count, err := h.listNotificationsUseCase.GetUnreadCount(c.Request.Context(), memberID.(uint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unreadCount": count,
	})
}

// UpdateNotificationSettingsRequest represents the request to update notification settings
type UpdateNotificationSettingsRequest struct {
	PrayerCompletion bool `json:"prayerCompletion"`
	RoomInvitation   bool `json:"roomInvitation"`
	DailyReminder    bool `json:"dailyReminder"`
}

// UpdateNotificationSettings handles PUT /notifications/settings
func (h *Handler) UpdateNotificationSettings(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req UpdateNotificationSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.updateNotificationSettingsUseCase.Execute(c.Request.Context(), &application.UpdateNotificationSettingsRequest{
		MemberID:         memberID.(uint64),
		PrayerCompletion: req.PrayerCompletion,
		RoomInvitation:   req.RoomInvitation,
		DailyReminder:    req.DailyReminder,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update notification settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "notification settings updated",
	})
}

// SendNotificationRequest represents the request to send a notification
type SendNotificationRequest struct {
	RecipientID uint64                 `json:"recipientId" binding:"required"`
	Title       string                 `json:"title" binding:"required"`
	Body        string                 `json:"body" binding:"required"`
	Type        string                 `json:"type" binding:"required"`
	Data        map[string]interface{} `json:"data"`
}

// SendNotification handles POST /admin/notifications
func (h *Handler) SendNotification(c *gin.Context) {
	var req SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.sendNotificationUseCase.Execute(c.Request.Context(), &application.SendNotificationRequest{
		RecipientID: req.RecipientID,
		Title:       req.Title,
		Body:        req.Body,
		Type:        req.Type,
		Data:        req.Data,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "notification sent successfully",
	})
}

// SendBulkNotificationRequest represents the request to send bulk notifications
type SendBulkNotificationRequest struct {
	RecipientIDs []uint64               `json:"recipientIds" binding:"required"`
	Title        string                 `json:"title" binding:"required"`
	Body         string                 `json:"body" binding:"required"`
	Type         string                 `json:"type" binding:"required"`
	Data         map[string]interface{} `json:"data"`
}

// SendBulkNotification handles POST /admin/notifications/bulk
func (h *Handler) SendBulkNotification(c *gin.Context) {
	var req SendBulkNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results := h.sendBulkNotificationUseCase.Execute(c.Request.Context(), &application.SendBulkNotificationRequest{
		RecipientIDs: req.RecipientIDs,
		Title:        req.Title,
		Body:         req.Body,
		Type:         req.Type,
		Data:         req.Data,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "bulk notifications sent",
		"results": results,
	})
}
