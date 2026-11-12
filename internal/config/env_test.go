package config

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestEnvLoaderDefaults(t *testing.T) {
	env := map[string]string{
		"LF_PG_DSN":     "postgres",
		"LF_MONGO_URI":  "mongodb",
		"LF_JWT_SECRET": "12345678901234567890123456789012",
	}

	lookup := func(key string) (string, bool) {
		value, ok := env[key]
		return value, ok
	}

	loader := NewEnvLoader(lookup)

	cfg, err := loader.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}

	if cfg.AdminAddr != "127.0.0.1:6060" {
		t.Errorf("AdminAddr = %q, want 127.0.0.1:6060", cfg.AdminAddr)
	}

	if cfg.TCPAddr != ":7070" {
		t.Errorf("TCPAddr = %q, want :7070", cfg.TCPAddr)
	}

	if cfg.CacheSize != 100000 {
		t.Errorf("CacheSize = %d, want 100000", cfg.CacheSize)
	}

	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want INFO", cfg.LogLevel)
	}
}

func TestEnvLoaderInvalid(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr error
	}{
		{"missing PG DSN", "LF_PG_DSN", "", ErrMissing},
		{"missing Mongo URI", "LF_MONGO_URI", " ", ErrMissing},
		{"missing JWT secret", "LF_JWT_SECRET", "", ErrMissing},
		{"short JWT secret", "LF_JWT_SECRET", "short", ErrInvalid},
		{"invalid HTTP port", "LF_HTTP_ADDR", "localhost:abc", ErrInvalid},
		{"zero HTTP port", "LF_HTTP_ADDR", "localhost:0", ErrInvalid},
		{"large TCP port", "LF_TCP_ADDR", ":70000", ErrInvalid},
		{"invalid admin address", "LF_ADMIN_ADDR", "localhost", ErrInvalid},
		{"invalid cache size", "LF_CACHE_SIZE", "abc", ErrInvalid},
		{"zero cache size", "LF_CACHE_SIZE", "0", ErrInvalid},
		{"negative cache size", "LF_CACHE_SIZE", "-10", ErrInvalid},
		{"invalid log level", "LF_LOG_LEVEL", "warning", ErrInvalid},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := map[string]string{
				"LF_PG_DSN":     "postgres",
				"LF_MONGO_URI":  "mongodb",
				"LF_JWT_SECRET": "12345678901234567890123456789012",
			}

			env[tc.key] = tc.value

			lookup := func(key string) (string, bool) {
				value, ok := env[key]
				return value, ok
			}

			cfg, err := NewEnvLoader(lookup).Load(context.Background())

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Load() error = %v, want %v", err, tc.wantErr)
			}

			if cfg != (Config{}) {
				t.Errorf("Load() returned non-zero Config: %+v", cfg)
			}
		})
	}
}

func TestEnvLoaderOverrides(t *testing.T) {
	env := map[string]string{
		"LF_HTTP_ADDR":  "  localhost:9090  ",
		"LF_ADMIN_ADDR": " 127.0.0.1:6061 ",
		"LF_TCP_ADDR":   " :7071 ",
		"LF_PG_DSN":     " postgres ",
		"LF_MONGO_URI":  " mongodb ",
		"LF_JWT_SECRET": " 12345678901234567890123456789012 ",
		"LF_CACHE_SIZE": " 2500 ",
		"LF_LOG_LEVEL":  " DeBuG ",
	}

	lookup := func(key string) (string, bool) {
		value, ok := env[key]
		return value, ok
	}

	cfg, err := NewEnvLoader(lookup).Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if cfg.HTTPAddr != "localhost:9090" {
		t.Errorf("HTTPAddr = %q", cfg.HTTPAddr)
	}

	if cfg.AdminAddr != "127.0.0.1:6061" {
		t.Errorf("AdminAddr = %q", cfg.AdminAddr)
	}

	if cfg.TCPAddr != ":7071" {
		t.Errorf("TCPAddr = %q", cfg.TCPAddr)
	}

	if cfg.PGDSN != "postgres" {
		t.Errorf("PGDSN was not trimmed")
	}

	if cfg.MongoURI != "mongodb" {
		t.Errorf("MongoURI was not trimmed")
	}

	if cfg.JWTSecret != "12345678901234567890123456789012" {
		t.Errorf("JWTSecret was not trimmed")
	}

	if cfg.CacheSize != 2500 {
		t.Errorf("CacheSize = %d, want 2500", cfg.CacheSize)
	}

	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel = %v, want DEBUG", cfg.LogLevel)
	}
}

func TestEnvLoaderMultipleErrors(t *testing.T) {
	lookup := func(key string) (string, bool) {
		return "", false
	}

	cfg, err := NewEnvLoader(lookup).Load(context.Background())

	if err == nil {
		t.Fatal("expected configuration errors")
	}

	for _, key := range []string{
		"LF_PG_DSN",
		"LF_MONGO_URI",
		"LF_JWT_SECRET",
	} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("error does not mention %s", key)
		}
	}

	if !errors.Is(err, ErrMissing) {
		t.Errorf("expected ErrMissing, got %v", err)
	}

	if cfg != (Config{}) {
		t.Errorf("expected zero Config, got %+v", cfg)
	}
}
func TestEnvLoaderCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0

	lookup := func(string) (string, bool) {
		calls++
		return "", false
	}

	loader := NewEnvLoader(lookup)

	cfg, err := loader.Load(ctx)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Load() error = %v, want context.Canceled", err)
	}

	if calls != 0 {
		t.Errorf("lookup called %d times, want 0", calls)
	}

	if cfg != (Config{}) {
		t.Errorf("Load() returned non-zero Config: %+v", cfg)
	}
}
func TestConfigLogValue(t *testing.T) {
	cfg := Config{
		HTTPAddr:  ":8080",
		AdminAddr: "127.0.0.1:6060",
		TCPAddr:   ":7070",
		PGDSN:     "postgres-secret",
		MongoURI:  "mongodb-secret",
		JWTSecret: "jwt-secret-value",
		CacheSize: 100000,
		LogLevel:  slog.LevelInfo,
	}

	var buf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&buf, nil))
	logger.Info("configuration loaded", "config", cfg)

	output := buf.String()

	for _, secret := range []string{
		cfg.PGDSN,
		cfg.MongoURI,
		cfg.JWTSecret,
	} {
		if strings.Contains(output, secret) {
			t.Errorf("secret leaked into logs: %q", secret)
		}
	}

	for _, key := range []string{
		"config.http_addr",
		"config.admin_addr",
		"config.tcp_addr",
		"config.pg_dsn",
		"config.mongo_uri",
		"config.jwt_secret",
		"config.cache_size",
		"config.log_level",
	} {
		if !strings.Contains(output, key) {
			t.Errorf("missing log attribute: %s", key)
		}
	}

	if count := strings.Count(output, "[redacted]"); count != 3 {
		t.Errorf("redacted fields = %d, want 3", count)
	}
}
