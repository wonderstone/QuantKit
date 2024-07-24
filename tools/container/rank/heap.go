package rank

import (
	"container/heap"
)

type Heap[Item any] struct {
	data []Item
	less func(i, j Item) bool
}

func (h Heap[Item]) Len() int           { return len(h.data) }
func (h Heap[Item]) Less(i, j int) bool { return h.less(h.data[i], h.data[j]) }
func (h Heap[Item]) Swap(i, j int)      { h.data[i], h.data[j] = h.data[j], h.data[i] }

func (h *Heap[Item]) Push(x any) {
	h.data = append(h.data, x.(Item))
}

func (h *Heap[Item]) Pop() any {
	old := h.data
	n := len(old)
	x := old[n-1]
	h.data = old[0 : n-1]
	return x
}

type SortedHeap[Item any] struct {
	h *Heap[Item]
	n int
}

func NewSortedHeap[Item any](n int, less func(i, j Item) bool) *SortedHeap[Item] {
	return &SortedHeap[Item]{h: &Heap[Item]{less: less}, n: n}
}

func (c *SortedHeap[Item]) Insert(i Item) {
	heap.Push(c.h, i)
	if c.h.Len() > c.n {
		heap.Pop(c.h)
	}
}

func (c *SortedHeap[Item]) Values() []Item {
	return c.h.data
}
