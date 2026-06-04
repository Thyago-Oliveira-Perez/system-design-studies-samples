// Topic exchange — pattern-based routing with wildcards.
//
// Routing keys are dot-separated words ("invoice.priority.high"). Bindings can
// use two wildcards: a star matches exactly one word, a hash matches zero or
// more words:
//
//	invoice.priority.*   matches invoice.priority.high  (one word after)
//	invoice.#            matches invoice.priority.high AND invoice.refund.high
//
// This enables more dynamic routing than a direct exchange: one publish can
// reach several queues depending on how specific each binding is.
//
//	binding "invoice.priority.normal" → normal-only queue
//	binding "invoice.priority.high"   → high-only queue
//	binding "invoice.priority.*"      → data-lake queue (every priority)
//	binding "invoice.#"               → audit queue (everything under invoice)
//
// Self-contained: declares the exchange + queues, publishes a mix of priorities,
// then drains each queue and prints what it received.
//
// Run (needs the broker from docker-compose.yml up):
//
//	docker compose up -d            # from the rabbitmq/ folder
//	go run ./topic
package main

import (
	"context"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const exchange = "billing.topic"

func main() {
	conn, ch := connect()
	defer conn.Close()
	defer ch.Close()

	must(ch.ExchangeDeclare(exchange, "topic", false, true, false, false, nil),
		"declare exchange")

	normal := bindQueue(ch, "invoice.priority.normal") // exact
	high := bindQueue(ch, "invoice.priority.high")     // exact
	lake := bindQueue(ch, "invoice.priority.*")        // any priority
	audit := bindQueue(ch, "invoice.#")                // anything under invoice

	// Publish a mix. Each key fans out to every queue whose pattern matches.
	publish(ch, "invoice.priority.normal", "inv-1")
	publish(ch, "invoice.priority.normal", "inv-2")
	publish(ch, "invoice.priority.high", "inv-3")
	publish(ch, "invoice.refund.high", "inv-4") // only the audit "#" matches

	log.Printf("normal (invoice.priority.normal) received: %v", drain(ch, normal))
	log.Printf("high   (invoice.priority.high)   received: %v", drain(ch, high))
	log.Printf("lake   (invoice.priority.*)      received: %v", drain(ch, lake))
	log.Printf("audit  (invoice.#)               received: %v", drain(ch, audit))
}

func bindQueue(ch *amqp.Channel, pattern string) string {
	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	must(err, "declare queue")
	must(ch.QueueBind(q.Name, pattern, exchange, false, nil), "bind queue")
	return q.Name
}

func publish(ch *amqp.Channel, key, body string) {
	must(ch.PublishWithContext(context.Background(), exchange, key, false, false,
		amqp.Publishing{ContentType: "text/plain", Body: []byte(body)}),
		"publish")
	log.Printf("publish key=%-24s body=%s", key, body)
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
