package main

import "fmt"

// Fanout — a 1:N delivery strategy.
//
// A single production is copied to *every* subscribed queue. Each consumer gets
// its own independent copy of the message, so they don't compete for it — the
// opposite of a work queue where one message is handled by exactly one worker.
//
// Analogy: when the churrasco is ready, the host shouts it once and *every*
// friend hears it. The announcement is duplicated to everyone, not handed to a
// single person.
//
//	                      ┌─> billing   queue: [order]
//	produce("order") ─────┼─> shipping  queue: [order]
//	                      └─> analytics queue: [order]
//
// Try it:
//
//	go run ./samples/asynchronous-protocols/messaging-concepts/fanout

// Queue is a minimal FIFO buffer standing in for one subscriber's mailbox.
type Queue struct {
	name  string
	items []string
}

func (q *Queue) enqueue(item string) {
	q.items = append(q.items, item)
}

// Exchange fans every published message out to all bound queues.
type Exchange struct {
	queues []*Queue
}

// bind subscribes a queue so it receives a copy of every future message.
func (e *Exchange) bind(q *Queue) {
	e.queues = append(e.queues, q)
}

// publish delivers one copy of msg to each bound queue.
func (e *Exchange) publish(msg string) {
	fmt.Printf("publish %q -> %d queues\n", msg, len(e.queues))
	for _, q := range e.queues {
		q.enqueue(msg)
	}
}

func main() {
	billing := &Queue{name: "billing"}
	shipping := &Queue{name: "shipping"}
	analytics := &Queue{name: "analytics"}

	exchange := &Exchange{}
	exchange.bind(billing)
	exchange.bind(shipping)
	exchange.bind(analytics)

	exchange.publish("order-42")
	exchange.publish("order-43")

	fmt.Println()

	// Every queue received an identical, independent copy of both messages.
	for _, q := range exchange.queues {
		fmt.Printf("%-9s got %v\n", q.name, q.items)
	}
}
