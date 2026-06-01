// WebSocket client.
//
// It opens one connection, then runs two things at the same time over that
// single connection — proving the link is full-duplex:
//   - a reader goroutine that prints every message the server pushes
//     (chat broadcasts plus the server's per-second "tick");
//   - the main goroutine that sends a few messages with pauses in between.
//
// To keep the sample runnable with one command it sends canned messages and
// exits after a few seconds. Run several clients at once to watch a message
// from one appear on all of them.
//
//	go run ./server      # start this first
//	go run ./client
package main

import (
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	// Optional name from the command line, e.g. `go run ./client alice`.
	name := "client"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatalf("dial failed (is the server running?): %v", err)
	}
	defer conn.Close()

	// Reader goroutine: print whatever the server sends, whenever it sends it.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			log.Printf("recv: %s", data)
		}
	}()

	// Writer: send a few messages, then leave.
	for i := 1; i <= 3; i++ {
		msg := name + ": hello #" + string(rune('0'+i))
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Println("write:", err)
			break
		}
		time.Sleep(time.Second)
	}

	log.Println("done sending — closing connection")
	conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	<-done
}
