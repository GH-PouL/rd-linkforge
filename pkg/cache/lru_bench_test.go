package cache

import "testing"

func BenchmarkLRUGet(b *testing.B) {
	c, err := NewLRU[string, int](1000)
	if err != nil {
		b.Fatal(err)
	}

	c.Put("test-key", 123)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		value, ok := c.Get("test-key")
		if !ok || value != 123 {
			b.Fatal("unexpected cache miss")
		}
	}
}
