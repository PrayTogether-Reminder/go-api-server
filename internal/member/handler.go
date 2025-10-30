package member

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/validator"
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

	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(err)
		if resp, ok := validator.ToErrorResponse(err); ok {
			c.JSON(http.StatusBadRequest, resp)
		} else {
			c.JSON(http.StatusBadRequest, sharedError.InvalidRequest.Response(""))
		}
		return
	}

	// TODO: 비즈니스 로직 구현
	// err := m.memberService.Signup(c.Request.Context(), &request)
	// if err != nil {
	//     c.Error(err)
	//     c.JSON(http.StatusInternalServerError, sharedError.InternalServerError.Response(""))
	//     return
	// }
	// c.JSON(http.StatusCreated, gin.H{"message": "회원가입이 완료되었습니다"})
}
