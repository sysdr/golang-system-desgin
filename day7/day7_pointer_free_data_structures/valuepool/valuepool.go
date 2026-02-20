package valuepool

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Item represents a pointer-free data structure.
type Item struct {
	ID    uint64
	Value [16]byte
}

func (i *Item) Reset() {
	i.ID = 0
	for j := range i.Value {
		i.Value[j] = 0
	}
}

type Pool struct {
	items    []Item
	freeList []int
	mu       sync.Mutex
	size     int
	inUse    atomic.Int64
}

func NewPool(capacity int) *Pool {
	if capacity <= 0 {
		capacity = 1
	}
	items := make([]Item, capacity)
	freeList := make([]int, capacity)
	for i := 0; i < capacity; i++ {
		items[i].ID = uint64(i)
		freeList[i] = i
	}
	return &Pool{items: items, freeList: freeList, size: capacity}
}

func (p *Pool) Acquire() (int, *Item, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.freeList) == 0 {
		return -1, nil, fmt.Errorf("value pool exhausted (capacity: %d, in use: %d)", p.size, p.inUse.Load())
	}
	idx := p.freeList[len(p.freeList)-1]
	p.freeList = p.freeList[:len(p.freeList)-1]
	p.inUse.Add(1)
	return idx, &p.items[idx], nil
}

func (p *Pool) Release(idx int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if idx < 0 || idx >= p.size {
		return fmt.Errorf("invalid item index %d for pool of size %d", idx, p.size)
	}
	p.freeList = append(p.freeList, idx)
	p.inUse.Add(-1)
	return nil
}

func (p *Pool) Stats() (capacity, available, inUse int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return int64(p.size), int64(len(p.freeList)), p.inUse.Load()
}

func (p *Pool) GetItem(idx int) *Item {
	if idx < 0 || idx >= p.size {
		return nil
	}
	return &p.items[idx]
}
