package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pray-together/internal/common/errors"
	"pray-together/internal/domains/prayer/application"
	"pray-together/internal/domains/prayer/domain"
)

// HandlerV2 handles HTTP requests for prayer operations with two-level structure
type HandlerV2 struct {
	createPrayerTitleUseCase   *application.CreatePrayerTitleUseCase
	addPrayerContentUseCase    *application.AddPrayerContentUseCase
	updatePrayerTitleUseCase   *application.UpdatePrayerTitleUseCase
	updatePrayerContentUseCase *application.UpdatePrayerContentUseCase
	deletePrayerTitleUseCase   *application.DeletePrayerTitleUseCase
	deletePrayerContentUseCase *application.DeletePrayerContentUseCase
	listPrayerTitlesUseCase    *application.ListPrayerTitlesUseCase
	getPrayerDetailsUseCase    *application.GetPrayerDetailsUseCase
	completePrayerUseCase      *application.CompletePrayerUseCase
}

// NewHandlerV2 creates a new prayer handler with two-level structure
func NewHandlerV2(
	createPrayerTitleUseCase *application.CreatePrayerTitleUseCase,
	addPrayerContentUseCase *application.AddPrayerContentUseCase,
	updatePrayerTitleUseCase *application.UpdatePrayerTitleUseCase,
	updatePrayerContentUseCase *application.UpdatePrayerContentUseCase,
	deletePrayerTitleUseCase *application.DeletePrayerTitleUseCase,
	deletePrayerContentUseCase *application.DeletePrayerContentUseCase,
	listPrayerTitlesUseCase *application.ListPrayerTitlesUseCase,
	getPrayerDetailsUseCase *application.GetPrayerDetailsUseCase,
	completePrayerUseCase *application.CompletePrayerUseCase,
) *HandlerV2 {
	return &HandlerV2{
		createPrayerTitleUseCase:   createPrayerTitleUseCase,
		addPrayerContentUseCase:    addPrayerContentUseCase,
		updatePrayerTitleUseCase:   updatePrayerTitleUseCase,
		updatePrayerContentUseCase: updatePrayerContentUseCase,
		deletePrayerTitleUseCase:   deletePrayerTitleUseCase,
		deletePrayerContentUseCase: deletePrayerContentUseCase,
		listPrayerTitlesUseCase:    listPrayerTitlesUseCase,
		getPrayerDetailsUseCase:    getPrayerDetailsUseCase,
		completePrayerUseCase:      completePrayerUseCase,
	}
}

// RegisterRoutesV2 registers prayer routes with two-level structure
func (h *HandlerV2) RegisterRoutesV2(router *gin.RouterGroup) {
	prayers := router.Group("/prayers")
	{
		// Match Java API exactly:
		// POST /prayers - create prayer
		prayers.POST("", h.CreatePrayerTitle)
		// GET /prayers - list prayers with infinite scroll
		prayers.GET("", h.ListPrayerTitles)
		// PUT /prayers/:id - update prayer title
		prayers.PUT("/:id", h.UpdatePrayerTitle)
		// DELETE /prayers/:id - delete prayer
		prayers.DELETE("/:id", h.DeletePrayerTitle)
		// GET /prayers/:id/contents - get prayer contents
		prayers.GET("/:id/contents", h.GetPrayerContents)
		// POST /prayers/:id/completion - complete prayer
		prayers.POST("/:id/completion", h.CompletePrayer)

		// Removed endpoints not in Java API:
		// GET /:id (prayer details)
		// POST /:id/contents (add content)
		// PUT /contents/:contentId (update content)
		// DELETE /contents/:contentId (delete content)
		// POST /contents/:contentId/completion (complete content)
	}
}

// PrayerRequestContent represents prayer content in request (matching Java)
type PrayerRequestContent struct {
	MemberID   uint64 `json:"memberId"`
	MemberName string `json:"memberName" binding:"required,min=1"`
	Content    string `json:"content" binding:"required,min=1"`
}

// CreatePrayerTitleRequest represents the request to create a prayer title
type CreatePrayerTitleRequest struct {
	RoomID   uint64                 `json:"roomId" binding:"required"`
	Title    string                 `json:"title" binding:"required,min=1,max=50"`
	Contents []PrayerRequestContent `json:"contents" binding:"required"`
}

// CreatePrayerTitle handles POST /prayers
func (h *HandlerV2) CreatePrayerTitle(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreatePrayerTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, err)
		return
	}

	// Additional validation for contents array
	for _, content := range req.Contents {
		if content.MemberName == "" {
			c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
				errors.MethodArgumentNotValid.Status,
				errors.MethodArgumentNotValid.Code,
				"회원 이름은(는) 필수입니다.",
			))
			return
		}
		if content.Content == "" {
			c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
				errors.MethodArgumentNotValid.Status,
				errors.MethodArgumentNotValid.Code,
				"내용은(는) 필수입니다.",
			))
			return
		}
	}

	// Convert contents to application layer format
	contents := make([]application.PrayerContentRequest, len(req.Contents))
	for i, content := range req.Contents {
		contents[i] = application.PrayerContentRequest{
			MemberID:   content.MemberID,
			MemberName: content.MemberName,
			Content:    content.Content,
		}
	}

	prayerTitle, err := h.createPrayerTitleUseCase.Execute(c.Request.Context(), &application.CreatePrayerTitleRequest{
		RoomID:    req.RoomID,
		CreatorID: memberID.(uint64),
		Title:     req.Title,
		Contents:  contents,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create prayer"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "기도 제목을 생성했습니다.", // Matching Java message
		"prayer":  prayerTitle,
	})
}

// ListPrayerTitles handles GET /prayers with infinite scroll
func (h *HandlerV2) ListPrayerTitles(c *gin.Context) {
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

	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid roomId"})
		return
	}

	// Infinite scroll parameter (matching Java)
	// Java uses "after" with default value "" for first page
	after := c.DefaultQuery("after", "")

	prayers, err := h.listPrayerTitlesUseCase.Execute(c.Request.Context(), &application.ListPrayerTitlesRequest{
		RoomID:   roomID,
		MemberID: memberID.(uint64),
		After:    after,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get prayers"})
		return
	}

	// Match Java response format: PrayerTitleInfiniteScrollResponse
	// Java only returns prayerTitles array, no hasMore or after fields
	c.JSON(http.StatusOK, gin.H{
		"prayerTitles": prayers,
	})
}

// GetPrayerDetails handles GET /prayers/:id
func (h *HandlerV2) GetPrayerDetails(c *gin.Context) {
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

	prayer, err := h.getPrayerDetailsUseCase.Execute(c.Request.Context(), &application.GetPrayerDetailsRequest{
		PrayerTitleID: prayerID,
		MemberID:      memberID.(uint64),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get prayer details"})
		return
	}

	c.JSON(http.StatusOK, prayer)
}

// AddPrayerContentRequest represents the request to add prayer content
type AddPrayerContentRequest struct {
	Content string `json:"content" binding:"required"`
}

// AddPrayerContent handles POST /prayers/:id/contents
func (h *HandlerV2) AddPrayerContent(c *gin.Context) {
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

	var req AddPrayerContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	content, err := h.addPrayerContentUseCase.Execute(c.Request.Context(), &application.AddPrayerContentRequest{
		PrayerTitleID: prayerID,
		AuthorID:      memberID.(uint64),
		Content:       req.Content,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add prayer content"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "prayer content added successfully",
		"content": content,
	})
}

// GetPrayerContents handles GET /prayers/:id/contents
func (h *HandlerV2) GetPrayerContents(c *gin.Context) {
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

	prayer, err := h.getPrayerDetailsUseCase.Execute(c.Request.Context(), &application.GetPrayerDetailsRequest{
		PrayerTitleID: prayerID,
		MemberID:      memberID.(uint64),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get prayer contents"})
		return
	}

	// Match Java response format: PrayerContentResponse
	c.JSON(http.StatusOK, gin.H{
		"prayerContents": prayer.Contents,
	})
}

// UpdatePrayerTitleRequest represents the request to update a prayer title
type UpdatePrayerTitleRequest struct {
	Title    string                 `json:"title" binding:"required,min=1,max=50"`
	Contents []PrayerRequestContent `json:"contents" binding:"required"`
}

// UpdatePrayerTitle handles PUT /prayers/:id
func (h *HandlerV2) UpdatePrayerTitle(c *gin.Context) {
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

	var req UpdatePrayerTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert contents
	contents := make([]application.PrayerContentRequest, len(req.Contents))
	for i, content := range req.Contents {
		contents[i] = application.PrayerContentRequest{
			MemberID:   content.MemberID,
			MemberName: content.MemberName,
			Content:    content.Content,
		}
	}

	prayer, err := h.updatePrayerTitleUseCase.Execute(c.Request.Context(), &application.UpdatePrayerTitleRequest{
		PrayerTitleID: prayerID,
		MemberID:      memberID.(uint64),
		Title:         req.Title,
		Contents:      contents,
	})
	if err != nil {
		// Check for specific errors (matching Java)
		if err == domain.ErrPrayerNotFound || err == domain.ErrPrayerTitleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "prayer title not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update prayer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "기도를 변경했습니다.", // Matching Java message
		"prayer":  prayer,
	})
}

// UpdatePrayerContentRequest represents the request to update prayer content
type UpdatePrayerContentRequest struct {
	Content string `json:"content" binding:"required"`
}

// UpdatePrayerContent handles PUT /prayers/contents/:contentId
func (h *HandlerV2) UpdatePrayerContent(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	contentIDStr := c.Param("contentId")
	contentID, err := strconv.ParseUint(contentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content ID"})
		return
	}

	var req UpdatePrayerContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	content, err := h.updatePrayerContentUseCase.Execute(c.Request.Context(), &application.UpdatePrayerContentRequest{
		ContentID: contentID,
		MemberID:  memberID.(uint64),
		Content:   req.Content,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update prayer content"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "prayer content updated successfully",
		"content": content,
	})
}

// DeletePrayerTitle handles DELETE /prayers/:id
func (h *HandlerV2) DeletePrayerTitle(c *gin.Context) {
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

	err = h.deletePrayerTitleUseCase.Execute(c.Request.Context(), &application.DeletePrayerTitleRequest{
		PrayerTitleID: prayerID,
		MemberID:      memberID.(uint64),
	})
	if err != nil {
		if err == domain.ErrPrayerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "prayer not found"})
			return
		}
		if err == domain.ErrUnauthorizedAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete prayer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "기도를 삭제했습니다.", // Matching Java message
	})
}

// DeletePrayerContent handles DELETE /prayers/contents/:contentId
func (h *HandlerV2) DeletePrayerContent(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	contentIDStr := c.Param("contentId")
	contentID, err := strconv.ParseUint(contentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content ID"})
		return
	}

	err = h.deletePrayerContentUseCase.Execute(c.Request.Context(), &application.DeletePrayerContentRequest{
		ContentID: contentID,
		MemberID:  memberID.(uint64),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete prayer content"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "prayer content deleted successfully",
	})
}

// CompletePrayerV2Request represents the request to complete a prayer (matching Java)
type CompletePrayerV2Request struct {
	RoomID uint64 `json:"roomId" binding:"required"`
}

// CompletePrayer handles POST /prayers/:id/completion
func (h *HandlerV2) CompletePrayer(c *gin.Context) {
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

	var req CompletePrayerV2Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘 못된 기도제목 입니다."})
		return
	}

	err = h.completePrayerUseCase.Execute(c.Request.Context(), &application.CompletePrayerRequest{
		PrayerTitleID: prayerID,
		MemberID:      memberID.(uint64),
		RoomID:        req.RoomID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete prayer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "기도 완료 알림을 전송했습니다.", // Matching Java message
	})
}

// CompletePrayerContent handles POST /prayers/contents/:contentId/completion
func (h *HandlerV2) CompletePrayerContent(c *gin.Context) {
	memberID, exists := c.Get("memberID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	contentIDStr := c.Param("contentId")
	contentID, err := strconv.ParseUint(contentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content ID"})
		return
	}

	err = h.completePrayerUseCase.ExecuteContent(c.Request.Context(), &application.CompletePrayerContentRequest{
		ContentID: contentID,
		MemberID:  memberID.(uint64),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete prayer content"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "prayer content completed successfully",
	})
}
