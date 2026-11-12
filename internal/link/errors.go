package link

import "fmt"

// AliasError describes an error related to a specific alias.
type AliasError struct {
	Alias string
	Err   error
}

// Error returns an error message containing the alias.
func (e *AliasError) Error() string {
	return fmt.Sprintf("link: alias %q: %v", e.Alias, e.Err)
}

// Unwrap returns the underlying error.
func (e *AliasError) Unwrap() error {
	return e.Err
}
