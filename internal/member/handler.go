package member

import (
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"github.com/gin-gonic/gin"
)

type MemberHandler struct {
	memberService *MemberService
}

func NewMemberHandler(memberService *MemberService) *MemberHandler {
	return &MemberHandler{
		memberService: memberService,
	}
}

func (h *MemberHandler) GetProfile(c *gin.Context) {
	MemberID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}

	response, err := h.memberService.GetProfile(c.Request.Context(), MemberID)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			sharedHttp.RespondError(c, err, resp)
			return
		}

		sharedHttp.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(200, response)
}
