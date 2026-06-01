package main

import (
	"fmt"
	"hash/fnv"
	"log"
)

// Server represents a backend server in the pool.
type Server struct {
	name string
	host string
}

// IPHash maps each client IP to a specific server deterministically.
// The same IP always lands on the same server — this property is called
// "session affinity" or "sticky sessions".
type IPHash struct {
	servers []Server
}

// Next hashes the client IP address and maps it to a fixed server index.
// As long as the server pool doesn't change, the mapping is stable.
func (h *IPHash) Next(clientIP string) Server {
	// FNV (Fowler-Noll-Vo) is a fast, non-cryptographic hash.
	// We don't need security here — just a stable, well-distributed mapping.
	hasher := fnv.New32a()
	hasher.Write([]byte(clientIP))
	idx := int(hasher.Sum32()) % len(h.servers)
	return h.servers[idx]
}

func main() {
	servers := []Server{
		{name: "Server A", host: "10.0.0.1:8080"},
		{name: "Server B", host: "10.0.0.2:8080"},
		{name: "Server C", host: "10.0.0.3:8080"},
	}

	lb := &IPHash{servers: servers}

	// Some IPs appear multiple times — they should always hit the same server,
	// no matter what other requests came in between.
	requests := []struct {
		id       int
		clientIP string
	}{
		{1, "192.168.1.10"},
		{2, "192.168.1.20"},
		{3, "192.168.1.30"},
		{4, "192.168.1.10"}, // revisit — must go to the same server as request 1
		{5, "192.168.1.40"},
		{6, "192.168.1.20"}, // revisit — must go to the same server as request 2
		{7, "192.168.1.50"},
		{8, "192.168.1.10"}, // revisit again
		{9, "203.0.113.5"},
		{10, "198.51.100.1"},
	}

	log.Println("=== IP Hash Load Balancer ===")
	log.Printf("Backend pool: %d servers\n", len(servers))
	fmt.Println()

	for _, req := range requests {
		server := lb.Next(req.clientIP)
		log.Printf("Request %2d | Client %-15s → %s (%s)", req.id, req.clientIP, server.name, server.host)
	}

	fmt.Println()
	log.Println("Notice: 192.168.1.10 (requests 1, 4, 8) always routes to the same server.")
	log.Println("Notice: 192.168.1.20 (requests 2, 6) always routes to the same server.")
}
