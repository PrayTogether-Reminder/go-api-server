package room

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"github.com/gin-gonic/gin"
)

// RoomHandler handles HTTP requests for room operations
type RoomHandler struct {
	roomUseCase *RoomUseCase
}

// NewRoomHandler creates a new RoomHandler instance
func NewRoomHandler(roomUseCase *RoomUseCase) *RoomHandler {
	return &RoomHandler{
		roomUseCase: roomUseCase,
	}
}

// GetRoomsByInfiniteScroll handles GET /api/v1/rooms
func (h *RoomHandler) GetRoomsByInfiniteScroll(c *gin.Context) {
	memberID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}

	var request InfiniteScrollRequest

	// Set default values for optional parameters
	request.OrderBy = c.DefaultQuery("orderBy", DefaultOrderBy)
	request.After = c.DefaultQuery("after", DefaultAfter)
	request.Dir = c.DefaultQuery("dir", DefaultDir)

	if !sharedHttp.BindQuery(c, &request) {
		return
	}

	response, err := h.roomUseCase.FetchInfiniteScroll(c.Request.Context(), memberID, &request)
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

// CreateRoom handles POST /api/v1/rooms
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	memberID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}

	var request CreateRoomRequest
	if !sharedHttp.BindJSON(c, &request) {
		return
	}

	response, err := h.roomUseCase.CreateRoom(c.Request.Context(), memberID, &request)
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

// DeleteMemberRoom handles DELETE /api/v1/rooms/:roomId
func (h *RoomHandler) DeleteMemberRoom(c *gin.Context) {
	memberID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}

	var request ExitRoomRequest
	if !sharedHttp.BindURI(c, &request) {
		return
	}

	response, err := h.roomUseCase.ExitRoom(c.Request.Context(), memberID, request.RoomID)
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

func (h *RoomHandler) FetchRoomMembers(c *gin.Context) {
	memberID, ok := sharedHttp.RequireMemberID(c)
	if !ok {
		return
	}

	var request FetchRoomMemberRequest
	if !sharedHttp.BindURI(c, &request) {
		return
	}

	response, err := h.roomUseCase.FetchMembersInRoom(c.Request.Context(), memberID, request.RoomID)
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
