package link

import (
	"context"
	"errors"
	"testing"
	"time"
)

type testCache struct {
	items   map[string]Link
	gets    int
	puts    int
	deletes int
}

func newTestCache() *testCache {
	return &testCache{
		items: make(map[string]Link),
	}
}

func (c *testCache) Get(code string) (Link, bool) {
	c.gets++
	l, ok := c.items[code]
	return l, ok
}

func (c *testCache) Put(code string, l Link) {
	c.puts++
	c.items[code] = l
}

func (c *testCache) Delete(code string) {
	c.deletes++
	delete(c.items, code)
}

func TestServiceResolveWithCache(t *testing.T) {
	ctx := context.Background()

	st := newHW2Store()
	cache := newTestCache()
	svc := NewService(st, WithCache(cache))

	l, err := svc.Shorten(ctx, ShortenRequest{
		Target:  "https://example.com",
		OwnerID: "user1",
	})
	if err != nil {
		t.Fatal(err)
	}

	// First Resolve: cache miss, store hit.
	_, err = svc.Resolve(ctx, l.Code)
	if err != nil {
		t.Fatal(err)
	}

	if cache.gets != 1 || cache.puts != 1 {
		t.Errorf("first Resolve: gets=%d, puts=%d; want 1, 1",
			cache.gets, cache.puts)
	}

	// Remove the link directly from the store.
	// The cached copy must still be available.
	delete(st.byID, l.ID)

	// Second Resolve: cache hit.
	_, err = svc.Resolve(ctx, l.Code)
	if err != nil {
		t.Fatal(err)
	}

	if cache.gets != 2 || cache.puts != 1 {
		t.Errorf("second Resolve: gets=%d, puts=%d; want 2, 1",
			cache.gets, cache.puts)
	}
}

func TestServiceDeleteWithCache(t *testing.T) {
	ctx := context.Background()

	st := newHW2Store()
	cache := newTestCache()
	svc := NewService(st, WithCache(cache))

	l, err := svc.Shorten(ctx, ShortenRequest{
		Target:  "https://example.com",
		OwnerID: "user1",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Fill the cache.
	if _, err := svc.Resolve(ctx, l.Code); err != nil {
		t.Fatal(err)
	}

	// Another owner cannot delete the link.
	if err := svc.Delete(ctx, "user2", l.Code); !errors.Is(err, ErrForbidden) {
		t.Errorf("Delete by another owner: got %v, want ErrForbidden", err)
	}

	if cache.deletes != 0 {
		t.Errorf("cache deletes = %d, want 0", cache.deletes)
	}

	// The actual owner deletes the link.
	if err := svc.Delete(ctx, "user1", l.Code); err != nil {
		t.Fatal(err)
	}

	if cache.deletes != 1 {
		t.Errorf("cache deletes = %d, want 1", cache.deletes)
	}

	if _, ok := cache.Get(l.Code); ok {
		t.Error("deleted link must not remain in cache")
	}

	if _, err := svc.Resolve(ctx, l.Code); !errors.Is(err, ErrNotFound) {
		t.Errorf("Resolve after Delete: got %v, want ErrNotFound", err)
	}
}

func TestServiceCachedLinkExpiry(t *testing.T) {
	ctx := context.Background()

	now := time.Date(2030, 1, 1, 12, 0, 0, 0, time.UTC)
	expires := now.Add(time.Hour)

	st := newHW2Store()
	cache := newTestCache()

	svc := NewService(
		st,
		WithCache(cache),
		WithClock(func() time.Time { return now }),
	)

	l, err := svc.Shorten(ctx, ShortenRequest{
		Target:    "https://example.com",
		OwnerID:   "user1",
		ExpiresAt: expires,
	})
	if err != nil {
		t.Fatal(err)
	}

	// First Resolve fills the cache.
	if _, err := svc.Resolve(ctx, l.Code); err != nil {
		t.Fatal(err)
	}

	// Move the clock to the expiry time.
	now = expires

	// The cached link must now be expired.
	if _, err := svc.Resolve(ctx, l.Code); !errors.Is(err, ErrExpired) {
		t.Errorf("Resolve at expiry: got %v, want ErrExpired", err)
	}

	if cache.puts != 1 {
		t.Errorf("cache puts = %d, want 1", cache.puts)
	}
}
