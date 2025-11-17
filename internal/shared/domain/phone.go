package domain

import (
	"fmt"
	"regexp"
	"strings"
)

type PhoneNumber struct {
	value string
}

func NewPhoneNumber(phoneNumber string) (PhoneNumber, error) {
	if phoneNumber == "" || strings.TrimSpace(phoneNumber) == "" {
		return PhoneNumber{}, fmt.Errorf("전화번호를 입력해 주세요")
	}

	if err := validate(phoneNumber); err != nil {
		return PhoneNumber{}, err
	}

	normalized := normalize(phoneNumber)
	return PhoneNumber{value: normalized}, nil
}

// validate checks if the phone number is a valid Korean mobile number
func validate(phoneNumber string) error {
	// 하이픈 제거하고 검증
	digitsOnly := strings.ReplaceAll(phoneNumber, "-", "")

	// 11자리 개인 휴대폰 번호 검증 (010, 011, 016, 017, 018, 019)
	matched, _ := regexp.MatchString(`^01[016789]\d{8}$`, digitsOnly)
	if !matched {
		return fmt.Errorf("올바른 휴대폰 번호 형식이 아닙니다. (010, 011, 016, 017, 018, 019만 가능)")
	}

	return nil
}

// normalize formats the phone number to XXX-XXXX-XXXX format
func normalize(phoneNumber string) string {
	// 하이픈 제거하고 숫자만 추출
	digitsOnly := strings.ReplaceAll(phoneNumber, "-", "")

	// 11자리 개인 휴대폰(010, 011, 016, 017, 018, 019): XXX-XXXX-XXXX 형식
	matched, _ := regexp.MatchString(`^01[016789]`, digitsOnly)
	if len(digitsOnly) == 11 && matched {
		return digitsOnly[0:3] + "-" + digitsOnly[3:7] + "-" + digitsOnly[7:11]
	}

	// 그 외는 그대로 반환 (여기 도달하면 안 됨, validate에서 걸림)
	return phoneNumber
}

func (p PhoneNumber) GetSuffix() string {
	if p.value == "" {
		return ""
	}

	return p.value[len(p.value)-4:]
}

func (p PhoneNumber) String() string {
	return p.value
}
