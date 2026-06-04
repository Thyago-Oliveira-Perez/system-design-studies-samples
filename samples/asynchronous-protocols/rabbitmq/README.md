# RabbitMQ — AMQP Exchanges

Three small, self-contained demos of how an **AMQP** broker (RabbitMQ) routes
messages from a producer to queues through an **exchange**. The exchange *type*
decides the routing rule:

| Demo | Exchange type | Routing rule |
|---|---|---|
| `direct` | direct | deliver to queues whose binding key **exactly** equals the routing key |
| `topic` | topic | match routing key against binding **patterns** (`*` = one word, `#` = zero+ words) |
| `fanout` | fanout | ignore the routing key — **broadcast** to every bound queue |

The AMQP vocabulary: a **producer** publishes to an **exchange** over a
**channel**; the exchange uses **binding keys** to route into **queues**;
**consumers** read from queues.

Each demo is one file that does the whole round trip in a single process —
declare the exchange, bind a few queues, publish a handful of messages, then
drain each queue and print what landed where. Read it top to bottom in a couple
of minutes.

This is a nested Go module (depends on
[`rabbitmq/amqp091-go`](https://github.com/rabbitmq/amqp091-go), the maintained
successor to `streadway/amqp`), separate from the repo's root module.

## Run it

Requires Docker. From this folder, start the broker, then run a demo:

```bash
docker compose up -d        # RabbitMQ broker (+ management UI on :15672)

go run ./direct
go run ./topic
go run ./fanout

docker compose down         # when finished
```

The management UI is at **http://localhost:15672** (login `guest` / `guest`) —
handy for watching exchanges, queues, and bindings appear while a demo runs.

### What you should see

- **direct** — `billing` receives only the `charge` messages, `shipping` only
  the `ship` message; the `refund` publish matches no binding and is dropped.
- **topic** — the `invoice.priority.*` data-lake queue gets every priority, the
  `invoice.#` audit queue gets those *and* the `invoice.refund.high` message,
  while the exact-match queues each get only their own priority.
- **fanout** — all three queues (`billing`, `stock`, `logistics`) receive every
  order, because fanout ignores the routing key.

### Environment variables

| Variable | Default | Used by |
|---|---|---|
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | all demos |

## See also

`samples/asynchronous-protocols/messaging-concepts/fanout` models the same 1:N
fanout idea in plain Go, with no broker.
