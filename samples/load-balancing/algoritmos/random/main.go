package main

import (
	"fmt"
	"log"
	"math/rand"
)

// Server represents a backend server in the pool.
type Server struct {
	name string
	host string
}

// Random picks a server at random for each incoming request.
// It requires no shared state between requests, which makes it trivially
// safe for concurrent use (rand is goroutine-safe since Go 1.20).
type Random struct {
	servers []Server
}

// Next returns a randomly selected server from the pool.
func (r *Random) Next() Server {
	idx := rand.Intn(len(r.servers))
	return r.servers[idx]
}

func main() {
	servers := []Server{
		{name: "Server A", host: "10.0.0.1:8080"},
		{name: "Server B", host: "10.0.0.2:8080"},
		{name: "Server C", host: "10.0.0.3:8080"},
	}

	lb := &Random{servers: servers}
	// Track how many requests each server received
	counts := make(map[string]int)

	log.Println("=== Random Load Balancer ===")
	log.Printf("Backend pool: %d servers\n", len(servers))
	fmt.Println()

	total := 30
	for i := 1; i <= total; i++ {
		server := lb.Next()
		counts[server.name]++
		log.Printf("Request %2d → %s (%s)", i, server.name, server.host)
	}

	// Show the final distribution — it won't be perfectly even, and that's expected.
	// With large enough traffic the law of large numbers brings it close to uniform.
	fmt.Println()
	log.Println("=== Distribution Summary ===")
	for _, s := range servers {
		pct := float64(counts[s.name]) / float64(total) * 100
		log.Printf("  %s: %d requests (%.1f%%)", s.name, counts[s.name], pct)
	}
	log.Printf("  Expected even split: %.1f%% per server", 100.0/float64(len(servers)))
}
