package prayer

import (
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"github.com/gin-gonic/gin"
	"net/http"
)

type PrayerHandler struct {
	prayerUseCase *PrayerUseCase
}

func NewPrayerHandler(prayerUseCase *PrayerUseCase) *PrayerHandler {
	return &PrayerHandler{
		prayerUseCase: prayerUseCase,
	}
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
