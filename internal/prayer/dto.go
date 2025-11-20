package prayer

import "time"

type CreatePrayerTitleRequest struct {
	RoomID int64  `json:"roomId" binding:"gt"`
	Title  string `json:"title" binding:"required"`
}

type CreatePrayerTitleResponse struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	CreatedTime time.Time `json:"createdTime"`
}
