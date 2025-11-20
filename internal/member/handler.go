package member

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"github.com/gin-gonic/gin"
)

// MemberHandler - Member MemberHandler 구조체
type MemberHandler struct {
	memberUseCase *MemberUseCase
}

// NewMemberHandler - MemberHandler 생성자
func NewMemberHandler(memberUseCase *MemberUseCase) *MemberHandler {
	return &MemberHandler{
		memberUseCase: memberUseCase,
	}
}

// FetchProfile - 회원 프로필 조회 API
func (h *MemberHandler) FetchProfile(c *gin.Context) {
	memberID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}

	// UseCase를 통해 조회
	member, err := h.memberUseCase.GetProfile(c.Request.Context(), memberID)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			sharedHttp.RespondError(c, err, resp)
			return
		}

		sharedHttp.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	response := &FetchProfileResponse{
		ID:          memberID,
		Name:        member.Name,
		Email:       member.Email,
		PhoneNumber: member.PhoneNumber,
	}
	c.JSON(http.StatusOK, response)
}
