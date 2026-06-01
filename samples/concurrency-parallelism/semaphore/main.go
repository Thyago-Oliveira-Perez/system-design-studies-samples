package main

import (
	"fmt"
	"sync"
	"time"
)

// Semaphore: limits how many goroutines can use a resource at the same time.
// A mutex allows exactly one; a semaphore allows up to N.
//
// Analogy: the grill has room for 3 items at once, but there are 10 items to
// cook. A buffered channel of capacity 3 acts as the semaphore: sending into
// it claims a slot, receiving from it frees a slot. When all 3 slots are
// taken, the next cook blocks until one finishes — so at most 3 items cook
// in parallel.

// grill cooks one item; the caller is responsible for holding a slot.
func grill(item int) {
	fmt.Printf("Grilling item %d...\n", item)
	time.Sleep(2 * time.Second)
	fmt.Printf("Item %d is done, freeing a grill slot.\n", item)
}

func main() {
	var wg sync.WaitGroup

	const grillCapacity = 3 // max items cooking at once
	const items = 10

	// Buffered channel used as a counting semaphore.
	slots := make(chan struct{}, grillCapacity)

	for i := 1; i <= items; i++ {
		wg.Add(1)
		slots <- struct{}{} // claim a slot (blocks when the grill is full)

		go func(item int) {
			defer wg.Done()
			grill(item)
			<-slots // free the slot for the next item
		}(i)
	}

	wg.Wait()
	fmt.Println("Every item is grilled :)")
}
