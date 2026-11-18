package topk

import (
	"errors"
	"slices"
	"testing"
)

func TestNew(t *testing.T) {
	for _, n := range []int{0, -1} {
		if _, err := New(n); !errors.Is(err, ErrInvalidN) {
			t.Errorf("New(%d): expected ErrInvalidN, got %v", n, err)
		}
	}

	if _, err := New(3); err != nil {
		t.Fatalf("New(3): unexpected error: %v", err)
	}
}

func TestTopK(t *testing.T) {
	top, err := New(3)
	if err != nil {
		t.Fatal(err)
	}

	counts := []int64{5, 1, 9, 3, 9, 7, 2}

	for i, count := range counts {
		key := "k" + string(rune('0'+i))
		top.Add(key, count)
	}

	want := []Item{
		{Key: "k2", Count: 9},
		{Key: "k4", Count: 9},
		{Key: "k5", Count: 7},
	}

	if got := top.Result(); !slices.Equal(got, want) {
		t.Errorf("Result() = %v, want %v", got, want)
	}
}

func TestTopKTieBreak(t *testing.T) {
	top, err := New(2)
	if err != nil {
		t.Fatal(err)
	}

	top.Add("z", 10)
	top.Add("b", 10)
	top.Add("a", 10)

	want := []Item{
		{Key: "a", Count: 10},
		{Key: "b", Count: 10},
	}

	if got := top.Result(); !slices.Equal(got, want) {
		t.Errorf("Result() = %v, want %v", got, want)
	}
}

func TestTopKResultDoesNotModifyHeap(t *testing.T) {
	top, err := New(2)
	if err != nil {
		t.Fatal(err)
	}

	top.Add("a", 5)
	top.Add("b", 10)

	first := top.Result()

	top.Add("c", 20)

	want := []Item{
		{Key: "c", Count: 20},
		{Key: "b", Count: 10},
	}

	if got := top.Result(); !slices.Equal(got, want) {
		t.Errorf("Result() = %v, want %v", got, want)
	}

	if !slices.Equal(first, []Item{
		{Key: "b", Count: 10},
		{Key: "a", Count: 5},
	}) {
		t.Errorf("previous result changed: %v", first)
	}
}
