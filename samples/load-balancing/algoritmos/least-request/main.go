package main

import (
	"fmt"
	"log"
	"sync/atomic"
)

// Server tracks the total number of requests it has served since startup.
// This is a cumulative, historical counter — it never goes down.
type Server struct {
	name     string
	host     string
	requests atomic.Int64
}

// LeastRequest routes each new request to the server that has handled
// the fewest requests overall. It's a long-term balancing strategy:
// servers that had more traffic in the past will receive less going forward.
//
// Key difference from LeastConnections: the counter is cumulative.
// It measures total work ever done, not current load.
type LeastRequest struct {
	servers []*Server
}

// Next returns the server with the lowest total request count.
func (lr *LeastRequest) Next() *Server {
	var chosen *Server
	for _, s := range lr.servers {
		if chosen == nil || s.requests.Load() < chosen.requests.Load() {
			chosen = s
		}
	}
	return chosen
}

// counts returns a snapshot of total request counts for display.
func counts(servers []*Server) string {
	return fmt.Sprintf("A=%d B=%d C=%d",
		servers[0].requests.Load(),
		servers[1].requests.Load(),
		servers[2].requests.Load(),
	)
}

func main() {
	servers := []*Server{
		{name: "Server A", host: "10.0.0.1:8080"},
		{name: "Server B", host: "10.0.0.2:8080"},
		{name: "Server C", host: "10.0.0.3:8080"},
	}

	// Pre-load Server A to simulate a historical imbalance (e.g. it was the only
	// server during a scale-out event). LeastRequest should compensate by routing
	// new traffic away from it until the others catch up.
	servers[0].requests.Store(10)

	lb := &LeastRequest{servers: servers}

	log.Println("=== Least Request Load Balancer ===")
	log.Printf("Backend pool: %d servers", len(servers))
	log.Println("Starting with imbalance: Server A already has 10 requests.")
	fmt.Println()

	for i := 1; i <= 20; i++ {
		server := lb.Next()
		server.requests.Add(1)
		log.Printf("Request %2d → %s | totals [%s]", i, server.name, counts(servers))
	}

	fmt.Println()
	log.Println("=== Final Totals ===")
	for _, s := range servers {
		log.Printf("  %s: %d requests", s.name, s.requests.Load())
	}
	log.Println("Servers B and C absorbed the early traffic to compensate for A's head start.")
}
