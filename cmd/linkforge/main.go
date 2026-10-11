package main

import (
	"fmt"
	"os"

	"github.com/skskuzan/rd-linkforge/internal/link"
	"github.com/skskuzan/rd-linkforge/internal/store"
)

func main() {
	s := store.New()
	var nextID uint64 = 1

	for _, target := range os.Args[1:] {
		l, err := link.New(nextID, target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid URL %q: %v\n", target, err)
			continue
		}

		if err := store.Add(s, l); err != nil {
			fmt.Fprintf(os.Stderr, "cannot store URL %q: %v\n", target, err)
			continue
		}

		fmt.Printf("%s  %s\n", l.Code, l.Target)
		nextID++
	}

	fmt.Printf("\nTotal: %d\n", store.Count(s))
}
