package shortid

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestCodeGenerator(t *testing.T) {
	random := bytes.NewReader([]byte{
		0, 0, 1,
		0, 0, 1,
		0, 0, 1,
	})

	g := NewCodeGenerator(random)

	want := []string{
		"000H32",
		"000H33",
		"000H34",
	}

	for i, expected := range want {
		got, err := g.Next()
		if err != nil {
			t.Fatalf("Next() call %d: %v", i+1, err)
		}

		if got != expected {
			t.Errorf("Next() = %q, want %q", got, expected)
		}
	}
}

func TestCodeGeneratorMaxRandom(t *testing.T) {
	g := NewCodeGenerator(bytes.NewReader([]byte{
		0xFF, 0xFF, 0xFF,
	}))

	got, err := g.Next()
	if err != nil {
		t.Fatal(err)
	}

	if got != "bVJgVU" {
		t.Errorf("Next() = %q, want bVJgVU", got)
	}
}

func TestCodeGeneratorReadError(t *testing.T) {
	g := NewCodeGenerator(bytes.NewReader([]byte{1, 2}))

	got, err := g.Next()

	if got != "" {
		t.Errorf("Next() = %q, want empty string", got)
	}

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("expected io.ErrUnexpectedEOF, got %v", err)
	}

	if g.counter != 0 {
		t.Errorf("counter = %d, want 0", g.counter)
	}
}

func TestCodeGeneratorCounterWrap(t *testing.T) {
	g := NewCodeGenerator(bytes.NewReader([]byte{
		0, 0, 0,
		0, 0, 0,
	}))

	g.counter = 65535

	first, err := g.Next()
	if err != nil {
		t.Fatal(err)
	}

	second, err := g.Next()
	if err != nil {
		t.Fatal(err)
	}

	if first != "000H31" {
		t.Errorf("first = %q, want 000H31", first)
	}

	if second != "000000" {
		t.Errorf("second = %q, want 000000", second)
	}

	if g.counter != 1 {
		t.Errorf("counter = %d, want 1", g.counter)
	}
}
