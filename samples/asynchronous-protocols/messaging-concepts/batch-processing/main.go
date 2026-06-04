package main

import (
	"fmt"
	"time"
)

// Batch processing — accumulate data at rest, then process a whole lot at once.
//
// One of the original reasons asynchronous communication exists: instead of
// handling every message the instant it arrives, you let them pile up and flush
// the batch on a trigger. Two triggers are classic and used together here:
//
//   - size — the batch is full (enough items accumulated), or
//   - time — the oldest buffered item has waited long enough, so we flush a
//     partial batch rather than let it go stale.
//
// Batching trades a little latency for much higher throughput (one expensive
// flush — a DB write, an API call — amortized over many items).
//
// Analogy: the cook doesn't fire up the grill for a single skewer. They wait
// until there are enough orders (size) — but never let early orders go cold,
// so they also fire on a timer (time).
//
// Try it:
//
//	go run ./samples/asynchronous-protocols/messaging-concepts/batch-processing

const (
	batchSize = 3                      // flush once this many items accumulate
	maxWait   = 300 * time.Millisecond // ...but never let the oldest item wait longer
)

// flush is the "expensive" operation we want to amortize across a batch.
func flush(reason string, batch []string) {
	fmt.Printf("flush (%s): %v\n", reason, batch)
}

func main() {
	// Incoming messages arrive at an uneven pace, like real traffic: three in
	// quick succession (fills a batch by size), then a couple followed by a
	// long idle gap (flushed by the timer), then a final straggler left in the
	// buffer when the producer stops (flushed by the drain on close).
	incoming := make(chan string)
	go func() {
		gaps := []time.Duration{40, 40, 40, 80, 80, 500, 40, 40, 40}
		for i, gap := range gaps {
			time.Sleep(gap * time.Millisecond)
			incoming <- fmt.Sprintf("msg-%d", i+1)
		}
		close(incoming)
	}()

	var batch []string

	// The timer measures the age of the *oldest* buffered item. We keep it
	// stopped while the batch is empty: a nil channel blocks forever in select,
	// so the timer can never fire spuriously and reset our timing.
	timer := time.NewTimer(maxWait)
	timer.Stop()
	var timerC <-chan time.Time

	// disarm stops the age timer and removes it from the select.
	disarm := func() {
		timer.Stop()
		timerC = nil
	}

	for {
		select {
		case msg, ok := <-incoming:
			if !ok {
				// Producer done: flush whatever is left and exit.
				if len(batch) > 0 {
					flush("drain", batch)
				}
				return
			}
			if len(batch) == 0 {
				// First item of a new batch: start its age clock.
				timer.Reset(maxWait)
				timerC = timer.C
			}
			batch = append(batch, msg)
			if len(batch) == batchSize {
				flush("size", batch)
				batch = nil
				disarm()
			}

		case <-timerC:
			// The oldest item hit maxWait before the batch filled up — flush the
			// partial batch so early messages don't sit waiting indefinitely.
			flush("time", batch)
			batch = nil
			disarm()
		}
	}
}
