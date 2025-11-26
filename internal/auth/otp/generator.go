package otp

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// NumericGenerator creates zero-padded numeric OTP strings.
type NumericGenerator struct {
	digits int
}

func NewNumericGenerator(digits int) *NumericGenerator {
	if digits <= 0 {
		digits = 6
	}
	return &NumericGenerator{digits: digits}
}

func (g *NumericGenerator) Generate() (string, error) {
	max := big.NewInt(1)
	for i := 0; i < g.digits; i++ {
		max.Mul(max, big.NewInt(10))
	}
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("OTP 생성 실패: %w", err)
	}
	return fmt.Sprintf("%0*d", g.digits, value.Int64()), nil
}
