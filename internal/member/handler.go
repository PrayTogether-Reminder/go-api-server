package member

import (
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/handler"
	"github.com/gin-gonic/gin"
)

type MemberHandler struct {
	memberService MemberService
}

func NewMemberHandler(memberService MemberService) *MemberHandler {
	return &MemberHandler{
		memberService: memberService,
	}
}

func (m *MemberHandler) Signup(c *gin.Context) {
	var request SignupRequest

	// Parse and validate JSON request
	if !handler.BindJSON(c, &request) {
		return
	}

	err := m.memberService.Signup(c.Request.Context(), &request)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			handler.RespondError(c, err, resp)
			return
		}

		handler.RespondError(c, err, sharedError.InternalServerError)
		return
	}
	c.JSON(201, gin.H{})
}
