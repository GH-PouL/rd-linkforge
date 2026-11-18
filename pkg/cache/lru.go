package cache

import (
	"container/list"
	"errors"
)

type entry[K comparable, V any] struct {
	key   K
	value V
}

// LRU is a least-recently-used cache.
// It is not safe for concurrent use.
type LRU[K comparable, V any] struct {
	capacity int
	items    map[K]*list.Element
	order    *list.List
}

var ErrInvalidCapacity = errors.New("cache: capacity must be positive")

func NewLRU[K comparable, V any](capacity int) (*LRU[K, V], error) {
	if capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	return &LRU[K, V]{
		capacity: capacity,
		items:    make(map[K]*list.Element),
		order:    list.New(),
	}, nil
}

func (c *LRU[K, V]) Len() int {
	return len(c.items)
}

var _ Cache[string, int] = (*LRU[string, int])(nil)

func (c *LRU[K, V]) Get(key K) (V, bool) {
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		return elem.Value.(entry[K, V]).value, true
	}

	var zero V
	return zero, false
}

func (c *LRU[K, V]) Put(key K, value V) {
	if elem, ok := c.items[key]; ok {
		elem.Value = entry[K, V]{key: key, value: value}
		c.order.MoveToFront(elem)
		return
	}

	if len(c.items) >= c.capacity {
		oldest := c.order.Back()
		oldEntry := oldest.Value.(entry[K, V])

		delete(c.items, oldEntry.key)
		c.order.Remove(oldest)
	}

	elem := c.order.PushFront(entry[K, V]{
		key:   key,
		value: value,
	})
	c.items[key] = elem
}

func (c *LRU[K, V]) Delete(key K) {
	if elem, ok := c.items[key]; ok {
		delete(c.items, key)
		c.order.Remove(elem)
	}
}

func (c *LRU[K, V]) Keys() []K {
	keys := make([]K, 0, len(c.items))

	for elem := c.order.Front(); elem != nil; elem = elem.Next() {
		item := elem.Value.(entry[K, V])
		keys = append(keys, item.key)
	}

	return keys
}
