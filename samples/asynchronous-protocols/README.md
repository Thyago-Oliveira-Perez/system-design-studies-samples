# Asynchronous Communication

Producers and consumers that don't talk at the same instant. Work is handed off
through an intermediary (a queue, a topic, a broker) so the two sides are
decoupled in time — the producer doesn't wait for the consumer.

This topic has two layers:

- **Messaging concepts** (`messaging-concepts/`) — the building blocks, each a
  tiny standard-library program. A queue is a *data structure* before it is ever
  a technology like RabbitMQ or SQS, so these come first.
- **Real brokers** — end-to-end demos against an actual broker, each its own
  nested module run with Docker:
  - `kafka/` — parallel consumption via a Kafka consumer group.
  - `rabbitmq/` — AMQP exchange routing (direct / topic / fanout).
  - `mqtt/` — lightweight IoT pub/sub; default vs shared subscriptions.

## Messaging concepts

| Sample | Idea in one line |
|---|---|
| `queue` | **FIFO** — items are consumed in the order produced (enqueue/dequeue). |
| `stack` | **LIFO** — the newest item is served first (push/pop); the opposite ordering. |
| `fanout` | **1:N** — one production is copied to *every* bound queue; consumers don't compete. |
| `dead-letter-queue` | After N failed retries a "poison" message is moved aside so the main queue keeps flowing. |
| `batch-processing` | Accumulate messages and flush the batch on a **size** or **time** trigger (throughput over latency). |

A useful contrast: a **message** is imperative ("do something", one producer →
one consumer), while an **event** is reactive ("something happened", one
producer → N consumers). `queue`/`stack`/`dead-letter-queue` model the message
side; `fanout` models the event side.

## Running the concept samples

All five use only the Go standard library and live in the repo's root module —
run them from the **repo root**:

```bash
go run ./samples/asynchronous-protocols/messaging-concepts/queue
go run ./samples/asynchronous-protocols/messaging-concepts/stack
go run ./samples/asynchronous-protocols/messaging-concepts/fanout
go run ./samples/asynchronous-protocols/messaging-concepts/dead-letter-queue
go run ./samples/asynchronous-protocols/messaging-concepts/batch-processing
```

Each is self-contained: it prints what it's doing and exits.

## Broker samples (nested modules, need Docker)

Each lives in its own folder with a `docker-compose.yml` and its own README:

| Folder | Protocol | What it shows |
|---|---|---|
| [`kafka/`](kafka/README.md) | Kafka | a consumer group splitting partitions for parallel consumption |
| [`rabbitmq/`](rabbitmq/README.md) | AMQP | direct / topic / fanout exchange routing |
| [`mqtt/`](mqtt/README.md) | MQTT | IoT pub/sub; default (everyone) vs shared (load-balanced) subscriptions |
