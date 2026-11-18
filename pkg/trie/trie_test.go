package trie

import (
	"slices"
	"testing"
)

func TestTrieInsertContains(t *testing.T) {
	tr := New()

	if tr.Insert("") {
		t.Error("Insert(empty) must return false")
	}

	if !tr.Insert("doc") {
		t.Error("Insert(doc) must return true")
	}

	if tr.Insert("doc") {
		t.Error("duplicate Insert(doc) must return false")
	}

	if !tr.Insert("docs") {
		t.Error("Insert(docs) must return true")
	}

	tests := []struct {
		word string
		want bool
	}{
		{"doc", true},
		{"docs", true},
		{"do", false},
		{"dog", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			if got := tr.Contains(tt.word); got != tt.want {
				t.Errorf("Contains(%q) = %v, want %v",
					tt.word, got, tt.want)
			}
		})
	}
}

func TestTrieWithPrefix(t *testing.T) {
	tr := New()

	for _, word := range []string{
		"docs", "doc", "dog", "do-it", "day", "api", "dock",
	} {
		tr.Insert(word)
	}

	tests := []struct {
		name   string
		prefix string
		limit  int
		want   []string
	}{
		{
			"all do", "do", 0,
			[]string{"do-it", "doc", "dock", "docs", "dog"},
		},
		{
			"limited", "do", 2,
			[]string{"do-it", "doc"},
		},
		{
			"unknown", "xyz", 0, nil,
		},
		{
			"exact word", "doc", 0,
			[]string{"doc", "dock", "docs"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tr.WithPrefix(tt.prefix, tt.limit)

			if !slices.Equal(got, tt.want) {
				t.Errorf("WithPrefix(%q, %d) = %v, want %v",
					tt.prefix, tt.limit, got, tt.want)
			}
		})
	}
}

func TestTrieDelete(t *testing.T) {
	tr := New()

	for _, word := range []string{
		"doc", "docs", "dog", "cat",
	} {
		tr.Insert(word)
	}

	// Delete a word that is a prefix of another word.
	if !tr.Delete("doc") {
		t.Error("Delete(doc) must return true")
	}

	if tr.Contains("doc") {
		t.Error("doc must be deleted")
	}

	if !tr.Contains("docs") {
		t.Error("docs must remain")
	}

	// Delete a leaf word.
	if !tr.Delete("docs") {
		t.Error("Delete(docs) must return true")
	}

	if tr.Contains("docs") {
		t.Error("docs must be deleted")
	}

	if !tr.Contains("dog") {
		t.Error("dog must remain")
	}

	// Delete a word with no shared prefix.
	if !tr.Delete("cat") {
		t.Error("Delete(cat) must return true")
	}

	if tr.Contains("cat") {
		t.Error("cat must be deleted")
	}

	// Deleting a missing word does nothing.
	if tr.Delete("missing") {
		t.Error("Delete(missing) must return false")
	}

	if tr.Delete("") {
		t.Error("Delete(empty) must return false")
	}
}
