package main

import (
	"fmt"
	"log"
)

// Server represents a backend server in the pool.
type Server struct {
	name string
	host string
}

// RoundRobin cycles through servers in order.
// Each incoming request gets the next server in the list.
// After reaching the last server it wraps back to the first.
// No server state is needed — just a counter.
type RoundRobin struct {
	servers []Server
	current int // index of the next server to use
}

// Next returns the next server in the rotation and advances the counter.
func (rr *RoundRobin) Next() Server {
	server := rr.servers[rr.current]
	// Modulo keeps the index within bounds and produces the wrap-around effect
	rr.current = (rr.current + 1) % len(rr.servers)
	return server
}

func main() {
	servers := []Server{
		{name: "Server A", host: "10.0.0.1:8080"},
		{name: "Server B", host: "10.0.0.2:8080"},
		{name: "Server C", host: "10.0.0.3:8080"},
	}

	lb := &RoundRobin{servers: servers}

	log.Println("=== Round Robin Load Balancer ===")
	log.Printf("Backend pool: %d servers\n", len(servers))
	fmt.Println()

	// 9 requests = 3 full cycles, making the rotation pattern obvious
	for i := 1; i <= 9; i++ {
		server := lb.Next()
		log.Printf("Request %d → %s (%s)", i, server.name, server.host)
	}

	fmt.Println()
	log.Println("Each server received exactly 3 requests — perfectly balanced.")
}
