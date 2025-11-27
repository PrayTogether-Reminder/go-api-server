package invitation

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/error"
	http2 "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/http"
	"github.com/gin-gonic/gin"
)

type InvitationHandler struct {
	invitationUseCase *InvitationUseCase
}

func NewInvitationHandler(invitationUseCase *InvitationUseCase) *InvitationHandler {
	return &InvitationHandler{
		invitationUseCase: invitationUseCase,
	}
}

// InviteMembers handles POST /api/v2/invitations - 회원 초대 (V2)
func (h *InvitationHandler) InviteMembers(c *gin.Context) {
	memberID, ok := http2.RequireMemberID(c)
	if !ok {
		return
	}

	var request InviteMembersRequest
	if !http2.BindJSON(c, &request) {
		return
	}

	response, err := h.invitationUseCase.InviteMembersToRoom(c.Request.Context(), memberID, &request)
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

// GetInvitations handles GET /api/v1/invitations - 내가 받은 초대 목록 조회
func (h *InvitationHandler) GetInvitations(c *gin.Context) {
	memberID, ok := http2.RequireMemberID(c)
	if !ok {
		return
	}

	response, err := h.invitationUseCase.GetInvitationInfoScroll(c.Request.Context(), memberID)
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

// RespondToInvitation handles PATCH /api/v1/invitations/:invitationId - 초대 응답
func (h *InvitationHandler) RespondToInvitation(c *gin.Context) {
	memberID, ok := http2.RequireMemberID(c)
	if !ok {
		return
	}

	var path InvitationIDParam
	if !http2.BindURI(c, &path) {
		return
	}

	var request InvitationStatusUpdateRequest
	if !http2.BindJSON(c, &request) {
		return
	}

	response, err := h.invitationUseCase.UpdateInvitationStatus(c.Request.Context(), memberID, path.InvitationID, &request)
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
