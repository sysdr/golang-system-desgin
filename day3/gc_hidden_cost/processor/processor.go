package processor

import (
	"fmt"
	"sync"
)

// Processor defines the interface for our processing logic.
type Processor interface {
	Process(size int) (string, error)
}

// NaiveProcessor allocates a new byte slice for each request.
type NaiveProcessor struct{}

func NewNaiveProcessor() *NaiveProcessor {
	return &NaiveProcessor{}
}

func (p *NaiveProcessor) Process(size int) (string, error) {
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = byte(i % 256) // Dummy write
	}
	return fmt.Sprintf("Naive: Processed %d bytes", size), nil
}

// PooledProcessor uses sync.Pool to reuse byte slices.
type PooledProcessor struct {
	pool sync.Pool
}

func NewPooledProcessor() *PooledProcessor {
	return &PooledProcessor{
		pool: sync.Pool{
			New: func() interface{} {
				// Initial capacity for new buffers. Will be adjusted if needed.
				// We return a slice with len 0 but capacity 4KB as a starting point.
				return make([]byte, 0, 4096)
			},
		},
	}
}

func (p *PooledProcessor) Process(size int) (string, error) {
	// Get a buffer from the pool.
	// It's returned as an interface{}, so we need a type assertion.
	// If the pool is empty, p.pool.New() will be called.
	bufIface := p.pool.Get()
	buf, ok := bufIface.([]byte)
	if !ok {
		return "", fmt.Errorf("failed to assert type from sync.Pool")
	}

	// Ensure the buffer has enough capacity. If not, create a new one.
	if cap(buf) < size {
		// The retrieved buffer is too small, discard it and create a new one.
		// Note: The old buffer will be GC'd if no other references exist.
		buf = make([]byte, size)
	} else {
		// The buffer has sufficient capacity, re-slice it to the desired length.
		buf = buf[:size]
	}

	for i := 0; i < size; i++ {
		buf[i] = byte(i % 256) // Dummy write
	}

	// Return the buffer to the pool.
	// IMPORTANT: Reset the length to 0 to prevent retaining stale data and
	// to allow subsequent Get() calls to correctly use len(0) if needed.
	p.pool.Put(buf[:0])

	return fmt.Sprintf("Pooled: Processed %d bytes", size), nil
}
