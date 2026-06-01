package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Parallelism = doing many things at the exact same instant, on different
// CPU cores.
//
// Analogy: the barbecue has a long list of chores. Instead of one cook,
// several friends help out — one friend per available CPU core. We split the
// chore list into roughly equal slices and hand one slice to each friend, so
// the chores are genuinely worked on in parallel.

type Chore struct {
	Name    string
	Minutes int
	Friend  int // which friend (CPU) handled this chore
}

// work simulates one friend going through their slice of chores and
// reporting each finished chore on the barbecue channel.
func work(chores []Chore, barbecue chan<- Chore, friend int, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, chore := range chores {
		chore.Friend = friend
		fmt.Printf("Friend %d started %s...\n", friend, chore.Name)
		time.Sleep(time.Duration(chore.Minutes) * time.Second)
		barbecue <- chore
	}
}

func main() {
	barbecue := make(chan Chore)
	var wg sync.WaitGroup

	// Number of CPU cores = number of friends available to help.
	friends := runtime.NumCPU()
	fmt.Printf("Friends (CPU cores) available: %d\n", friends)

	chores := []Chore{
		{"steak", 5, 0},
		{"ribs", 7, 0},
		{"sausage", 3, 0},
		{"salad", 2, 0},
		{"chill the beer", 1, 0},
		{"organize the fridge", 1, 0},
		{"cheese", 3, 0},
		{"caipirinha", 2, 0},
		{"pork belly", 4, 0},
		{"skewers", 3, 0},
		{"grilled pineapple", 3, 0},
		{"clean the pool", 1, 0},
		{"sauces", 2, 0},
		{"garlic bread", 4, 0},
		{"rice", 4, 0},
		{"farofa", 4, 0},
	}
	fmt.Printf("Total chores: %d\n", len(chores))

	// Split the chores into one slice per friend. We round the slice size up
	// so no chore is left out and every friend gets a fair share.
	sliceSize := (len(chores) + friends - 1) / friends
	fmt.Printf("Chores per friend: %d\n\n", sliceSize)

	friend := 0
	for i := 0; i < len(chores); i += sliceSize {
		end := i + sliceSize
		if end > len(chores) {
			end = len(chores)
		}
		friend++
		wg.Add(1)
		go work(chores[i:end], barbecue, friend, &wg)
	}

	go func() {
		wg.Wait()
		close(barbecue)
	}()

	for chore := range barbecue {
		fmt.Printf("Friend %d finished %s.\n", chore.Friend, chore.Name)
	}
}
