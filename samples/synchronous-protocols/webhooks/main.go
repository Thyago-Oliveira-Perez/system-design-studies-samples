// Webhooks: instead of the client polling the server, the server calls the
// client. The client registers a callback URL once; when an event happens the
// server makes an HTTP POST to that URL with the event payload.
//
// This is "reverse HTTP" or a push model:
//   - Polling  = client asks "anything new?" over and over   (see ../polling)
//   - Webhook  = server says "here's something new" when it happens
//
// Scenario: a payment provider lets a merchant register a webhook. When a
// payment is confirmed, the provider POSTs the event to the merchant's URL.
//
// This program runs both the provider and the subscriber (merchant) in one
// process so a single `go run` shows registration + delivery.
//
// Run:
//
//	go run ./samples/synchronous-protocols/webhooks
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

type Event struct {
	Type   string  `json:"type"`
	Order  string  `json:"order"`
	Amount float64 `json:"amount"`
}

// provider holds the URLs that subscribers asked to be notified on.
type provider struct {
	mu          sync.Mutex
	subscribers []string
}

func (p *provider) subscribe(url string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subscribers = append(p.subscribers, url)
	log.Printf("provider: registered webhook -> %s", url)
}

// emit pushes an event to every registered subscriber via HTTP POST.
func (p *provider) emit(ev Event) {
	p.mu.Lock()
	urls := append([]string(nil), p.subscribers...)
	p.mu.Unlock()

	payload, _ := json.Marshal(ev)
	for _, url := range urls {
		log.Printf("provider: delivering %q event to %s", ev.Type, url)
		resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
		if err != nil {
			log.Printf("provider: delivery to %s failed: %v", url, err)
			continue
		}
		resp.Body.Close()
	}
}

func main() {
	prov := &provider{}

	// ── Subscriber (the merchant) ──────────────────────────────────────────
	// A plain HTTP endpoint that receives event POSTs from the provider.
	received := make(chan Event, 1)
	subMux := http.NewServeMux()
	subMux.HandleFunc("POST /webhook", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var ev Event
		json.Unmarshal(body, &ev)
		log.Printf("subscriber: received webhook: %+v", ev)
		w.WriteHeader(http.StatusOK) // ack so the provider knows delivery succeeded
		received <- ev
	})
	subscriber := httptest.NewServer(subMux)
	defer subscriber.Close()

	// ── Registration ───────────────────────────────────────────────────────
	// The merchant tells the provider where to send events. This normally
	// happens once, out of band (a dashboard or an API call).
	prov.subscribe(subscriber.URL + "/webhook")

	// ── An event occurs ─────────────────────────────────────────────────────
	fmt.Println()
	log.Println("provider: a payment was just confirmed...")
	prov.emit(Event{Type: "payment.confirmed", Order: "order-42", Amount: 199.90})

	// Wait for the subscriber to confirm it handled the event.
	select {
	case ev := <-received:
		fmt.Printf("\nDone: the subscriber was pushed event %q for %s with no polling.\n", ev.Type, ev.Order)
	case <-time.After(2 * time.Second):
		log.Fatal("timed out waiting for webhook delivery")
	}
}
