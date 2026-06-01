// WebSocket server: a tiny chat/broadcast hub.
//
// Unlike plain HTTP (one request -> one response), a WebSocket is a single
// long-lived, full-duplex connection: either side can send a message at any
// time. That's what makes real-time features (chat, live dashboards, game
// state) possible without polling.
//
// This hub:
//   - upgrades each /ws HTTP request into a WebSocket connection;
//   - broadcasts every message it receives to all connected clients;
//   - also pushes a server-initiated "tick" every second, to show that the
//     server can talk first — not just reply.
//
// Run the server, then one or more clients:
//
//	go run ./server
//	go run ./client       # in another terminal (run a few for broadcast)
package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// Accept connections from any origin — fine for a local study sample.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// hub keeps the set of live connections and fans messages out to all of them.
type hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]bool
}

func newHub() *hub { return &hub{clients: make(map[*websocket.Conn]bool)} }

func (h *hub) add(c *websocket.Conn) {
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()
}

func (h *hub) remove(c *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	c.Close()
}

// broadcast writes a text message to every connected client.
func (h *hub) broadcast(msg string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		if err := c.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			c.Close()
			delete(h.clients, c)
		}
	}
}

func main() {
	h := newHub()

	// Server push: announce a tick to everyone once per second.
	go func() {
		for t := range time.Tick(time.Second) {
			h.broadcast("server: tick " + t.Format("15:04:05"))
		}
	}()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade error:", err)
			return
		}
		h.add(conn)
		log.Printf("client connected (%s)", conn.RemoteAddr())
		h.broadcast("server: a client joined")

		// Read loop: every message from this client is broadcast to all.
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				log.Printf("client disconnected (%s)", conn.RemoteAddr())
				h.remove(conn)
				h.broadcast("server: a client left")
				return
			}
			h.broadcast(string(data))
		}
	})

	log.Println("WebSocket server listening on ws://localhost:8080/ws")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
