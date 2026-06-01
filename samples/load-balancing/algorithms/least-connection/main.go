package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Server tracks its number of currently open connections.
// atomic.Int32 is used so reads/writes are safe across goroutines without a mutex.
type Server struct {
	name        string
	host        string
	connections atomic.Int32
}

// LeastConnections routes each new request to the server with the fewest
// active connections at the moment of routing.
// It reads a snapshot of all counters inside a mutex so the comparison is consistent.
type LeastConnections struct {
	servers []*Server
	mu      sync.Mutex // guards the selection decision, not the counters themselves
}

// Next scans all servers and returns the one with the smallest connection count.
func (lc *LeastConnections) Next() *Server {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	var chosen *Server
	for _, s := range lc.servers {
		if chosen == nil || s.connections.Load() < chosen.connections.Load() {
			chosen = s
		}
	}
	return chosen
}

// state returns a human-readable snapshot of all server connection counts.
func state(servers []*Server) string {
	return fmt.Sprintf("A=%d B=%d C=%d",
		servers[0].connections.Load(),
		servers[1].connections.Load(),
		servers[2].connections.Load(),
	)
}

func main() {
	servers := []*Server{
		{name: "Server A", host: "10.0.0.1:8080"},
		{name: "Server B", host: "10.0.0.2:8080"},
		{name: "Server C", host: "10.0.0.3:8080"},
	}

	lb := &LeastConnections{servers: servers}

	log.Println("=== Least Connection Load Balancer ===")
	log.Printf("Backend pool: %d servers", len(servers))
	log.Println("Processing times are random (100–500ms) — watch how busy servers receive fewer new requests.")
	fmt.Println()

	var wg sync.WaitGroup

	for i := 1; i <= 12; i++ {
		wg.Add(1)
		go func(reqID int) {
			defer wg.Done()

			// Route at this instant, then increment before anyone else can read the count
			server := lb.Next()
			server.connections.Add(1)

			duration := time.Duration(100+rand.Intn(400)) * time.Millisecond
			log.Printf("→ req #%2d assigned to %s | connections [%s] | will take %dms",
				reqID, server.name, state(servers), duration.Milliseconds())

			time.Sleep(duration)

			server.connections.Add(-1)
			log.Printf("← req #%2d done on    %s | connections [%s]",
				reqID, server.name, state(servers))
		}(i)

		// Stagger arrivals slightly so the routing decisions are visible in the log
		time.Sleep(80 * time.Millisecond)
	}

	wg.Wait()
	fmt.Println()
	log.Println("Servers that ran slow jobs ended up with higher connection counts")
	log.Println("and therefore received fewer new requests while they were busy.")
}
