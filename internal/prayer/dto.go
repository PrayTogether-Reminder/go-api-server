package prayer

import "time"

const (
	DefaultPrayerTitleAfter        = "0"
	PrayerTitleInfiniteScrollLimit = 10
)

type CreatePrayerTitleRequest struct {
	RoomID int64  `json:"roomId" binding:"gt=0"`
	Title  string `json:"title" binding:"required"`
}

type CreatePrayerTitleResponse struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	CreatedTime time.Time `json:"createdTime"`
}

type PrayerTitleInfiniteScrollRequest struct {
	RoomID int64  `form:"roomId" binding:"gt=0"`
	After  string `form:"after" binding:"omitempty"`
}

type PrayerTitleInfo struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	CreatedTime time.Time `json:"createdTime"`
}

type PrayerTitleInfiniteScrollResponse struct {
	PrayerTitles []PrayerTitleInfo `json:"prayerTitles"`
}

type CreatePrayerContentRequest struct {
	MemberID   *int64 `json:"memberId" binding:"omitempty,gt=0"`
	MemberName string `json:"memberName" binding:"required"`
	Content    string `json:"content" binding:"required"`
}
