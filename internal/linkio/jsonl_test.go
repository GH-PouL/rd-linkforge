package linkio

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/skskuzan/rd-linkforge/internal/link"
)

func TestLinksRoundTrip(t *testing.T) {
	now := time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)

	want := []link.Link{
		{
			ID:        1,
			Code:      "000001",
			Target:    "https://example.com",
			OwnerID:   "cli",
			CreatedAt: now,
		},
		{
			ID:        2,
			Code:      "000002",
			Target:    "https://go.dev",
			OwnerID:   "cli",
			CreatedAt: now,
			ExpiresAt: now.Add(24 * time.Hour),
		},
	}

	var buf bytes.Buffer

	if err := WriteLinks(&buf, want); err != nil {
		t.Fatalf("WriteLinks() error = %v", err)
	}

	if count := strings.Count(buf.String(), "\n"); count != len(want) {
		t.Errorf("JSONL lines = %d, want %d", count, len(want))
	}

	got, err := ReadLinks(&buf)
	if err != nil {
		t.Fatalf("ReadLinks() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadLinks() = %+v, want %+v", got, want)
	}
}

func TestReadLinksInvalidJSON(t *testing.T) {
	input := `{"ID":1,"Code":"000001","Target":"https://example.com"}
{invalid json}
{"ID":3,"Code":"000003","Target":"https://go.dev"}
`

	links, err := ReadLinks(strings.NewReader(input))

	if err == nil {
		t.Fatal("expected an error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "record 2:") {
		t.Errorf("error = %q, want record 2", err)
	}

	if links != nil {
		t.Errorf("links = %+v, want nil", links)
	}
}

func TestReadLinksEmpty(t *testing.T) {
	links, err := ReadLinks(strings.NewReader(""))

	if err != nil {
		t.Fatalf("ReadLinks() error = %v", err)
	}

	if len(links) != 0 {
		t.Errorf("links count = %d, want 0", len(links))
	}
}
