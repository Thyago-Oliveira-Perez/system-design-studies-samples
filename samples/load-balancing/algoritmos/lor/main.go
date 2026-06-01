package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Server tracks the number of requests currently being processed (in-flight).
// The counter goes up when a request arrives and down when it finishes.
//
// LOR operates at the HTTP request level, not the TCP connection level.
// A single keep-alive connection (or HTTP/2 stream) can carry many concurrent
// requests — LOR sees each of them individually, making it more granular than
// LeastConnections.
type Server struct {
	name     string
	host     string
	inflight atomic.Int32
}

// LOR (Least Outstanding Requests) routes each request to the server with
// the fewest requests currently in-flight.
//
// It reacts faster than LeastConnections to slow backends because it tracks
// request completions rather than connection teardowns.
// Envoy uses this algorithm as its default for HTTP/2 and gRPC upstreams.
type LOR struct {
	servers []*Server
	mu      sync.Mutex // protects the selection so the snapshot is consistent
}

// Next returns the server with the lowest in-flight request count.
func (l *LOR) Next() *Server {
	l.mu.Lock()
	defer l.mu.Unlock()

	var chosen *Server
	for _, s := range l.servers {
		if chosen == nil || s.inflight.Load() < chosen.inflight.Load() {
			chosen = s
		}
	}
	return chosen
}

// state returns a human-readable snapshot of in-flight counts.
func state(servers []*Server) string {
	return fmt.Sprintf("A=%d B=%d C=%d",
		servers[0].inflight.Load(),
		servers[1].inflight.Load(),
		servers[2].inflight.Load(),
	)
}

func main() {
	servers := []*Server{
		{name: "Server A", host: "10.0.0.1:8080"},
		{name: "Server B", host: "10.0.0.2:8080"},
		{name: "Server C", host: "10.0.0.3:8080"},
	}

	// Server B is intentionally slow to simulate a degraded backend.
	// LOR should detect the backlog and route fewer new requests to it.
	processingMs := map[string]int{
		"Server A": 80,  // fast
		"Server B": 500, // slow / degraded
		"Server C": 150, // medium
	}

	lb := &LOR{servers: servers}

	log.Println("=== LOR (Least Outstanding Requests) Load Balancer ===")
	log.Printf("Backend pool: %d servers", len(servers))
	log.Println("Server B is intentionally slow (500ms) — watch LOR route around it.")
	fmt.Println()

	var wg sync.WaitGroup

	for i := 1; i <= 18; i++ {
		wg.Add(1)
		go func(reqID int) {
			defer wg.Done()

			server := lb.Next()
			server.inflight.Add(1)

			base := processingMs[server.name]
			// Small jitter makes the simulation feel more realistic
			duration := time.Duration(base+rand.Intn(40)) * time.Millisecond

			log.Printf("→ req #%2d → %s | in-flight [%s] | ~%dms",
				reqID, server.name, state(servers), duration.Milliseconds())

			time.Sleep(duration)

			server.inflight.Add(-1)
			log.Printf("← req #%2d ← %s | in-flight [%s]",
				reqID, server.name, state(servers))
		}(i)

		// Requests arrive every 60ms — faster than Server B can finish them,
		// so its in-flight count grows and LOR steers new traffic elsewhere.
		time.Sleep(60 * time.Millisecond)
	}

	wg.Wait()
	fmt.Println()
	log.Println("Server B accumulated in-flight requests and received fewer new ones over time.")
	log.Println("This is LOR's key strength: automatic detection of slow or degraded backends.")
}
