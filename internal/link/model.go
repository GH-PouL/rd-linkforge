package link

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/skskuzan/rd-linkforge/internal/base62"
)

var (
	ErrEmptyTarget       = errors.New("link: empty target")
	ErrUnsupportedScheme = errors.New("link: unsupported URL scheme")
	ErrNotFound          = errors.New("link: not found")
	ErrAliasTaken        = errors.New("link: alias already taken")
	ErrInvalidAlias      = errors.New("link: invalid alias")
	ErrReservedAlias     = errors.New("link: reserved alias")
	ErrForbidden         = errors.New("link: forbidden")
	ErrExpired           = errors.New("link: expired")
	ErrInvalidExpiry     = errors.New("link: invalid expiry")
	ErrEmptyOwner        = errors.New("link: owner is required")
	ErrInvalidCursor     = errors.New("link: invalid cursor")
)

// Link is a shortened URL.
type Link struct {
	ID        uint64
	Code      string
	Target    string
	OwnerID   string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// ShortenRequest carries everything needed to create a link.
type ShortenRequest struct {
	Target    string
	Alias     string
	OwnerID   string
	ExpiresAt time.Time
}

// Page describes one slice of a cursor-paginated listing.
type Page struct {
	Cursor string
	Limit  int
}

const (
	DefaultPageLimit = 20
	MaxPageLimit     = 100
)

// Normalize returns a copy of Page with a valid Limit.
func (p Page) Normalize() Page {
	if p.Limit <= 0 {
		p.Limit = DefaultPageLimit
	} else if p.Limit > MaxPageLimit {
		p.Limit = MaxPageLimit
	}

	return p
}

// AfterID decodes the cursor into a numeric ID.
func (p Page) AfterID() (uint64, error) {
	if p.Cursor == "" {
		return 0, nil
	}

	id, err := base62.Decode(p.Cursor)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidCursor, err)
	}

	return id, nil
}

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
