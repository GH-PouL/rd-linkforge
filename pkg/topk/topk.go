package topk

import (
	"container/heap"
	"errors"
	"slices"
	"strings"
)

// ErrInvalidN indicates that N must be positive.
var ErrInvalidN = errors.New("topk: n must be positive")

type Item struct {
	Key   string
	Count int64
}

// TopK keeps the N best items.
// It is not safe for concurrent use.
type TopK struct {
	n     int
	items itemHeap
}

func New(n int) (*TopK, error) {
	if n <= 0 {
		return nil, ErrInvalidN
	}

	t := &TopK{
		n:     n,
		items: make(itemHeap, 0, n),
	}

	heap.Init(&t.items)
	return t, nil
}

type itemHeap []Item

func (h itemHeap) Len() int {
	return len(h)
}

func (h itemHeap) Less(i, j int) bool {
	if h[i].Count == h[j].Count {
		return h[i].Key > h[j].Key
	}
	return h[i].Count < h[j].Count
}

func (h itemHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *itemHeap) Push(x any) {
	*h = append(*h, x.(Item))
}

func (h *itemHeap) Pop() any {
	old := *h
	n := len(old)

	item := old[n-1]
	*h = old[:n-1]

	return item
}

func (t *TopK) Add(key string, count int64) {
	item := Item{
		Key:   key,
		Count: count,
	}

	if t.items.Len() < t.n {
		heap.Push(&t.items, item)
		return
	}

	worst := t.items[0]

	if item.Count > worst.Count ||
		(item.Count == worst.Count && item.Key < worst.Key) {
		heap.Pop(&t.items)
		heap.Push(&t.items, item)
	}
}

func (t *TopK) Result() []Item {
	result := make([]Item, len(t.items))
	copy(result, t.items)

	slices.SortFunc(result, func(a, b Item) int {
		if a.Count > b.Count {
			return -1
		}
		if a.Count < b.Count {
			return 1
		}
		return strings.Compare(a.Key, b.Key)
	})

	return result
}
