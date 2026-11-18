package cache

import (
	"errors"
	"slices"
	"testing"
)

func TestNewLRU(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		wantErr  bool
	}{
		{"positive", 3, false},
		{"zero", 0, true},
		{"negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewLRU[string, int](tt.capacity)

			if tt.wantErr {
				if !errors.Is(err, ErrInvalidCapacity) {
					t.Errorf("expected ErrInvalidCapacity, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if c.Len() != 0 {
				t.Errorf("new cache length = %d, want 0", c.Len())
			}
		})
	}
}

func TestLRUEviction(t *testing.T) {
	c, err := NewLRU[string, int](3)
	if err != nil {
		t.Fatal(err)
	}

	c.Put("A", 10)
	c.Put("B", 20)
	c.Put("C", 30)

	// A becomes the most recently used.
	value, ok := c.Get("A")
	if !ok || value != 10 {
		t.Errorf("Get(A) = %d, %v; want 10, true", value, ok)
	}

	// The cache is full, so B must be evicted.
	c.Put("D", 40)

	if _, ok := c.Get("B"); ok {
		t.Error("B should have been evicted")
	}

	if c.Len() != 3 {
		t.Errorf("Len() = %d; want 3", c.Len())
	}
}

func TestLRUOperations(t *testing.T) {
	c, err := NewLRU[string, int](3)
	if err != nil {
		t.Fatal(err)
	}

	c.Put("A", 10)
	c.Put("B", 20)
	c.Put("C", 30)

	// Update existing key.
	c.Put("A", 100)

	if value, ok := c.Get("A"); !ok || value != 100 {
		t.Errorf("Get(A) = %d, %v; want 100, true", value, ok)
	}

	want := []string{"A", "C", "B"}
	if got := c.Keys(); !slices.Equal(got, want) {
		t.Errorf("Keys() = %v; want %v", got, want)
	}

	c.Delete("C")
	c.Delete("missing")

	if _, ok := c.Get("C"); ok {
		t.Error("C should have been deleted")
	}

	if c.Len() != 2 {
		t.Errorf("Len() = %d; want 2", c.Len())
	}
}
