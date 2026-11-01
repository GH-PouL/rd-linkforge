package link

import (
	"context"
	"errors"
	"strings"
	"time"
)

// Service implements the Shortener interface.
type Service struct {
	store Store
}

// NewService creates a Shortener service.
func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) Shorten(ctx context.Context, req ShortenRequest) (Link, error) {
	if err := ctx.Err(); err != nil {
		return Link{}, err
	}

	if strings.TrimSpace(req.OwnerID) == "" {
		return Link{}, ErrEmptyOwner
	}

	if _, err := New(0, req.Target); err != nil {
		return Link{}, err
	}

	if !req.ExpiresAt.IsZero() && !req.ExpiresAt.After(time.Now()) {
		return Link{}, ErrInvalidExpiry
	}

	if req.Alias != "" {
		if len(req.Alias) < 3 || len(req.Alias) > 32 {
			return Link{}, &AliasError{
				Alias: req.Alias,
				Err:   ErrInvalidAlias,
			}
		}

		for _, ch := range req.Alias {
			if (ch >= 'a' && ch <= 'z') ||
				(ch >= 'A' && ch <= 'Z') ||
				(ch >= '0' && ch <= '9') ||
				ch == '-' || ch == '_' {
				continue
			}

			return Link{}, &AliasError{
				Alias: req.Alias,
				Err:   ErrInvalidAlias,
			}
		}
		switch strings.ToLower(req.Alias) {
		case "api", "admin", "health":
			return Link{}, &AliasError{
				Alias: req.Alias,
				Err:   ErrReservedAlias,
			}
		}
	}
	if req.Alias != "" {
		_, err := s.store.ByCode(ctx, req.Alias)

		switch {
		case err == nil:
			return Link{}, &AliasError{
				Alias: req.Alias,
				Err:   ErrAliasTaken,
			}

		case errors.Is(err, ErrNotFound):
			// Код вільний — продовжуємо.

		default:
			return Link{}, err
		}
	}
	return Link{}, nil
}
