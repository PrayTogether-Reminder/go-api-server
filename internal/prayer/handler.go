package prayer

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
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
	memberID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}

	var request PrayerTitleInfiniteScrollRequest
	if !sharedHttp.BindQuery(c, &request) {
		return
	}

	if request.After == "" {
		request.After = DefaultPrayerTitleAfter
	}

	response, err := h.prayerUseCase.FetchTitlesByInfiniteScroll(c.Request.Context(), memberID, &request)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			sharedHttp.RespondError(c, err, resp)
			return
		}

		sharedHttp.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *PrayerHandler) CreatePrayerTitle(c *gin.Context) {
	memberID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}
	var request CreatePrayerTitleRequest
	if !sharedHttp.BindJSON(c, &request) {
		return
	}

	response, err := h.prayerUseCase.CreatePrayerTitle(c.Request.Context(), memberID, &request)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			sharedHttp.RespondError(c, err, resp)
			return
		}

		sharedHttp.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *PrayerHandler) CreatePrayerContent(c *gin.Context) {
	writerID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}

	var uriParam struct {
		TitleID int64 `uri:"titleId" binding:"gt=0"`
	}
	if !sharedHttp.BindURI(c, &uriParam) {
		return
	}

	var request CreatePrayerContentRequest
	if !sharedHttp.BindJSON(c, &request) {
		return
	}

	response, err := h.prayerUseCase.CreatePrayerContent(c.Request.Context(), writerID, uriParam.TitleID, &request)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			sharedHttp.RespondError(c, err, resp)
			return
		}

		sharedHttp.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *PrayerHandler) FetchPrayerContents(c *gin.Context) {
	memberID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}

	var path TitleIDParam
	if !sharedHttp.BindURI(c, &path) {
		return
	}

	response, err := h.prayerUseCase.FetchPrayerContents(c.Request.Context(), memberID, path.TitleID)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			sharedHttp.RespondError(c, err, resp)
			return
		}

		sharedHttp.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *PrayerHandler) UpdatePrayerTitle(c *gin.Context) {
	memberID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}

	var path TitleIDParam
	if !sharedHttp.BindURI(c, &path) {
		return
	}

	var request UpdatePrayerTitleRequest
	if !sharedHttp.BindJSON(c, &request) {
		return
	}

	response, err := h.prayerUseCase.UpdatePrayerTitle(c.Request.Context(), memberID, path.TitleID, &request)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			sharedHttp.RespondError(c, err, resp)
			return
		}

		sharedHttp.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}
