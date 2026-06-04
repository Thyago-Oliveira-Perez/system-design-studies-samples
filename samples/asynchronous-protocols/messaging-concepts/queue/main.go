package main

import "fmt"

// FIFO queue — First In, First Out.
//
// A queue is a *data structure* before it is ever a technology (RabbitMQ, SQS,
// ZeroMQ...). It models the most natural ordering for asynchronous work: items
// are consumed in the exact order they were produced.
//
// Analogy: friends arrive at the barbecue and place food orders. The cook
// serves them in arrival order — whoever asked first eats first.
//
//	Enqueue -> [ Pizza, Hamburger, Churrasco ] -> Dequeue (Pizza leaves first)
//
// Try it:
//
//	go run ./samples/asynchronous-protocols/messaging-concepts/queue

// Queue is a generic FIFO collection.
type Queue[T any] struct {
	items []T
}

// Enqueue adds an item to the back of the queue (the producer side).
func (q *Queue[T]) Enqueue(item T) {
	q.items = append(q.items, item)
}

// Dequeue removes and returns the item at the front — the one that has been
// waiting longest. The bool is false when the queue is empty.
func (q *Queue[T]) Dequeue() (T, bool) {
	var zero T
	if len(q.items) == 0 {
		return zero, false
	}
	front := q.items[0]
	q.items = q.items[1:]
	return front, true
}

func main() {
	queue := &Queue[string]{}

	orders := []string{"Pizza", "Hamburger", "Churrasco"}
	for _, order := range orders {
		fmt.Println("Enqueue:", order)
		queue.Enqueue(order)
	}

	fmt.Println()

	// Drain the queue: items come out in the same order they went in.
	for {
		item, ok := queue.Dequeue()
		if !ok {
			break
		}
		fmt.Println("Dequeue:", item)
	}
}
