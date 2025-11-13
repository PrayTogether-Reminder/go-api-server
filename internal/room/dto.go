package room

// Constants for infinite scroll
const (
	DefaultOrderBy         = "time"
	DefaultAfter           = "0"
	DefaultDir             = "desc"
	InfiniteScrollPageSize = 10
)

// InfiniteScrollRequest represents the request for infinite scroll pagination
type InfiniteScrollRequest struct {
	OrderBy string `form:"orderBy" binding:"omitempty,oneof=time name memberCnt"` // 정렬 기준
	After   string `form:"after" binding:"required"`                              // 커서 (마지막 조회 항목)
	Dir     string `form:"dir" binding:"omitempty,oneof=asc desc"`                // 정렬 방향
}

// InfiniteScrollResponse represents the response for infinite scroll
type InfiniteScrollResponse struct {
	Rooms []RoomInfo `json:"rooms"` // 방 목록
}

// CreateRoomRequest represents the request for creating a new room
type CreateRoomRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Description string `json:"description" binding:"required,min=1,max=200"`
}
