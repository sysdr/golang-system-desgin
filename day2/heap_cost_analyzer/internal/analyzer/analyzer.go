package analyzer

import (
	"net/http"
	"time"
)

// RequestStats represents statistics for a single request.
type RequestStats struct {
	ID        uint64
	Timestamp int64
	Duration  time.Duration
	Status    int
}

// ProcessRequestPointer simulates processing a request and generating stats.
// It returns a *pointer* to RequestStats. This will likely escape to the heap.
func ProcessRequestPointer(requestID uint64) *RequestStats {
	stats := &RequestStats{ // &RequestStats literal escapes to heap
		ID:        requestID,
		Timestamp: time.Now().UnixNano(),
		Duration:  time.Millisecond * 10, // Simulate work
		Status:    http.StatusOK,
	}
	return stats
}

// ProcessRequestValue simulates processing a request and generating stats.
// It returns a *value* of RequestStats. This might stay on the stack.
func ProcessRequestValue(requestID uint64) RequestStats {
	stats := RequestStats{ // RequestStats struct does not escape (potentially)
		ID:        requestID,
		Timestamp: time.Now().UnixNano(),
		Duration:  time.Millisecond * 10, // Simulate work
		Status:    http.StatusOK,
	}
	return stats
}
