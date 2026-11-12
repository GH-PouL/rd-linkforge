package linkio

import (
	"bytes"
	"testing"
	"time"

	"github.com/skskuzan/rd-linkforge/internal/link"
)

func TestWriteCSV(t *testing.T) {
	created := time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)

	links := []link.Link{
		{
			ID:        1,
			Code:      "000001",
			Target:    "https://example.com",
			OwnerID:   "cli",
			CreatedAt: created,
		},
		{
			ID:        2,
			Code:      "000002",
			Target:    "https://go.dev",
			OwnerID:   "cli",
			CreatedAt: created,
			ExpiresAt: created.Add(24 * time.Hour),
		},
	}

	want := "id,code,target,owner_id,created_at,expires_at\n" +
		"1,000001,https://example.com,cli,2026-11-12T12:00:00Z,\n" +
		"2,000002,https://go.dev,cli,2026-11-12T12:00:00Z,2026-11-13T12:00:00Z\n"

	var buf bytes.Buffer

	if err := WriteCSV(&buf, links); err != nil {
		t.Fatalf("WriteCSV() error = %v", err)
	}

	if got := buf.String(); got != want {
		t.Errorf("WriteCSV() =\n%q\nwant:\n%q", got, want)
	}
}
