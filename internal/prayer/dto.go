package prayer

import "time"

type CreatePrayerTitleRequest struct {
	RoomID int64  `json:"roomId" binding:"gt=0"`
	Title  string `json:"title" binding:"required"`
}

type CreatePrayerTitleResponse struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	CreatedTime time.Time `json:"createdTime"`
}

type CreatePrayerContentRequest struct {
	MemberID   *int64 `json:"memberId" binding:"omitempty,gt=0"`
	MemberName string `json:"memberName" binding:"required"`
	Content    string `json:"content" binding:"required"`
}
