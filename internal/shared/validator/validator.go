package validator

import (
	"fmt"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/member"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"log/slog"
)

func Register() error {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return fmt.Errorf("validator 엔진을 가져올 수 없습니다")
	}

	// Member validators
	if err := v.RegisterValidation("phone", member.ValidatePhone); err != nil {
		return fmt.Errorf("Phone Validator 등록 실패: %w", err)
	}

	slog.Info("Custom Validator 등록 완료")
	return nil
}
