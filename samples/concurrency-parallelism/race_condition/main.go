package main

import (
	"fmt"
	"sync"
	"time"
)

// Race condition: two or more goroutines read and write the same variable at
// the same time without coordination, and updates get lost.
//
// Analogy: every cook at the barbecue grills an item and then bumps a shared
// counter of "grilled items". The counter increment (grilled++) is actually
// three steps under the hood — read, add one, write back. If two cooks do
// this at the same time they can read the same old value and one increment is
// silently lost. The final total ends up lower than the number of items.
//
// Run it with the race detector to see the data race reported:
//
//	go run -race .

// Shared counter of grilled items — deliberately unprotected.
var grilled int

func grill() {
	grilled++ // not atomic: read -> increment -> write, can be interrupted
	time.Sleep(time.Millisecond * 100)
}

func main() {
	var wg sync.WaitGroup

	const items = 100

	for i := 0; i < items; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			grill()
		}()
	}

	wg.Wait()

	fmt.Printf("Items grilled: %d (expected %d)\n", grilled, items)
	if grilled != items {
		fmt.Printf("%d increments were lost to the race condition.\n", items-grilled)
	} else {
		fmt.Println("No loss this run — try again or run with: go run -race .")
	}
}
