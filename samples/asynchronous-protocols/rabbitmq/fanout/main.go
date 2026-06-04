// Fanout exchange — broadcast to every bound queue, routing key ignored.
//
// The simplest exchange: it copies each message to *all* queues bound to it,
// regardless of routing key. Use it when one event must reach several
// independent subscribers — e.g. a new sale that billing, stock, and logistics
// all need to react to, each with its own queue and its own pace.
//
//	                       ┌─▶ (billing queue)
//	publish "order-1" ──▶ [orders.fanout] ─▶ (stock queue)
//	                       └─▶ (logistics queue)
//
// Compare with samples/asynchronous-protocols/messaging-concepts/fanout, which
// models this same 1:N idea in plain Go with no broker.
//
// Run (needs the broker from docker-compose.yml up):
//
//	docker compose up -d            # from the rabbitmq/ folder
//	go run ./fanout
package main

import (
	"context"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const exchange = "orders.fanout"

func main() {
	conn, ch := connect()
	defer conn.Close()
	defer ch.Close()

	must(ch.ExchangeDeclare(exchange, "fanout", false, true, false, false, nil),
		"declare exchange")

	// Three independent subscribers. The binding key is ignored for fanout.
	billing := bindQueue(ch)
	stock := bindQueue(ch)
	logistics := bindQueue(ch)

	// One publish per order → every queue gets its own copy.
	publish(ch, "order-1")
	publish(ch, "order-2")

	log.Printf("billing   received: %v", drain(ch, billing))
	log.Printf("stock     received: %v", drain(ch, stock))
	log.Printf("logistics received: %v", drain(ch, logistics))
}

func bindQueue(ch *amqp.Channel) string {
	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	must(err, "declare queue")
	// Routing key "" — fanout ignores it and delivers to every bound queue.
	must(ch.QueueBind(q.Name, "", exchange, false, nil), "bind queue")
	return q.Name
}

func publish(ch *amqp.Channel, body string) {
	must(ch.PublishWithContext(context.Background(), exchange, "", false, false,
		amqp.Publishing{ContentType: "text/plain", Body: []byte(body)}),
		"publish")
	log.Printf("publish body=%s → all queues", body)
}

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
