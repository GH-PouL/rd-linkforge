package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testEnv(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}

func TestRunBasicCommands(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"version", []string{"version"}, 0},
		{"no command", nil, 2},
		{"unknown command", []string{"unknown"}, 2},
		{"version extra argument", []string{"version", "extra"}, 2},
		{"import missing flags", []string{"import"}, 2},
		{"export missing flags", []string{"export"}, 2},
		{"import unknown flag", []string{"import", "-unknown"}, 2},
		{"export unknown flag", []string{"export", "-unknown"}, 2},
		{"config extra argument", []string{"config", "extra"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			got := run(
				context.Background(),
				tt.args,
				testEnv(nil),
				&stdout,
				&stderr,
			)

			if got != tt.want {
				t.Errorf("run() = %d, want %d; stderr: %s",
					got, tt.want, stderr.String())
			}
		})
	}
}

func TestRunImportExport(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "urls.txt")
	db := filepath.Join(dir, "links.jsonl")
	output := filepath.Join(dir, "links.csv")

	data := "https://example.com\ninvalid-url\nhttps://go.dev\n"
	if err := os.WriteFile(input, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	code := run(
		context.Background(),
		[]string{"import", "-in", input, "-db", db},
		testEnv(nil),
		&stdout,
		&stderr,
	)

	if code != 1 {
		t.Fatalf("import code = %d, want 1; stderr: %s",
			code, stderr.String())
	}

	if !strings.Contains(stdout.String(), "imported: 2") {
		t.Errorf("unexpected import output: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()

	code = run(
		context.Background(),
		[]string{"export", "-db", db, "-out", output},
		testEnv(nil),
		&stdout,
		&stderr,
	)

	if code != 0 {
		t.Fatalf("export code = %d; stderr: %s",
			code, stderr.String())
	}

	csvData, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(csvData), "https://example.com") ||
		!strings.Contains(string(csvData), "https://go.dev") {
		t.Errorf("unexpected CSV: %s", csvData)
	}

	stdout.Reset()
	stderr.Reset()

	code = run(
		context.Background(),
		[]string{"export", "-db", db},
		testEnv(nil),
		&stdout,
		&stderr,
	)

	if code != 0 {
		t.Fatalf("stdout export code = %d; stderr: %s",
			code, stderr.String())
	}

	if !strings.HasPrefix(stdout.String(), "id,code,target,") {
		t.Errorf("unexpected CSV header: %s", stdout.String())
	}
}

func TestRunImportErrors(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name string
		args []string
	}{
		{
			"missing input",
			[]string{
				"import",
				"-in", filepath.Join(dir, "missing.txt"),
				"-db", filepath.Join(dir, "db.jsonl"),
			},
		},
		{
			"missing database",
			[]string{
				"export",
				"-db", filepath.Join(dir, "missing.jsonl"),
			},
		},
		{
			"invalid output directory",
			[]string{
				"export",
				"-db", filepath.Join(dir, "db.jsonl"),
				"-out", filepath.Join(dir, "missing", "out.csv"),
			},
		},
	}

	// Prepare a valid empty database for the output-path test.
	if err := os.WriteFile(
		filepath.Join(dir, "db.jsonl"),
		nil,
		0600,
	); err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			code := run(
				context.Background(),
				tt.args,
				testEnv(nil),
				&stdout,
				&stderr,
			)

			if code != 1 {
				t.Errorf("run() = %d, want 1; stderr: %s",
					code, stderr.String())
			}
		})
	}
}

func TestRunCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var stdout, stderr bytes.Buffer

	code := run(
		ctx,
		[]string{"export", "-db", "anything.jsonl"},
		testEnv(nil),
		&stdout,
		&stderr,
	)

	if code != 1 {
		t.Errorf("run() = %d, want 1", code)
	}
}

func TestRunInvalidLogLevel(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run(
		context.Background(),
		[]string{"version"},
		testEnv(map[string]string{
			"LF_LOG_LEVEL": "invalid",
		}),
		&stdout,
		&stderr,
	)

	if code != 2 {
		t.Errorf("run() = %d, want 2", code)
	}
}
