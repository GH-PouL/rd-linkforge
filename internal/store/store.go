package store

import (
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
func New() Store {
	return Store{
		byID: make(map[uint64]link.Link),
	}
}

// Add adds a link unless its ID is already taken.
func Add(s Store, l link.Link) error {
	if _, exists := s.byID[l.ID]; exists {
		return ErrDuplicateID
	}

	s.byID[l.ID] = l
	return nil
}

// Get returns a link and reports whether it was found.
func Get(s Store, id uint64) (link.Link, bool) {
	l, ok := s.byID[id]
	return l, ok
}

// All returns links sorted by ID in ascending order.
func All(s Store) []link.Link {
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
func Count(s Store) int {
	return len(s.byID)
}
