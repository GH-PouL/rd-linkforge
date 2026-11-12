package config

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	ErrMissing = errors.New("config: required variable is not set")
	ErrInvalid = errors.New("config: invalid value")
)

type EnvLoader struct {
	lookup func(string) (string, bool)
}

// NewEnvLoader creates a loader that reads environment variables.
// A nil lookup uses os.LookupEnv.
func NewEnvLoader(lookup func(string) (string, bool)) *EnvLoader {
	if lookup == nil {
		lookup = os.LookupEnv
	}

	return &EnvLoader{
		lookup: lookup,
	}
}

func validateAddr(addr string) error {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	n, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	if n < 1 || n > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	return nil
}

func (l *EnvLoader) Load(ctx context.Context) (Config, error) {
	if err := ctx.Err(); err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTPAddr:  ":8080",
		AdminAddr: "127.0.0.1:6060",
		TCPAddr:   ":7070",
		CacheSize: 100000,
		LogLevel:  slog.LevelInfo,
	}

	var errs []error

	if value, ok := l.lookup("LF_HTTP_ADDR"); ok {
		if value = strings.TrimSpace(value); value != "" {
			cfg.HTTPAddr = value
		}
	}

	if value, ok := l.lookup("LF_ADMIN_ADDR"); ok {
		if value = strings.TrimSpace(value); value != "" {
			cfg.AdminAddr = value
		}
	}

	if value, ok := l.lookup("LF_TCP_ADDR"); ok {
		if value = strings.TrimSpace(value); value != "" {
			cfg.TCPAddr = value
		}
	}

	if value, ok := l.lookup("LF_PG_DSN"); ok {
		cfg.PGDSN = strings.TrimSpace(value)
	}

	if value, ok := l.lookup("LF_MONGO_URI"); ok {
		cfg.MongoURI = strings.TrimSpace(value)
	}

	if value, ok := l.lookup("LF_JWT_SECRET"); ok {
		cfg.JWTSecret = strings.TrimSpace(value)
	}

	for _, item := range []struct {
		name string
		addr string
	}{
		{"LF_HTTP_ADDR", cfg.HTTPAddr},
		{"LF_ADMIN_ADDR", cfg.AdminAddr},
		{"LF_TCP_ADDR", cfg.TCPAddr},
	} {
		if err := validateAddr(item.addr); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", item.name, ErrInvalid))
		}
	}

	if cfg.PGDSN == "" {
		errs = append(errs,
			fmt.Errorf("LF_PG_DSN: %w", ErrMissing))
	}

	if cfg.MongoURI == "" {
		errs = append(errs,
			fmt.Errorf("LF_MONGO_URI: %w", ErrMissing))
	}

	if cfg.JWTSecret == "" {
		errs = append(errs,
			fmt.Errorf("LF_JWT_SECRET: %w", ErrMissing))
	} else if len(cfg.JWTSecret) < 32 {
		errs = append(errs,
			fmt.Errorf("LF_JWT_SECRET: %w", ErrInvalid))
	}

	if value, ok := l.lookup("LF_CACHE_SIZE"); ok {
		value = strings.TrimSpace(value)

		if value != "" {
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				errs = append(errs,
					fmt.Errorf("LF_CACHE_SIZE: %w", ErrInvalid))
			} else {
				cfg.CacheSize = size
			}
		}
	}

	if value, ok := l.lookup("LF_LOG_LEVEL"); ok {
		value = strings.TrimSpace(value)

		if value != "" {
			level, err := ParseLevel(value)
			if err != nil {
				errs = append(errs, err)
			} else {
				cfg.LogLevel = level
			}
		}
	}

	if len(errs) > 0 {
		return Config{}, errors.Join(errs...)
	}

	return cfg, nil
}

func ParseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("LF_LOG_LEVEL: %w", ErrInvalid)
	}
}

var _ Loader = (*EnvLoader)(nil)
