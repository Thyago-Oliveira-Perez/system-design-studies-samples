package main

import (
	"fmt"
	"sync"
	"time"
)

// Concurrency = dealing with many things at once.
//
// Analogy: we are hosting a barbecue. Several items need to be prepared
// (steak, ribs, sausage...). One cook would do them one after another, but
// here we start every item "at the same time" using goroutines. While one
// item is resting (time.Sleep), the cook can make progress on the others.
//
// Each item reports back through a channel as soon as it is ready, and a
// WaitGroup lets us know when every goroutine has finished so we can close
// the channel.

type Item struct {
	Name    string
	Minutes int // how long this item takes to prepare (simulated)
}

// prepare simulates the time it takes to get one item ready and then
// announces it on the barbecue channel.
func prepare(item Item, barbecue chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Preparing %s...\n", item.Name)
	time.Sleep(time.Duration(item.Minutes) * time.Second)
	barbecue <- item.Name
}

func main() {
	// Channel that collects every item as it becomes ready.
	barbecue := make(chan string)

	// WaitGroup so we know when all goroutines are done.
	var wg sync.WaitGroup

	items := []Item{
		{"steak", 5},
		{"ribs", 7},
		{"sausage", 3},
		{"salad", 2},
		{"drinks", 1},
		{"grill setup", 2},
		{"cheese", 3},
	}

	// Start preparing every item concurrently.
	for _, item := range items {
		wg.Add(1)
		go prepare(item, barbecue, &wg)
	}

	// Close the channel once every goroutine has reported in. This runs in
	// its own goroutine so the range loop below can keep draining the channel.
	go func() {
		wg.Wait()
		close(barbecue)
		fmt.Println("\nThe barbecue is over :)")
	}()

	// Drain ready items until the channel is closed.
	for name := range barbecue {
		fmt.Printf("%s is ready.\n", name)
	}
}
