package base62

import (
	"errors"
	"math"
	"strings"
)

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var (
	ErrEmpty            = errors.New("base62: empty code")
	ErrInvalidCharacter = errors.New("base62: invalid character")
	ErrOverflow         = errors.New("base62: code is too long")
)

// Encode renders a number with no padding.
func Encode(v uint64) string {
	if v == 0 {
		return "0"
	}

	var buf [11]byte
	i := len(buf)

	for v > 0 {
		i--
		buf[i] = alphabet[v%62]
		v /= 62
	}

	return string(buf[i:])
}

// EncodeWidth encodes v and left-pads it with zeroes to at least width characters.
// The encoded value is never truncated.
func EncodeWidth(v uint64, width int) string {
	code := Encode(v)
	if len(code) >= width {
		return code
	}

	return strings.Repeat("0", width-len(code)) + code
}

// Decode parses a Base62 code into a uint64.
func Decode(s string) (uint64, error) {
	if s == "" {
		return 0, ErrEmpty
	}

	if len(s) > 11 {
		return 0, ErrOverflow
	}

	var value uint64

	for i := 0; i < len(s); i++ {
		digit := strings.IndexByte(alphabet, s[i])
		if digit < 0 {
			return 0, ErrInvalidCharacter
		}

		d := uint64(digit)
		if value > (math.MaxUint64-d)/62 {
			return 0, ErrOverflow
		}

		value = value*62 + d
	}

	return value, nil
}
