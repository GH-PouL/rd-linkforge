package store

import (
	"context"
	"errors"
	"sort"

	"github.com/skskuzan/rd-linkforge/internal/link"
)

var ErrDuplicateID = errors.New("store: duplicate link ID")

// Store keeps links in memory, indexed by ID.
type Store struct {
	byID map[uint64]link.Link
}

// New creates an empty store.
func New() *Store {
	return &Store{
		byID: make(map[uint64]link.Link),
	}
}

// Add adds a link unless its ID is already taken.
func (s *Store) Add(l link.Link) error {
	if _, exists := s.byID[l.ID]; exists {
		return ErrDuplicateID
	}

	s.byID[l.ID] = l
	return nil
}

// Create persists a link and returns the stored value.
func (s *Store) Create(ctx context.Context, l link.Link) (link.Link, error) {
	if _, exists := s.byID[l.ID]; exists {
		return link.Link{}, ErrDuplicateID
	}

	for _, existing := range s.byID {
		if existing.Code == l.Code {
			return link.Link{}, link.ErrAliasTaken
		}
	}

	s.byID[l.ID] = l
	return l, nil
}

// ByCode returns a link by its short code.
func (s *Store) ByCode(ctx context.Context, code string) (link.Link, error) {
	if err := ctx.Err(); err != nil {
		return link.Link{}, err
	}

	for _, l := range s.byID {
		if l.Code == code {
			return l, nil
		}
	}

	return link.Link{}, link.ErrNotFound
}

// ListByOwner returns a page of links belonging to the specified owner.
func (s *Store) ListByOwner(
	ctx context.Context,
	ownerID string,
	page link.Page,
) ([]link.Link, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	page = page.Normalize()

	afterID, err := page.AfterID()
	if err != nil {
		return nil, err
	}

	links := make([]link.Link, 0)

	for _, l := range s.byID {
		if l.OwnerID == ownerID && l.ID > afterID {
			links = append(links, l)
		}
	}

	sort.Slice(links, func(i, j int) bool {
		return links[i].ID < links[j].ID
	})

	if len(links) > page.Limit {
		links = links[:page.Limit]
	}

	return links, nil
}

// Delete removes a link by its ID.
func (s *Store) Delete(ctx context.Context, id uint64) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if _, exists := s.byID[id]; !exists {
		return link.ErrNotFound
	}

	delete(s.byID, id)
	return nil
}

var _ link.Store = (*Store)(nil)

// Get returns a link and reports whether it was found.
func (s *Store) Get(id uint64) (link.Link, bool) {
	l, ok := s.byID[id]
	return l, ok
}

// All returns links sorted by ID in ascending order.
func (s *Store) All() []link.Link {
	links := make([]link.Link, 0, len(s.byID))

	for _, l := range s.byID {
		links = append(links, l)
	}

	sort.Slice(links, func(i, j int) bool {
		return links[i].ID < links[j].ID
	})

	return links
}

// Count returns the number of stored links.
func (s *Store) Count() int {
	return len(s.byID)
}
