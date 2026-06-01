package main

import (
	"fmt"
	"sync"
	"time"
)

// Mutex (mutual exclusion): the fix for the race condition.
//
// Analogy: there is only one grill. A cook must hold the grill (Lock) before
// touching the shared counter, and release it (Unlock) when done. While one
// cook holds the lock, every other cook waits their turn. The shared counter
// is now updated by exactly one goroutine at a time, so no increment is lost.
//
// Compare this file with ../race_condition: same program, one mutex added.

// Shared counter of grilled items.
var grilled int

// The single grill — only one cook can hold it at a time.
var grill sync.Mutex

func cook() {
	grill.Lock() // take the grill; other cooks wait here
	grilled++
	time.Sleep(time.Millisecond * 100)
	grill.Unlock() // hand the grill to the next cook
}

func main() {
	var wg sync.WaitGroup

	const items = 100

	for i := 0; i < items; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cook()
		}()
	}

	wg.Wait()

	fmt.Printf("Items grilled: %d (expected %d)\n", grilled, items)
}
