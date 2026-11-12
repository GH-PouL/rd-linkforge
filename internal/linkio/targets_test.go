package linkio

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestReadTargets(t *testing.T) {
	input := `# Мої посилання
https://example.com

  https://go.dev
  # Ще один коментар
https://github.com
`

	want := []Target{
		{Line: 2, URL: "https://example.com"},
		{Line: 4, URL: "https://go.dev"},
		{Line: 6, URL: "https://github.com"},
	}

	got, err := ReadTargets(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadTargets() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadTargets() = %+v, want %+v", got, want)
	}
}

type failingReader struct {
	err error
}

func (r failingReader) Read([]byte) (int, error) {
	return 0, r.err
}

func TestReadTargetsReaderError(t *testing.T) {
	wantErr := errors.New("read failed")

	reader := failingReader{err: wantErr}

	_, err := ReadTargets(reader)

	if !errors.Is(err, wantErr) {
		t.Errorf("ReadTargets() error = %v, want %v", err, wantErr)
	}
}
