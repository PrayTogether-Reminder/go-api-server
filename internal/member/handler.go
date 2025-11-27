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

// UpdateProfile - 회원 프로필 수정 API
func (h *MemberHandler) UpdateProfile(c *gin.Context) {
	memberID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}

	var request UpdateProfileRequest
	if !sharedHttp.BindJSON(c, &request) {
		return
	}

	response, err := h.memberUseCase.UpdateProfile(c.Request.Context(), memberID, &request)
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

// SearchMembers - 회원 검색 API
func (h *MemberHandler) SearchMembers(c *gin.Context) {
	memberID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}

	var request SearchMemberRequest
	if !sharedHttp.BindQuery(c, &request) {
		return
	}

	response, err := h.memberUseCase.SearchMembers(c.Request.Context(), memberID, request.Name)
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
