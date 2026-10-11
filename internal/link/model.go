package link

import (
	"errors"
	"strings"
	"time"

	"github.com/skskuzan/rd-linkforge/internal/base62"
)

var (
	ErrEmptyTarget       = errors.New("link: empty target")
	ErrUnsupportedScheme = errors.New("link: unsupported URL scheme")
)

// Link is a shortened URL.
type Link struct {
	ID        uint64
	Code      string
	Target    string
	CreatedAt time.Time
}

// ShortenRequest carries everything needed to create a link.
type ShortenRequest struct{}

// Page describes one slice of a cursor-paginated listing.
type Page struct{}

// New builds a link, validating the target address.
func New(id uint64, target string) (Link, error) {
	if target == "" {
		return Link{}, ErrEmptyTarget
	}

	if !strings.HasPrefix(target, "http://") &&
		!strings.HasPrefix(target, "https://") {
		return Link{}, ErrUnsupportedScheme
	}

	return Link{
		ID:        id,
		Code:      base62.EncodeWidth(id, 6),
		Target:    target,
		CreatedAt: time.Now().UTC(),
	}, nil
}
