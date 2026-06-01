package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Spinlock: a lock where a waiting goroutine "spins" in a busy loop, actively
// retrying until the lock is free, instead of being put to sleep like a mutex.
//
// Analogy: only one friend can use the grill at a time. A waiting friend keeps
// checking "is it free yet? is it free yet?" in a tight loop. We call
// runtime.Gosched() inside the loop to let the Go scheduler run other
// goroutines while we spin, so we don't completely hog the CPU.
//
// Spinlocks are only worth it when the wait is expected to be extremely short.
// For anything longer, a sync.Mutex (which parks the goroutine) is better.

type SpinLock struct {
	state int32 // 0 = free, 1 = held
}

// Lock spins until it manages to flip the state from 0 to 1 atomically.
func (s *SpinLock) Lock() {
	for !atomic.CompareAndSwapInt32(&s.state, 0, 1) {
		runtime.Gosched() // give other goroutines a chance to run
	}
}

// Unlock sets the state back to 0 so a spinning goroutine can grab it.
func (s *SpinLock) Unlock() {
	atomic.StoreInt32(&s.state, 0)
}

func grill(friend int, lock *SpinLock, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("Friend %d is waiting for the grill\n", friend)
	lock.Lock()

	fmt.Printf("Friend %d is grilling\n", friend)
	time.Sleep(1 * time.Second)

	fmt.Printf("Friend %d is done with the grill\n", friend)
	lock.Unlock()
}

func main() {
	var wg sync.WaitGroup
	var lock SpinLock

	const friends = 10

	for i := 1; i <= friends; i++ {
		wg.Add(1)
		go grill(i, &lock, &wg)
	}

	wg.Wait()
	fmt.Println("The barbecue is over :)")
}
