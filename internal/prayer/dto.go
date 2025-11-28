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

type PrayerContentInfo struct {
	ID         int64  `json:"id"`
	WriterID   int64  `json:"writerId"`
	WriterName string `json:"writerName"`
	MemberID   *int64 `json:"memberId,omitempty"`
	MemberName string `json:"memberName"`
	Content    string `json:"content"`
}

type PrayerContentResponse struct {
	PrayerContents []PrayerContentInfo `json:"prayerContents"`
}

type UpdatePrayerTitleRequest struct {
	ChangedTitle string `json:"changedTitle" binding:"required,min=1,max=50"`
}

type UpdatePrayerContentRequest struct {
	ChangedContent string `json:"changedContent" binding:"required"`
}

type TitleIDParam struct {
	TitleID int64 `uri:"titleId" binding:"gt=0"`
}

type TitleIDContentIDParam struct {
	TitleID   int64 `uri:"titleId" binding:"gt=0"`
	ContentID int64 `uri:"contentId" binding:"gt=0"`
}

// PrayerCompletionCreateRequest represents the request body for completing a prayer
type PrayerCompletionCreateRequest struct {
	RoomID int64 `json:"roomId" binding:"required,gt=0"`
}
