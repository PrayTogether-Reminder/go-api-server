package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"pray-together/internal/common/errors"
	"pray-together/internal/domains/room/application"
	"pray-together/internal/domains/room/domain"
)

// Handler represents room HTTP handler
type Handler struct {
	createRoomUseCase     *application.CreateRoomUseCase
	joinRoomUseCase       *application.JoinRoomUseCase
	getRoomDetailsUseCase *application.GetRoomDetailsUseCase
	roomService           *domain.Service
	getMemberName         func(ctx context.Context, memberID uint64) (string, error)
}

// NewHandler creates a new room handler
func NewHandler(
	createRoomUseCase *application.CreateRoomUseCase,
	joinRoomUseCase *application.JoinRoomUseCase,
	getRoomDetailsUseCase *application.GetRoomDetailsUseCase,
	roomService *domain.Service,
	getMemberName func(ctx context.Context, memberID uint64) (string, error),
) *Handler {
	return &Handler{
		createRoomUseCase:     createRoomUseCase,
		joinRoomUseCase:       joinRoomUseCase,
		getRoomDetailsUseCase: getRoomDetailsUseCase,
		roomService:           roomService,
		getMemberName:         getMemberName,
	}
}

// RegisterRoutes registers room routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	rooms := router.Group("/rooms")
	{
		rooms.POST("", h.CreateRoom)
		rooms.GET("", h.GetMyRooms)
		rooms.DELETE("/:id", h.DeleteRoom)
		rooms.GET("/:id/members", h.GetRoomMembers)
		// Removed endpoints not in Java API:
		// GET /:id, PUT /:id, POST /:id/join, POST /:id/leave
		// PUT /:id/members/:memberId/role, PUT /:id/notification
	}
}

// CreateRoomRequest represents create room request (matching Java)
type CreateRoomRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Description string `json:"description" binding:"required,min=1,max=200"`
}

// CreateRoom handles POST /rooms
func (h *Handler) CreateRoom(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, err)
		return
	}

	// Additional validation for whitespace-only strings (matching Java's @NotBlank)
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			http.StatusBadRequest,
			errors.MethodArgumentNotValid.Code,
			"방 이름을 작성해 주세요.",
		))
		return
	}

	if strings.TrimSpace(req.Description) == "" {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			http.StatusBadRequest,
			errors.MethodArgumentNotValid.Code,
			"방 설명을 작성해 주세요.",
		))
		return
	}

	createReq := &application.CreateRoomRequest{
		CreatorID:             memberID.(uint64),
		RoomName:              req.Name,
		Description:           req.Description,
		IsPrivate:             false, // Default value matching Java
		PrayStartTime:         "00:00",
		PrayEndTime:           "23:59",
		NotificationStartTime: "00:00",
		NotificationEndTime:   "23:59",
	}

	resp, err := h.createRoomUseCase.Execute(c.Request.Context(), createReq)
	if err != nil {
		if err == domain.ErrInvalidRoomName {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetMyRooms handles GET /rooms with infinite scroll
func (h *Handler) GetMyRooms(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get infinite scroll parameters (matching Java)
	_ = c.DefaultQuery("orderBy", "time") // For future use
	after := c.DefaultQuery("after", "0")
	dir := c.DefaultQuery("dir", "desc")

	// Match Java: if after is "0" or empty, it's the first page
	if after == "" {
		after = "0"
	}

	// Get paginated rooms using repository method (matching Java fetchRoomsByMember)
	roomInfos, err := h.roomService.GetMemberRoomsPaginated(c.Request.Context(), memberID.(uint64), after, dir, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Match Java RoomInfiniteScrollResponse format - just return the rooms array
	c.JSON(http.StatusOK, gin.H{
		"rooms": roomInfos,
	})
}

// GetRoomDetails handles GET /rooms/:id
func (h *Handler) GetRoomDetails(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}

	req := &application.GetRoomDetailsRequest{
		RoomID:   roomID,
		MemberID: memberID.(uint64),
	}

	resp, err := h.getRoomDetailsUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		if err == domain.ErrRoomNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
			return
		}
		if err == domain.ErrMemberNotInRoom {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateRoomRequest represents update room request
type UpdateRoomRequest struct {
	RoomName              string `json:"roomName,omitempty"`
	IsPrivate             *bool  `json:"isPrivate,omitempty"`
	PrayStartTime         string `json:"prayStartTime,omitempty"`
	PrayEndTime           string `json:"prayEndTime,omitempty"`
	NotificationStartTime string `json:"notificationStartTime,omitempty"`
	NotificationEndTime   string `json:"notificationEndTime,omitempty"`
}

// UpdateRoom handles PUT /rooms/:id
func (h *Handler) UpdateRoom(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}

	var req UpdateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.RoomName != "" {
		updates["roomName"] = req.RoomName
	}
	if req.IsPrivate != nil {
		updates["isPrivate"] = *req.IsPrivate
	}
	if req.PrayStartTime != "" {
		updates["prayStartTime"] = req.PrayStartTime
	}
	if req.PrayEndTime != "" {
		updates["prayEndTime"] = req.PrayEndTime
	}
	if req.NotificationStartTime != "" {
		updates["notificationStartTime"] = req.NotificationStartTime
	}
	if req.NotificationEndTime != "" {
		updates["notificationEndTime"] = req.NotificationEndTime
	}

	// Update room using service
	roomService := domain.NewService(nil)
	room, err := roomService.UpdateRoom(c.Request.Context(), roomID, memberID.(uint64), updates)
	if err != nil {
		if err == domain.ErrNotRoomOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "only room owner can update room"})
			return
		}
		if err == domain.ErrRoomNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, room.ToInfo())
}

// DeleteRoom handles DELETE /rooms/:id
func (h *Handler) DeleteRoom(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}

	// Validate positive ID (matching Java's @Positive)
	if roomID == 0 {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			http.StatusBadRequest,
			errors.MethodArgumentNotValid.Code,
			"잘 못된 방을 선택하셨습니다.",
		))
		return
	}

	// Java: deleteMemberRoomById - removes member from room, not delete the room itself
	err = h.roomService.LeaveRoom(c.Request.Context(), roomID, memberID.(uint64))
	if err != nil {
		if err == domain.ErrMemberNotInRoom {
			c.JSON(http.StatusNotFound, gin.H{"error": "not a member of this room"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "방을 나갔습니다."}) // Matching Java message
}

// JoinRoom handles POST /rooms/:id/join
func (h *Handler) JoinRoom(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}

	req := &application.JoinRoomRequest{
		RoomID:   roomID,
		MemberID: memberID.(uint64),
	}

	resp, err := h.joinRoomUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		if err == domain.ErrRoomNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
			return
		}
		if err == domain.ErrMemberAlreadyInRoom {
			c.JSON(http.StatusConflict, gin.H{"error": "already a member of this room"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// LeaveRoom handles POST /rooms/:id/leave
func (h *Handler) LeaveRoom(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}

	roomService := domain.NewService(nil)
	err = roomService.LeaveRoom(c.Request.Context(), roomID, memberID.(uint64))
	if err != nil {
		if err == domain.ErrMemberNotInRoom {
			c.JSON(http.StatusNotFound, gin.H{"error": "not a member of this room"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "left room successfully"})
}

// GetRoomMembers handles GET /rooms/:id/members
func (h *Handler) GetRoomMembers(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}

	// Validate positive ID (matching Java's @Min(value = 1))
	if roomID == 0 {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			http.StatusBadRequest,
			errors.MethodArgumentNotValid.Code,
			"잘 못된 방을 선택하셨습니다.",
		))
		return
	}

	// Validate member has access to room
	if err := h.roomService.ValidateRoomAccess(c.Request.Context(), roomID, memberID.(uint64)); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	members, err := h.roomService.GetRoomMembers(c.Request.Context(), roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Convert to response format with member names (matching Java MemberIdName)
	var memberInfos []gin.H
	for _, member := range members {
		memberName := ""
		if h.getMemberName != nil {
			name, _ := h.getMemberName(c.Request.Context(), member.MemberID)
			memberName = name
		}

		memberInfos = append(memberInfos, gin.H{
			"id":   member.MemberID,
			"name": memberName,
		})
	}

	c.JSON(http.StatusOK, gin.H{"members": memberInfos})
}

// UpdateMemberRoleRequest represents update member role request
type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=OWNER MEMBER"`
}

// UpdateMemberRole handles PUT /rooms/:id/members/:memberId/role
func (h *Handler) UpdateMemberRole(c *gin.Context) {
	updaterID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}

	memberIDStr := c.Param("memberId")
	targetMemberID, err := strconv.ParseUint(memberIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member ID"})
		return
	}

	var req UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roomService := domain.NewService(nil)
	err = roomService.UpdateMemberRole(
		c.Request.Context(),
		roomID,
		updaterID.(uint64),
		targetMemberID,
		domain.RoomRole(req.Role),
	)

	if err != nil {
		if err == domain.ErrNotRoomOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "only room owner can update member roles"})
			return
		}
		if err == domain.ErrMemberNotInRoom {
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found in room"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member role updated successfully"})
}

// UpdateNotificationSettingsRequest represents notification settings update request
type UpdateNotificationSettingsRequest struct {
	Enabled bool `json:"enabled"`
}

// UpdateNotificationSettings handles PUT /rooms/:id/notification
func (h *Handler) UpdateNotificationSettings(c *gin.Context) {
	_, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	roomIDStr := c.Param("id")
	_, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}

	var req UpdateNotificationSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// This would need to be implemented in the domain service
	// For now, return success
	c.JSON(http.StatusOK, gin.H{
		"message": "notification settings updated",
		"enabled": req.Enabled,
	})
}
