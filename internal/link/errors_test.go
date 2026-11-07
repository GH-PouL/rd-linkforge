package link

import (
	"errors"
	"strings"
	"testing"
)

func TestAliasErrorCustom(t *testing.T) {
	err := &AliasError{
		Alias: "google",
		Err:   ErrAliasTaken,
	}

	// Перевіряємо, що повідомлення містить alias.
	if !strings.Contains(err.Error(), "google") {
		t.Errorf("Error() = %q, want alias google", err.Error())
	}

	// Перевіряємо доступ до початкової помилки.
	if !errors.Is(err, ErrAliasTaken) {
		t.Error("errors.Is() should match ErrAliasTaken")
	}

	// Перевіряємо отримання конкретного типу помилки.
	var aliasErr *AliasError

	if !errors.As(err, &aliasErr) {
		t.Fatal("errors.As() should find AliasError")
	}

	if aliasErr.Alias != "google" {
		t.Errorf("Alias = %q, want google", aliasErr.Alias)
	}
}
