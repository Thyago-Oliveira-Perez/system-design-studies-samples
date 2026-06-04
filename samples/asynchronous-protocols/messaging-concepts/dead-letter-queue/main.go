package main

import "fmt"

// Dead-letter queue (DLQ) — a fallback for messages that can't be processed.
//
// A consumer retries a failing message up to a limit. Once it has burned through
// its retries (or its TTL expires), the message is moved aside to a *dead-letter
// queue* instead of blocking the main queue forever. The DLQ keeps the poison
// message for later inspection while healthy traffic keeps flowing.
//
// Analogy: a food order the cook just can't make (out of an ingredient). After
// a few tries it's set aside on a "problem orders" tray so the line keeps
// moving, and someone looks at the tray later.
//
//	main queue -> process -> ok?  ── yes ─> done
//	                          │
//	                          └── no, retries left ─> requeue
//	                          └── no, max retries   ─> dead-letter queue
//
// Try it:
//
//	go run ./samples/asynchronous-protocols/messaging-concepts/dead-letter-queue

const maxRetries = 3

// Message carries a payload plus how many times delivery has been attempted.
type Message struct {
	body     string
	attempts int
}

// process succeeds for normal orders but always fails for the poison message,
// simulating a permanently un-processable record.
func process(m Message) error {
	if m.body == "poison" {
		return fmt.Errorf("cannot process %q", m.body)
	}
	return nil
}

func main() {
	queue := []Message{
		{body: "order-1"},
		{body: "poison"}, // will never succeed
		{body: "order-2"},
	}
	var dlq []Message

	// Process the queue, requeuing failures until they exhaust their retries.
	for len(queue) > 0 {
		m := queue[0]
		queue = queue[1:]

		err := process(m)
		if err == nil {
			fmt.Printf("ok       %-8s (attempt %d)\n", m.body, m.attempts+1)
			continue
		}

		m.attempts++
		if m.attempts >= maxRetries {
			fmt.Printf("dead     %-8s after %d attempts: %v\n", m.body, m.attempts, err)
			dlq = append(dlq, m)
			continue
		}

		fmt.Printf("retry    %-8s (attempt %d failed)\n", m.body, m.attempts)
		queue = append(queue, m) // back of the queue for another try
	}

	fmt.Println()
	fmt.Printf("dead-letter queue holds %d message(s):\n", len(dlq))
	for _, m := range dlq {
		fmt.Printf("  %s (gave up after %d attempts)\n", m.body, m.attempts)
	}
}
