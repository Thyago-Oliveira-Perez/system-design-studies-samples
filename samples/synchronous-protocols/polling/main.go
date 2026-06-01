// Polling: the client repeatedly asks the server "are you done yet?" until the
// answer changes. It's the simplest way to learn about an event when the
// server has no way to push to the client.
//
// Scenario: a client submits a long-running job (e.g. "generate a report").
// The server answers immediately with a job ID and status "pending". The
// client then *polls* the status endpoint on a fixed interval until the job
// reports "done".
//
// Trade-off shown here: polling is easy and works everywhere, but it wastes
// requests — most of them just say "still pending". A shorter interval reacts
// faster but burns more requests; a longer one is cheaper but adds latency.
// (Webhooks and WebSockets — the sibling samples — solve this by pushing.)
//
// This program runs the server and the client in one process so a single
// `go run` shows the whole exchange.
//
// Run:
//
//	go run ./samples/synchronous-protocols/polling
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

type job struct {
	mu     sync.Mutex
	status string // "pending" -> "done"
}

func main() {
	j := &job{status: "pending"}

	// The server: one endpoint to submit work, one to check its status.
	mux := http.NewServeMux()

	mux.HandleFunc("POST /jobs", func(w http.ResponseWriter, r *http.Request) {
		// Pretend the work takes 3 seconds, done in the background.
		go func() {
			time.Sleep(3 * time.Second)
			j.mu.Lock()
			j.status = "done"
			j.mu.Unlock()
		}()
		json.NewEncoder(w).Encode(map[string]string{"id": "job-1", "status": "pending"})
	})

	mux.HandleFunc("GET /jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		j.mu.Lock()
		status := j.status
		j.mu.Unlock()
		json.NewEncoder(w).Encode(map[string]string{"id": r.PathValue("id"), "status": status})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	// The client: submit the job, then poll until it is done.
	fmt.Println("client: submitting job...")
	http.Post(srv.URL+"/jobs", "application/json", nil)

	const interval = 500 * time.Millisecond
	attempts := 0
	for {
		attempts++
		resp, err := http.Get(srv.URL + "/jobs/job-1")
		if err != nil {
			log.Fatal(err)
		}
		var body map[string]string
		json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()

		fmt.Printf("client: poll #%d -> status=%q\n", attempts, body["status"])
		if body["status"] == "done" {
			break
		}
		time.Sleep(interval)
	}

	fmt.Printf("\nclient: job finished after %d polls.\n", attempts)
	fmt.Printf("Notice: %d of those requests returned \"pending\" — wasted work that\n", attempts-1)
	fmt.Println("a push-based protocol (webhooks, websockets) would avoid.")
}
