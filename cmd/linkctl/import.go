package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"sort"

	"github.com/skskuzan/rd-linkforge/internal/link"
	"github.com/skskuzan/rd-linkforge/internal/linkio"
	"github.com/skskuzan/rd-linkforge/internal/store"
)

func runImport(
	ctx context.Context,
	args []string,
	stdout, stderr io.Writer,
	logger *slog.Logger,
) int {
	flags := flag.NewFlagSet("import", flag.ContinueOnError)
	flags.SetOutput(stderr)

	input := flags.String("in", "", "input file")
	db := flags.String("db", "", "database file")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	if *input == "" || *db == "" || flags.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: linkctl import -in FILE -db FILE")
		return 2
	}

	if err := ctx.Err(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	logger.Info("import started")

	file, err := os.Open(*input)
	if err != nil {
		fmt.Fprintf(stderr, "open input: %v\n", err)
		return 1
	}

	targets, readErr := linkio.ReadTargets(file)
	closeErr := file.Close()

	if err := errors.Join(readErr, closeErr); err != nil {
		fmt.Fprintf(stderr, "read input: %v\n", err)
		return 1
	}

	existing, err := linkio.LoadFile(*db)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintf(stderr, "load database: %v\n", err)
		return 1
	}

	st := store.New()
	var maxID uint64

	for _, l := range existing {
		if err := ctx.Err(); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}

		if _, err := st.Create(ctx, l); err != nil {
			fmt.Fprintf(stderr, "restore link: %v\n", err)
			return 1
		}

		if l.ID > maxID {
			maxID = l.ID
		}
	}

	if maxID == ^uint64(0) {
		fmt.Fprintln(stderr, "maximum link ID reached")
		return 1
	}

	svc := link.NewService(st, link.WithFirstID(maxID+1))

	imported, skipped := 0, 0

	for _, target := range targets {
		if err := ctx.Err(); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}

		l, err := svc.Shorten(ctx, link.ShortenRequest{
			Target:  target.URL,
			OwnerID: "cli",
		})

		if err != nil {
			if ctx.Err() != nil {
				fmt.Fprintln(stderr, ctx.Err())
				return 1
			}

			skipped++
			logger.Warn("import skipped",
				"line", target.Line,
				"error", err,
			)
			continue
		}

		imported++
		logger.Debug("link imported", "code", l.Code)
	}

	if err := ctx.Err(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	all := st.All()
	sort.Slice(all, func(i, j int) bool {
		return all[i].ID < all[j].ID
	})

	if err := linkio.SaveFile(*db, all); err != nil {
		fmt.Fprintf(stderr, "save database: %v\n", err)
		return 1
	}

	logger.Info("import finished",
		"imported", imported,
		"skipped", skipped,
	)

	fmt.Fprintf(stdout, "imported: %d\nskipped: %d\n", imported, skipped)

	if skipped > 0 {
		return 1
	}

	return 0
}
