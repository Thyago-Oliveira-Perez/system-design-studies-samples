// Direct exchange — point-to-point routing by an *exact* routing key.
//
// The default and most common exchange type. A message is delivered to the
// queues whose binding key exactly equals the message's routing key. Used to
// hand specific commands to specific workers ("charge", "ship", "cancel"...).
//
//	                          binding key "charge"
//	publish key "charge" ──▶ [orders.direct] ──▶ (billing queue)
//	publish key "ship"   ──▶              └────▶ (shipping queue)  binding "ship"
//	publish key "refund" ──▶  no matching binding → message dropped
//
// Self-contained: declares the exchange + queues, publishes a few messages,
// then drains each queue and prints what it received.
//
// Run (needs the broker from docker-compose.yml up):
//
//	docker compose up -d            # from the rabbitmq/ folder
//	go run ./direct
package main

import (
	"context"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const exchange = "orders.direct"

func main() {
	conn, ch := connect()
	defer conn.Close()
	defer ch.Close()

	// A direct exchange routes by exact routing-key match.
	must(ch.ExchangeDeclare(exchange, "direct", false, true, false, false, nil),
		"declare exchange")

	// Two queues, each bound to one exact routing key.
	billing := bindQueue(ch, "charge")
	shipping := bindQueue(ch, "ship")

	// Publish: two "charge" commands, one "ship", and one "refund" that no
	// queue is bound to (it will simply be dropped by the exchange).
	publish(ch, "charge", "order-1")
	publish(ch, "charge", "order-2")
	publish(ch, "ship", "order-1")
	publish(ch, "refund", "order-3") // no binding → goes nowhere

	// Drain each queue and report. "refund" appears in neither.
	log.Printf("billing  (key=charge) received: %v", drain(ch, billing))
	log.Printf("shipping (key=ship)   received: %v", drain(ch, shipping))
}

// bindQueue declares a temporary queue and binds it to the exchange with an
// exact routing key. Returns the consumer delivery channel name (queue name).
func bindQueue(ch *amqp.Channel, key string) string {
	q, err := ch.QueueDeclare("", false, true, true, false, nil) // server-named, auto-delete
	must(err, "declare queue")
	must(ch.QueueBind(q.Name, key, exchange, false, nil), "bind queue")
	return q.Name
}

func publish(ch *amqp.Channel, key, body string) {
	must(ch.PublishWithContext(context.Background(), exchange, key, false, false,
		amqp.Publishing{ContentType: "text/plain", Body: []byte(body)}),
		"publish")
	log.Printf("publish  key=%-7s body=%s", key, body)
}

// drain reads every message currently queued, stopping after a short idle gap.
func drain(ch *amqp.Channel, queue string) []string {
	deliveries, err := ch.Consume(queue, "", true, false, false, false, nil)
	must(err, "consume")
	var got []string
	for {
		select {
		case d := <-deliveries:
			got = append(got, string(d.Body))
		case <-time.After(300 * time.Millisecond):
			return got
		}
	}
}

func connect() (*amqp.Connection, *amqp.Channel) {
	url := env("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	conn, err := amqp.Dial(url)
	must(err, "dial broker")
	ch, err := conn.Channel()
	must(err, "open channel")
	return conn, ch
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func must(err error, what string) {
	if err != nil {
		log.Fatalf("%s: %v", what, err)
	}
}
