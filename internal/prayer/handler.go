package prayer

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/error"
	http2 "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/http"
	"github.com/gin-gonic/gin"
)

type PrayerHandler struct {
	prayerUseCase *PrayerUseCase
}

func NewPrayerHandler(prayerUseCase *PrayerUseCase) *PrayerHandler {
	return &PrayerHandler{
		prayerUseCase: prayerUseCase,
	}
}

func (h *PrayerHandler) FetchTitlesByInfiniteScroll(c *gin.Context) {
	memberID, ok := http2.RequireMemberID(c)
	if !ok {
		return
	}

	var request PrayerTitleInfiniteScrollRequest
	if !http2.BindQuery(c, &request) {
		return
	}

	if request.After == "" {
		request.After = DefaultPrayerTitleAfter
	}

	response, err := h.prayerUseCase.FetchTitlesByInfiniteScroll(c.Request.Context(), memberID, &request)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			http2.RespondError(c, err, resp)
			return
		}

		http2.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *PrayerHandler) CreatePrayerTitle(c *gin.Context) {
	memberID, ok := http2.RequireMemberID(c)
	if !ok {
		return
	}
	var request CreatePrayerTitleRequest
	if !http2.BindJSON(c, &request) {
		return
	}

	response, err := h.prayerUseCase.CreatePrayerTitle(c.Request.Context(), memberID, &request)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			http2.RespondError(c, err, resp)
			return
		}

		http2.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *PrayerHandler) CreatePrayerContent(c *gin.Context) {
	writerID, ok := http2.RequireMemberID(c)
	if !ok {
		return
	}

	var uriParam struct {
		TitleID int64 `uri:"titleId" binding:"gt=0"`
	}
	if !http2.BindURI(c, &uriParam) {
		return
	}

	var request CreatePrayerContentRequest
	if !http2.BindJSON(c, &request) {
		return
	}

	response, err := h.prayerUseCase.CreatePrayerContent(c.Request.Context(), writerID, uriParam.TitleID, &request)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			http2.RespondError(c, err, resp)
			return
		}

		http2.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *PrayerHandler) FetchPrayerContents(c *gin.Context) {
	memberID, ok := http2.RequireMemberID(c)
	if !ok {
		return
	}

	var path TitleIDParam
	if !http2.BindURI(c, &path) {
		return
	}

	response, err := h.prayerUseCase.FetchPrayerContents(c.Request.Context(), memberID, path.TitleID)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			http2.RespondError(c, err, resp)
			return
		}

		http2.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *PrayerHandler) UpdatePrayerTitle(c *gin.Context) {
	memberID, ok := http2.RequireMemberID(c)
	if !ok {
		return
	}

	var path TitleIDParam
	if !http2.BindURI(c, &path) {
		return
	}

	var request UpdatePrayerTitleRequest
	if !http2.BindJSON(c, &request) {
		return
	}

	response, err := h.prayerUseCase.UpdatePrayerTitle(c.Request.Context(), memberID, path.TitleID, &request)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			http2.RespondError(c, err, resp)
			return
		}

		http2.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *PrayerHandler) UpdatePrayerContent(c *gin.Context) {
	memberID, ok := http2.RequireMemberID(c)
	if !ok {
		return
	}

	var path TitleIDContentIDParam
	if !http2.BindURI(c, &path) {
		return
	}

	var request UpdatePrayerContentRequest
	if !http2.BindJSON(c, &request) {
		return
	}

	response, err := h.prayerUseCase.UpdatePrayerContent(c.Request.Context(), memberID, path.TitleID, path.ContentID, &request)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			http2.RespondError(c, err, resp)
			return
		}

		http2.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *PrayerHandler) DeletePrayerTitle(c *gin.Context) {
	memberID, ok := http2.RequireMemberID(c)
	if !ok {
		return
	}

	var path TitleIDParam
	if !http2.BindURI(c, &path) {
		return
	}

	response, err := h.prayerUseCase.DeletePrayerTitle(c.Request.Context(), memberID, path.TitleID)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			http2.RespondError(c, err, resp)
			return
		}

		http2.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *PrayerHandler) DeletePrayerContent(c *gin.Context) {
	memberID, ok := http2.RequireMemberID(c)
	if !ok {
		return
	}

	var path TitleIDContentIDParam
	if !http2.BindURI(c, &path) {
		return
	}

	response, err := h.prayerUseCase.DeletePrayerContent(c.Request.Context(), memberID, path.TitleID, path.ContentID)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			http2.RespondError(c, err, resp)
			return
		}

		http2.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}
