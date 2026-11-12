package linkio

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Target represents one URL from an import file.
type Target struct {
	Line int
	URL  string
}

// ReadTargets reads non-empty, non-comment lines from a reader.
func ReadTargets(r io.Reader) ([]Target, error) {
	scanner := bufio.NewScanner(r)
	var targets []Target
	lineNum := 0

	for scanner.Scan() {
		lineNum++

		url := strings.TrimSpace(scanner.Text())

		if url == "" || strings.HasPrefix(url, "#") {
			continue
		}

		targets = append(targets, Target{
			Line: lineNum,
			URL:  url,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read targets: %w", err)
	}

	return targets, nil
}
