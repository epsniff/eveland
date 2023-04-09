package dbmarketorders

import (
	"container/heap"

	"github.com/epsniff/eveland/src/evesdk"
)

type MinHeap []*evesdk.MarketOrder

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].Price < h[j].Price }
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(*evesdk.MarketOrder))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func NewMinHeap() *MinHeap {
	h := &MinHeap{}
	heap.Init(h)
	return h
}

type MaxHeap []*evesdk.MarketOrder

func (h MaxHeap) Len() int           { return len(h) }
func (h MaxHeap) Less(i, j int) bool { return h[i].Price > h[j].Price }
func (h MaxHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MaxHeap) Push(x interface{}) {
	*h = append(*h, x.(*evesdk.MarketOrder))
}

func (h *MaxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func NewMaxHeap() *MaxHeap {
	h := &MaxHeap{}
	heap.Init(h)
	return h
}
