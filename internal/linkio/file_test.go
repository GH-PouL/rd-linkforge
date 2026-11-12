package linkio

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/skskuzan/rd-linkforge/internal/link"
)

func TestSaveLoadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "links.jsonl")

	want := []link.Link{
		{
			ID:        1,
			Code:      "000001",
			Target:    "https://example.com",
			OwnerID:   "cli",
			CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	if err := SaveFile(path, want); err != nil {
		t.Fatalf("SaveFile: %v", err)
	}

	got, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLoadFileNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.jsonl")

	_, err := LoadFile(path)

	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("got %v, want fs.ErrNotExist", err)
	}
}

func TestLoadFileInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.jsonl")

	if err := os.WriteFile(path, []byte("{invalid json}\n"), 0600); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadFile(path); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestSaveFileInvalidPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "links.jsonl")

	if err := SaveFile(path, nil); err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestSaveFileOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "links.jsonl")

	if err := SaveFile(path, []link.Link{{ID: 1}}); err != nil {
		t.Fatal(err)
	}

	if err := SaveFile(path, []link.Link{{ID: 2}}); err != nil {
		t.Fatal(err)
	}

	got, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 || got[0].ID != 2 {
		t.Errorf("overwrite failed: %+v", got)
	}
}
