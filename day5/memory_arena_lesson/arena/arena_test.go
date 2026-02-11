package arena_test

import (
	"testing"
	"memory_arena_lesson/arena"
)

// BenchmarkArenaAlloc measures the performance of allocating MyData structs using the custom arena.
func BenchmarkArenaAlloc(b *testing.B) {
	// Initialize an arena with enough capacity for all allocations in the benchmark.
	a, err := arena.NewArena[arena.MyData](b.N)
	if err != nil {
		b.Fatalf("Failed to create arena: %v", err)
	}
	b.ResetTimer() // Reset timer to exclude setup time
	for i := 0; i < b.N; i++ {
		_, _ = a.Alloc() // Allocate a MyData struct from the arena
	}
}

// BenchmarkNewAlloc measures the performance of allocating MyData structs using Go's built-in new().
func BenchmarkNewAlloc(b *testing.B) {
	b.ResetTimer() // Reset timer to exclude setup time
	for i := 0; i < b.N; i++ {
		_ = new(arena.MyData) // Allocate a MyData struct using new()
	}
}

// BenchmarkArenaAllocAndFill measures the performance of allocating and filling MyData structs using the custom arena.
func BenchmarkArenaAllocAndFill(b *testing.B) {
	a, err := arena.NewArena[arena.MyData](b.N)
	if err != nil {
		b.Fatalf("Failed to create arena: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := a.Alloc()
		data.ID = uint64(i)
		copy(data.Name[:], "testname")
		data.Value = float64(i)
	}
}

// BenchmarkNewAllocAndFill measures the performance of allocating and filling MyData structs using Go's built-in new().
func BenchmarkNewAllocAndFill(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := new(arena.MyData)
		data.ID = uint64(i)
		copy(data.Name[:], "testname")
		data.Value = float64(i)
	}
}
