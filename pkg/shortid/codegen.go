package shortid

import (
	"fmt"
	"io"
)

// CodeGenerator generates six-character Base62 codes.
// It is not safe for concurrent use.
type CodeGenerator struct {
	random  io.Reader
	counter uint16
}

func NewCodeGenerator(random io.Reader) *CodeGenerator {
	return &CodeGenerator{
		random: random,
	}
}

var _ Generator = (*CodeGenerator)(nil)

func (g *CodeGenerator) Next() (string, error) {
	var buf [3]byte

	if _, err := io.ReadFull(g.random, buf[:]); err != nil {
		return "", fmt.Errorf("shortid: read random bytes: %w", err)
	}

	r := uint64(buf[0])<<16 |
		uint64(buf[1])<<8 |
		uint64(buf[2])

	r &= (1 << 19) - 1

	value := (r << 16) | uint64(g.counter)

	code := encodeBase62(value)

	g.counter++
	return code, nil
}

func encodeBase62(value uint64) string {
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	var code [6]byte

	for i := len(code) - 1; i >= 0; i-- {
		code[i] = alphabet[value%62]
		value /= 62
	}

	return string(code[:])
}
