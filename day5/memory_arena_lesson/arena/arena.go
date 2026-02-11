package arena

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Arena represents a simple memory arena for a specific type T.
// It allocates a large chunk of memory and doles out pointers to it.
// When reset, all "allocated" memory is effectively freed without GC.
type Arena[T any] struct {
	basePtr unsafe.Pointer // Base pointer to the start of the allocated memory block
	capacity int          // Total number of T elements this arena can hold
	offset   int          // Current offset in number of T elements
	elementSize uintptr  // Size of a single T element in bytes
	elementAlign uintptr  // Alignment of a single T element in bytes
}

// NewArena initializes a new memory arena for type T with a given capacity.
// It allocates a raw byte slice and stores its pointer.
func NewArena[T any](capacity int) (*Arena[T], error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("capacity must be positive")
	}

	var zeroT T
	elementSize := unsafe.Sizeof(zeroT)
	elementAlign := unsafe.Alignof(zeroT)

	// Allocate a byte slice to hold the raw memory.
	// This slice itself is GC-managed, but the memory *within* it
	// is managed manually by the arena for allocations.
	// We ensure enough space for alignment padding if needed.
	totalBytes := uintptr(capacity)*elementSize + elementAlign - 1 // Add padding for alignment
	rawBytes := make([]byte, totalBytes)

	// Get the base pointer to the allocated memory.
	// We align it to the element's required alignment.
	basePtr := unsafe.Pointer(&rawBytes[0])
	alignedBasePtr := unsafe.Pointer((uintptr(basePtr) + elementAlign - 1) &^ (elementAlign - 1))

	fmt.Printf("[Arena] Initialized for type %T, capacity %d. Total raw bytes: %d\n", zeroT, capacity, len(rawBytes))
	fmt.Printf("[Arena] Base pointer: %p, Aligned base pointer: %p\n", basePtr, alignedBasePtr)

	return &Arena[T]{
		basePtr: alignedBasePtr,
		capacity: capacity,
		offset: 0,
		elementSize: elementSize,
		elementAlign: elementAlign,
	}, nil
}

// Alloc returns a pointer to a new element of type T from the arena.
// It does not involve the Go garbage collector for this allocation.
func (a *Arena[T]) Alloc() (*T, error) {
	if a.offset >= a.capacity {
		return nil, fmt.Errorf("arena capacity exceeded")
	}

	// Calculate the memory address for the next element.
	// unsafe.Add performs pointer arithmetic.
	elementPtr := unsafe.Add(a.basePtr, uintptr(a.offset)*a.elementSize)

	// Convert the unsafe.Pointer to a typed pointer *T.
	typedPtr := (*T)(elementPtr)

	a.offset++
	return typedPtr, nil
}

// Reset clears the arena, making all previously allocated slots available again.
// This is a "zero-cost" deallocation from the GC perspective.
func (a *Arena[T]) Reset() {
	a.offset = 0
	fmt.Printf("[Arena] Reset. All %d elements are now free for reuse.\n", a.capacity)
}

// CurrentUsage returns the number of elements currently "allocated".
func (a *Arena[T]) CurrentUsage() int {
	return a.offset
}

// GetCapacity returns the total capacity of the arena.
func (a *Arena[T]) GetCapacity() int {
	return a.capacity
}

// Validate ensures the allocated pointer is within the arena's bounds.
// This is a safety check not typically done in production, but useful for debugging.
func (a *Arena[T]) Validate(ptr *T) bool {
	p := uintptr(unsafe.Pointer(ptr))
	start := uintptr(a.basePtr)
	end := start + uintptr(a.capacity)*a.elementSize
	return p >= start && p < end && (p-start)%a.elementSize == 0
}

// Example struct for demonstration
type MyData struct {
	ID   uint64
	Name [16]byte // Fixed size to avoid inner allocations
	Value float64
}

func (d *MyData) String() string {
	return fmt.Sprintf("ID: %d, Name: %s, Value: %.2f", d.ID, string(d.Name[:]), d.Value)
}

// Helper to convert byte array to string without allocation (unsafe)
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// Helper to convert string to byte array without allocation (unsafe)
func StringToBytes(s string) []byte {
	strHdr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	sliceHdr := &reflect.SliceHeader{
		Data: strHdr.Data,
		Len:  strHdr.Len,
		Cap:  strHdr.Len,
	}
	return *(*[]byte)(unsafe.Pointer(sliceHdr))
}
