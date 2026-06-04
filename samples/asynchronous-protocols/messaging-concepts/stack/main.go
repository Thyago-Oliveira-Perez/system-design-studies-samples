package main

import "fmt"

// LIFO stack — Last In, First Out.
//
// Like the queue, a stack is a *data structure* before it is any technology.
// The newest item is the first one served, so older items can wait a long time
// (or forever) under constant load — the opposite ordering of a FIFO queue.
//
// Analogy: clean plates are stacked at the barbecue. You take from the top, so
// the plate placed *last* is the one used *first*.
//
//	Push -> [ Pizza, Hamburger, Churrasco ] -> Pop (Churrasco leaves first)
//
// Try it:
//
//	go run ./samples/asynchronous-protocols/messaging-concepts/stack

// Stack is a generic LIFO collection.
type Stack[T any] struct {
	items []T
}

// Push places an item on top of the stack.
func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

// Pop removes and returns the item on top — the most recently pushed one.
// The bool is false when the stack is empty.
func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	top := len(s.items) - 1
	item := s.items[top]
	s.items = s.items[:top]
	return item, true
}

func main() {
	stack := &Stack[string]{}

	plates := []string{"Pizza", "Hamburger", "Churrasco"}
	for _, plate := range plates {
		fmt.Println("Push:", plate)
		stack.Push(plate)
	}

	fmt.Println()

	// Drain the stack: items come out in reverse order (newest first).
	for {
		item, ok := stack.Pop()
		if !ok {
			break
		}
		fmt.Println("Pop:", item)
	}
}
