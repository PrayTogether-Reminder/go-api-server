package prayer

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
)

const (
	prayerTitleCreateFailed   = "PRAYER_TITLE_CREATE_FAILED"   // errInfo
	prayerTitleNotFound       = "PRAYER_TITLE_NOT_FOUND"       // errInfo
	prayerContentCreateFailed = "PRAYER_CONTENT_CREATE_FAILED" // errInfo
	prayerContentNotFound     = "PRAYER_CONTENT_NOT_FOUND"     // errInfo
	prayerTitleInvalidCursor  = "PRAYER_TITLE_INVALID_CURSOR"  // errInfo
)

var (
	ErrPrayerTitleCreateFailed   = sharedError.NewDomainError(prayerTitleCreateFailed)
	ErrPrayerTitleNotFound       = sharedError.NewDomainError(prayerTitleNotFound)
	ErrPrayerContentCreateFailed = sharedError.NewDomainError(prayerContentCreateFailed)
	ErrPrayerContentNotFound     = sharedError.NewDomainError(prayerContentNotFound)
	ErrPrayerTitleInvalidCursor  = sharedError.NewDomainError(prayerTitleInvalidCursor)
)

func init() {
	// Register domain error responses
	sharedError.RegisterDomainErrorResponse(prayerTitleCreateFailed, sharedError.ErrorResponse{
		Status:  http.StatusInternalServerError,
		Code:    "PRAYER-004",
		Message: "기도제목 생성에 실패했습니다.",
	})

	sharedError.RegisterDomainErrorResponse(prayerTitleNotFound, sharedError.ErrorResponse{
		Status:  http.StatusNotFound,
		Code:    "PRAYER-001",
		Message: "기도 제목을 찾을 수 없습니다.",
	})

	sharedError.RegisterDomainErrorResponse(prayerContentCreateFailed, sharedError.ErrorResponse{
		Status:  http.StatusInternalServerError,
		Code:    "PRAYER-005",
		Message: "기도 내용 작성에 실패했습니다.",
	})

	sharedError.RegisterDomainErrorResponse(prayerContentNotFound, sharedError.ErrorResponse{
		Status:  http.StatusNotFound,
		Code:    "PRAYER-002",
		Message: "기도 내용을 찾을 수 없습니다.",
	})

	sharedError.RegisterDomainErrorResponse(prayerTitleInvalidCursor, sharedError.ErrorResponse{
		Status:  http.StatusBadRequest,
		Code:    "PRAYER-006",
		Message: "잘못된 커서 값입니다.",
	})
}
