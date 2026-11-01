package store

import (
	"context"
	"errors"
	"testing"

	"github.com/skskuzan/rd-linkforge/internal/link"
)

func TestCreateAndByCode(t *testing.T) {
	s := New()
	ctx := context.Background()

	original := link.Link{
		ID:      1,
		Code:    "000001",
		Target:  "https://example.com",
		OwnerID: "user1",
	}

	created, err := s.Create(ctx, original)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created != original {
		t.Errorf("Create() = %+v, want %+v", created, original)
	}

	found, err := s.ByCode(ctx, original.Code)
	if err != nil {
		t.Fatalf("ByCode() error = %v", err)
	}

	if found != original {
		t.Errorf("ByCode() = %+v, want %+v", found, original)
	}

	_, err = s.ByCode(ctx, "missing")
	if !errors.Is(err, link.ErrNotFound) {
		t.Errorf("ByCode(missing) error = %v, want ErrNotFound", err)
	}
}

func TestCreateDuplicateID(t *testing.T) {
	s := New()
	ctx := context.Background()

	first := link.Link{
		ID:     1,
		Code:   "000001",
		Target: "https://example.com",
	}

	second := link.Link{
		ID:     1,
		Code:   "000002",
		Target: "https://example.org",
	}

	if _, err := s.Create(ctx, first); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	_, err := s.Create(ctx, second)
	if !errors.Is(err, ErrDuplicateID) {
		t.Errorf("duplicate Create() error = %v, want ErrDuplicateID", err)
	}

	if s.Count() != 1 {
		t.Errorf("Count() = %d, want 1", s.Count())
	}
}

func TestListByOwner(t *testing.T) {
	s := New()
	ctx := context.Background()

	links := []link.Link{
		{ID: 5, Code: "5", OwnerID: "user1"},
		{ID: 2, Code: "2", OwnerID: "user1"},
		{ID: 3, Code: "3", OwnerID: "user2"},
		{ID: 1, Code: "1", OwnerID: "user1"},
		{ID: 4, Code: "4", OwnerID: "user1"},
	}

	for _, l := range links {
		if _, err := s.Create(ctx, l); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Перша сторінка: два найменші ID власника user1.
	first, err := s.ListByOwner(ctx, "user1", link.Page{
		Limit: 2,
	})
	if err != nil {
		t.Fatalf("first page error = %v", err)
	}

	if len(first) != 2 || first[0].ID != 1 || first[1].ID != 2 {
		t.Fatalf("first page = %+v, want IDs [1 2]", first)
	}

	// Друга сторінка: записи після ID = 2.
	second, err := s.ListByOwner(ctx, "user1", link.Page{
		Cursor: "2",
		Limit:  2,
	})
	if err != nil {
		t.Fatalf("second page error = %v", err)
	}

	if len(second) != 2 || second[0].ID != 4 || second[1].ID != 5 {
		t.Fatalf("second page = %+v, want IDs [4 5]", second)
	}
}

func TestDelete(t *testing.T) {
	s := New()
	ctx := context.Background()

	original := link.Link{
		ID:     1,
		Code:   "000001",
		Target: "https://example.com",
	}

	// Створюємо посилання.
	if _, err := s.Create(ctx, original); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Видаляємо посилання.
	if err := s.Delete(ctx, original.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Перевіряємо, що сховище порожнє.
	if got := s.Count(); got != 0 {
		t.Errorf("Count() = %d, want 0", got)
	}

	// Перевіряємо, що посилання більше не знаходиться.
	_, err := s.ByCode(ctx, original.Code)
	if !errors.Is(err, link.ErrNotFound) {
		t.Errorf("ByCode() error = %v, want ErrNotFound", err)
	}

	// Повторне видалення повинно повернути ErrNotFound.
	err = s.Delete(ctx, original.ID)
	if !errors.Is(err, link.ErrNotFound) {
		t.Errorf("second Delete() error = %v, want ErrNotFound", err)
	}
}

func TestCanceledContext(t *testing.T) {
	s := New()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	l := link.Link{
		ID:     1,
		Code:   "1",
		Target: "https://example.com",
	}

	// Create
	_, err := s.Create(ctx, l)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Create() error = %v, want context.Canceled", err)
	}

	// ByCode
	_, err = s.ByCode(ctx, "1")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("ByCode() error = %v, want context.Canceled", err)
	}

	// ListByOwner
	_, err = s.ListByOwner(ctx, "user1", link.Page{})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("ListByOwner() error = %v, want context.Canceled", err)
	}

	// Delete
	err = s.Delete(ctx, 1)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Delete() error = %v, want context.Canceled", err)
	}

	// Скасований Create не повинен змінити сховище.
	if got := s.Count(); got != 0 {
		t.Errorf("Count() = %d, want 0", got)
	}
}
