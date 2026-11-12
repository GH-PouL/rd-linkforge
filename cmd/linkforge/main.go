package main

import (
	"context"
	"fmt"
	"os"

	"github.com/skskuzan/rd-linkforge/internal/link"
	"github.com/skskuzan/rd-linkforge/internal/store"
)

func main() {
	s := store.New()
	shortener := link.NewService(s)
	ctx := context.Background()

	for _, target := range os.Args[1:] {
		_, err := shortener.Shorten(ctx, link.ShortenRequest{
			Target:  target,
			OwnerID: "demo",
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid URL %q: %v\n", target, err)
			continue
		}
	}

	for _, l := range s.All() {
		fmt.Printf("%s  %s\n", l.Code, l.Target)
	}

	fmt.Printf("\nTotal: %d\n", s.Count())
}
