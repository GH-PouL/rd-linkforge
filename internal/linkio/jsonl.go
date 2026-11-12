package linkio

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/skskuzan/rd-linkforge/internal/link"
)

// WriteLinks writes one JSON object per line.
func WriteLinks(w io.Writer, links []link.Link) error {
	encoder := json.NewEncoder(w)

	for i, l := range links {
		if err := encoder.Encode(l); err != nil {
			return fmt.Errorf("write record %d: %w", i+1, err)
		}
	}

	return nil
}

// ReadLinks reads JSON Lines, one link per line.
func ReadLinks(r io.Reader) ([]link.Link, error) {
	scanner := bufio.NewScanner(r)
	var links []link.Link
	record := 0

	for scanner.Scan() {
		record++

		var l link.Link

		if err := json.Unmarshal(scanner.Bytes(), &l); err != nil {
			return nil, fmt.Errorf("record %d: %w", record, err)
		}

		links = append(links, l)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read links: %w", err)
	}

	return links, nil
}
