package fcm_token

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/http"
	"github.com/gin-gonic/gin"
)

// Handler exposes HTTP endpoints for managing FCM tokens.
type Handler struct {
	useCase *FcmTokenUseCase
}

// NewHandler constructs a new handler.
func NewHandler(useCase *FcmTokenUseCase) *Handler {
	return &Handler{useCase: useCase}
}

// RegisterToken handles POST /api/v1/fcm-token requests.
func (h *Handler) RegisterToken(c *gin.Context) {
	memberID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}

	var request RegisterFcmTokenRequest
	if !sharedHttp.BindJSON(c, &request) {
		return
	}

	token := request.TrimmedToken()

	if err := h.useCase.RegisterFcmToken(c.Request.Context(), memberID, token); err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			sharedHttp.RespondError(c, err, resp)
			return
		}

		sharedHttp.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

// DeleteToken handles DELETE /api/v1/fcm-token requests.
func (h *Handler) DeleteToken(c *gin.Context) {
	memberID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}

	var request DeleteFcmTokenRequest
	if !sharedHttp.BindJSON(c, &request) {
		return
	}

	token := request.TrimmedToken()

	if err := h.useCase.DeleteFcmToken(c.Request.Context(), memberID, token); err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			sharedHttp.RespondError(c, err, resp)
			return
		}

		sharedHttp.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
