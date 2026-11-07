package link

import (
	"errors"
	"testing"

	"github.com/skskuzan/rd-linkforge/internal/base62"
)

func TestPageNormalizeCustom(t *testing.T) {
	tests := []struct {
		name  string
		limit int
		want  int
	}{
		{"zero", 0, 20},
		{"negative", -5, 20},
		{"normal", 50, 50},
		{"maximum", 100, 100},
		{"too large", 150, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := Page{Cursor: "ABC", Limit: tt.limit}
			got := page.Normalize()

			if got.Limit != tt.want {
				t.Errorf("Normalize().Limit = %d, want %d",
					got.Limit, tt.want)
			}

			if got.Cursor != page.Cursor {
				t.Errorf("Normalize() changed Cursor")
			}

			if page.Limit != tt.limit {
				t.Errorf("Normalize() modified original Page")
			}
		})
	}
}

func TestPageAfterIDCustom(t *testing.T) {
	tests := []struct {
		name    string
		cursor  string
		want    uint64
		wantErr error
	}{
		{"empty", "", 0, nil},
		{"zero", "0", 0, nil},
		{"single digit", "A", 10, nil},
		{"multiple digits", "K", 20, nil},
		{"two digits", "10", 62, nil},
		{"invalid character", "!", 0, base62.ErrInvalidCharacter},
		{"overflow", "zzzzzzzzzzzz", 0, base62.ErrOverflow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := Page{Cursor: tt.cursor}

			got, err := page.AfterID()

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("AfterID() error = %v, want %v",
					err, tt.wantErr)
			}

			// Invalid cursors must be wrapped with ErrInvalidCursor.
			if tt.wantErr != nil && !errors.Is(err, ErrInvalidCursor) {
				t.Errorf(
					"AfterID() error = %v, want ErrInvalidCursor",
					err,
				)
			}

			if err == nil && got != tt.want {
				t.Errorf("AfterID() = %d, want %d",
					got, tt.want)
			}
		})
	}
}
