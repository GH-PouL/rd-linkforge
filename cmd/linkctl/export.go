package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/skskuzan/rd-linkforge/internal/linkio"
)

func runExport(
	ctx context.Context,
	args []string,
	stdout, stderr io.Writer,
) int {
	flags := flag.NewFlagSet("export", flag.ContinueOnError)
	flags.SetOutput(stderr)

	db := flags.String("db", "", "database file")
	out := flags.String("out", "", "output CSV file")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	if *db == "" || flags.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: linkctl export -db FILE [-out FILE]")
		return 2
	}

	if err := ctx.Err(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	links, err := linkio.LoadFile(*db)
	if err != nil {
		fmt.Fprintf(stderr, "load database: %v\n", err)
		return 1
	}

	if *out == "" {
		if err := linkio.WriteCSV(stdout, links); err != nil {
			fmt.Fprintf(stderr, "export CSV: %v\n", err)
			return 1
		}
		return 0
	}

	file, err := os.Create(*out)
	if err != nil {
		fmt.Fprintf(stderr, "create output: %v\n", err)
		return 1
	}

	writeErr := linkio.WriteCSV(file, links)
	closeErr := file.Close()

	if err := errors.Join(writeErr, closeErr); err != nil {
		fmt.Fprintf(stderr, "export CSV: %v\n", err)
		return 1
	}

	return 0
}
