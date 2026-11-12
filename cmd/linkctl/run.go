package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/skskuzan/rd-linkforge/internal/config"
)

func run(
	ctx context.Context,
	args []string,
	env func(string) (string, bool),
	stdout, stderr io.Writer,
) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: linkctl <import|export|config|version>")
		return 2
	}

	level := slog.LevelInfo

	if env == nil {
		env = func(string) (string, bool) {
			return "", false
		}
	}

	if value, ok := env("LF_LOG_LEVEL"); ok && strings.TrimSpace(value) != "" {
		parsed, err := config.ParseLevel(value)
		if err != nil {
			fmt.Fprintln(stderr, "invalid LF_LOG_LEVEL")
			return 2
		}
		level = parsed
	}

	logger := slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{
		Level: level,
	}))

	switch args[0] {
	case "version":
		if len(args) != 1 {
			fmt.Fprintln(stderr, "usage: linkctl version")
			return 2
		}
		fmt.Fprintf(stdout, "linkctl %s\n", version)
		return 0

	case "config":
		if len(args) != 1 {
			fmt.Fprintln(stderr, "usage: linkctl config")
			return 2
		}

		cfg, err := config.NewEnvLoader(env).Load(ctx)
		if err != nil {
			fmt.Fprintf(stderr, "invalid configuration: %v\n", err)
			return 1
		}

		logger.Info("configuration loaded", "config", cfg)
		fmt.Fprintln(stdout, "config ok")
		return 0

	case "import":
		return runImport(ctx, args[1:], stdout, stderr, logger)

	case "export":
		return runExport(ctx, args[1:], stdout, stderr)

	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", args[0])
		return 2
	}
}
