package trie

import (
	"fmt"
	"testing"
)

func BenchmarkTrieWithPrefix(b *testing.B) {
	tr := New()

	// Build a trie with 10,000 words.
	for i := 0; i < 10000; i++ {
		tr.Insert(fmt.Sprintf("link%05d", i))
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := tr.WithPrefix("link000", 10)
		if len(result) != 10 {
			b.Fatalf("got %d words, want 10", len(result))
		}
	}
}
