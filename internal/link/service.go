package link

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/skskuzan/rd-linkforge/internal/base62"
)

// Service implements the Shortener interface.
type Service struct {
	store  Store
	nextID uint64
	now    func() time.Time
}

type Option func(*Service)

func WithClock(now func() time.Time) Option {
	return func(s *Service) {
		s.now = now
	}
}

// WithFirstID sets the ID handed out first. Values below 1 are ignored.
func WithFirstID(id uint64) Option {
	return func(s *Service) {
		if id >= 1 {
			s.nextID = id
		}
	}
}

// NewService creates a Shortener service.
func NewService(store Store, opts ...Option) *Service {
	s := &Service{
		store:  store,
		nextID: 1,
		now:    time.Now,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
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

	if !req.ExpiresAt.IsZero() && !req.ExpiresAt.After(s.now()) {
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
		case "api", "ws", "healthz", "readyz", "metrics", "admin":
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

	attempts := 1
	if req.Alias == "" {
		attempts = 3
	}

	for attempt := 0; attempt < attempts; attempt++ {
		id := s.nextID

		code := req.Alias
		if code == "" {
			code = base62.EncodeWidth(id, 6)
		}

		l := Link{
			ID:        id,
			Code:      code,
			Target:    req.Target,
			OwnerID:   req.OwnerID,
			CreatedAt: s.now().UTC(),
			ExpiresAt: req.ExpiresAt,
		}

		created, err := s.store.Create(ctx, l)
		if err == nil {
			s.nextID++
			return created, nil
		}

		if errors.Is(err, ErrAliasTaken) {
			if req.Alias != "" {
				return Link{}, &AliasError{
					Alias: req.Alias,
					Err:   ErrAliasTaken,
				}
			}

			// A generated-code collision consumes an ID.
			s.nextID++
			continue
		}

		return Link{}, fmt.Errorf("create link: %w", err)
	}

	return Link{}, fmt.Errorf("generated code collisions: %w", ErrAliasTaken)

}

func (s *Service) Resolve(ctx context.Context, code string) (Link, error) {
	if err := ctx.Err(); err != nil {
		return Link{}, err
	}

	l, err := s.store.ByCode(ctx, code)
	if err != nil {
		return Link{}, fmt.Errorf("resolve link: %w", err)
	}

	if !l.ExpiresAt.IsZero() && !s.now().Before(l.ExpiresAt) {
		return Link{}, ErrExpired
	}

	return l, nil
}

func (s *Service) List(ctx context.Context, ownerID string, page Page) ([]Link, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if strings.TrimSpace(ownerID) == "" {
		return nil, ErrEmptyOwner
	}

	links, err := s.store.ListByOwner(ctx, ownerID, page.Normalize())
	if err != nil {
		return nil, fmt.Errorf("list links: %w", err)
	}

	return links, nil
}

func (s *Service) Delete(ctx context.Context, ownerID, code string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if strings.TrimSpace(ownerID) == "" {
		return ErrEmptyOwner
	}

	l, err := s.store.ByCode(ctx, code)
	if err != nil {
		return fmt.Errorf("find link to delete: %w", err)
	}

	if l.OwnerID != ownerID {
		return ErrForbidden
	}

	if err := s.store.Delete(ctx, l.ID); err != nil {
		return fmt.Errorf("delete link: %w", err)
	}

	return nil
}

var _ Shortener = (*Service)(nil)
