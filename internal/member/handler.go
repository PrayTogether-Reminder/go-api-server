package member

import (
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

	// TODO: 비즈니스 로직 구현
	// err := m.memberService.Signup(c.Request.Context(), &request)
	// if err != nil {
	//     handler.RespondError(c, err, sharedError.InternalServerError)
	//     return
	// }
	// handler.RespondJSON(c, 201, gin.H{"message": "회원가입이 완료되었습니다"})
}
