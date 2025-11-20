package prayer

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
)

const (
	prayerTitleCreateFailed = "PRAYER_TITLE_CREATE_FAILED" // errInfo
)

var (
	ErrPrayerTitleCreateFailed = sharedError.NewDomainError(prayerTitleCreateFailed)
)

func init() {
	// Register domain error responses
	sharedError.RegisterDomainErrorResponse(prayerTitleCreateFailed, sharedError.ErrorResponse{
		Status:  http.StatusInternalServerError,
		Code:    "PRAYER-004",
		Message: "기도제목 생성에 실패했습니다.",
	})
}
