package linkio

import (
	"fmt"
	"os"

	"github.com/skskuzan/rd-linkforge/internal/link"
)

// LoadFile reads links from a JSON Lines file.
func LoadFile(path string) ([]link.Link, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	links, err := ReadLinks(file)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	return links, nil
}

// SaveFile writes links to a JSON Lines file.
func SaveFile(path string, links []link.Link) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}

	writeErr := WriteLinks(file, links)
	closeErr := file.Close()

	if writeErr != nil {
		return fmt.Errorf("write %s: %w", path, writeErr)
	}

	if closeErr != nil {
		return fmt.Errorf("close %s: %w", path, closeErr)
	}

	return nil
}
