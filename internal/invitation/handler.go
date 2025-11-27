package invitation

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
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
	memberID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}

	var request InviteMembersRequest
	if !sharedHttp.BindJSON(c, &request) {
		return
	}

	response, err := h.invitationUseCase.InviteMembersToRoom(c.Request.Context(), memberID, &request)
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

// GetInvitations handles GET /api/v1/invitations - 내가 받은 초대 목록 조회
func (h *InvitationHandler) GetInvitations(c *gin.Context) {
	memberID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}

	response, err := h.invitationUseCase.GetInvitationInfoScroll(c.Request.Context(), memberID)
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

// RespondToInvitation handles PATCH /api/v1/invitations/:invitationId - 초대 응답
//func (h *InvitationHandler) RespondToInvitation(c *gin.Context) {
//	memberID, ok := sharedHttp.RequireMemberID(c)
//	if !ok {
//		return
//	}
//
//	// Parse invitation ID from URL
//	invitationIDStr := c.Param("invitationId")
//	invitationID, err := strconv.ParseInt(invitationIDStr, 10, 64)
//	if err != nil || invitationID <= 0 {
//		sharedHttp.RespondError(c, err, sharedError.ErrorResponse{
//			Status:  http.StatusBadRequest,
//			Code:    "INVITATION-003",
//			Message: "잘 못된 초대장 입니다.",
//		})
//		return
//	}
//
//	var request InvitationStatusUpdateRequest
//	if !sharedHttp.BindJSON(c, &request) {
//		return
//	}
//
//	response, err := h.invitationUseCase.UpdateInvitationStatus(c.Request.Context(), memberID, invitationID, &request)
//	if err != nil {
//		if resp, ok := sharedError.ResolveDomainError(err); ok {
//			sharedHttp.RespondError(c, err, resp)
//			return
//		}
//
//		sharedHttp.RespondError(c, err, sharedError.InternalServerError)
//		return
//	}
//
//	c.JSON(http.StatusOK, response)
//}
