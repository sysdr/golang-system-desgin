package main

import (
	"fmt"
	"log"
	"unsafe" // Required for pointer manipulation in demo

	"memory_arena_lesson/arena" // Import our arena package
)

func main() {
	fmt.Println("--- Manual Memory Arena Demonstration ---")

	// 1. Initialize an Arena for MyData structs
	const arenaCapacity = 10
	dataArena, err := arena.NewArena[arena.MyData](arenaCapacity)
	if err != nil {
		log.Fatalf("Failed to create arena: %v", err)
	}

	fmt.Printf("\nArena created with capacity: %d\n", dataArena.GetCapacity())

	// 2. Allocate and populate 5 MyData structs
	fmt.Println("nAllocating 5 MyData structs:")
	allocatedData := make([]*arena.MyData, 0, 5)
	for i := 0; i < 5; i++ {
		data, err := dataArena.Alloc()
		if err != nil {
			log.Fatalf("Error allocating data: %v", err)
		}
		data.ID = uint64(100 + i)
		name := fmt.Sprintf("Item-%d", i)
		copy(data.Name[:], name)
		data.Value = float64(i) * 1.5
		allocatedData = append(allocatedData, data)
			fmt.Printf("  Allocated: %s (Address: %p)\n", data.String(), data)
	}

	// 3. Print the contents of the allocated MyData structs (already done above)
	fmt.Printf("\nCurrent arena usage: %d/%d\n", dataArena.CurrentUsage(), dataArena.GetCapacity())

	// 4. Attempt to allocate an 11th MyData struct to observe capacity exceeded
	fmt.Println("nAttempting to allocate an 11th MyData struct (should fail if capacity is 10):")
	for i := 5; i < arenaCapacity+1; i++ { // Try to allocate up to capacity + 1
		data, err := dataArena.Alloc()
		if err != nil {
			fmt.Printf("  Error allocating data for item %d: %v\n", i, err)
			break // Stop on error
		}
		data.ID = uint64(100 + i)
		name := fmt.Sprintf("Item-%d", i)
		copy(data.Name[:], name)
		data.Value = float64(i) * 1.5
		allocatedData = append(allocatedData, data)
			fmt.Printf("  Allocated: %s (Address: %p)\n", data.String(), data)
	}
	fmt.Printf("Current arena usage after exceeding capacity attempt: %d/%d\n", dataArena.CurrentUsage(), dataArena.GetCapacity())


	// 5. Reset the arena
	fmt.Println("nResetting the arena...")
	dataArena.Reset()
	fmt.Printf("Current arena usage after reset: %d/%d\n", dataArena.CurrentUsage(), dataArena.GetCapacity())


	// 6. Allocate 3 new MyData structs and populate them
	fmt.Println("nAllocating 3 new MyData structs after reset:")
	newAllocatedData := make([]*arena.MyData, 0, 3)
	for i := 0; i < 3; i++ {
		data, err := dataArena.Alloc()
		if err != nil {
			log.Fatalf("Error allocating data after reset: %v", err)
		}
		data.ID = uint64(200 + i)
		name := fmt.Sprintf("ResetItem-%d", i)
		copy(data.Name[:], name)
		data.Value = float64(i) * 2.0
		newAllocatedData = append(newAllocatedData, data)
			fmt.Printf("  Allocated: %s (Address: %p)\n", data.String(), data)
	}
	fmt.Printf("Current arena usage: %d/%d\n", dataArena.CurrentUsage(), dataArena.GetCapacity())

	// Demonstrate unsafe string/byte conversion helpers
	fmt.Println("\n--- Unsafe String/Byte Conversion Demo ---")
	testString := "Hello, Unsafe World!"
	fmt.Printf("Original string: %s (Address: %p)\n", testString, &testString)
	
	bytes := arena.StringToBytes(testString)
	fmt.Printf("Converted to bytes (unsafe): %s (Address: %p)\n", string(bytes), &bytes) // &bytes is address of slice header
	if len(bytes) > 0 {
		fmt.Printf("  Underlying data pointer (bytes): %p\n", unsafe.Pointer(&bytes[0]))
	} else {
		fmt.Printf("  Underlying data pointer (bytes): <empty slice>\n")
	}

	reconvertedString := arena.BytesToString(bytes)
	fmt.Printf("Reconverted to string (unsafe): %s (Address: %p)\n", reconvertedString, &reconvertedString)
	
	// This is a tricky part to get the string data pointer reliably across Go versions without copying.
	// For Go 1.20+, unsafe.StringData is available. For simplicity here, we can infer.
	// The key is that the underlying memory for 'testString', 'bytes', and 'reconvertedString' *should* be the same.
	// If you want to confirm, you can add print statements inside the unsafe helpers to show the Data field of the headers.
	fmt.Println("Note: For 'unsafe' string/byte conversion, the underlying data is shared, not copied.")

}
