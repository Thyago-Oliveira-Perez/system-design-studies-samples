# System Design Studies — Go Samples

Small, focused Go programs I wrote while studying system design. Each one isolates **a single concept** so it can be read in a couple of minutes and run with one command.

The folder layout (my `samples/` tree) mirrors the [LinuxTips *Descomplicando o System Design*](https://github.com/) course repo (its `exemplos/` tree), so each sample sits next to the lesson it belongs to. As the course advances I add the missing topics in the same structure.

> Learning approach: many of these examples use a **barbecue** analogy — cooks, a grill, friends helping out — to make abstract concurrency ideas concrete. It comes from the course and I kept it because it sticks.

---

## Repository layout

```
samples/
├── concurrency-parallelism/
│   ├── concurrency/        # goroutines + channel + WaitGroup
│   ├── parallelism/        # work split across CPU cores
│   ├── race_condition/     # the bug: lost updates on a shared counter
│   ├── mutex/              # the fix: mutual exclusion
│   ├── semaphore/          # limit N concurrent users of a resource
│   └── spinlock/           # busy-wait lock built on atomics
├── cache/
│   └── cdn/                # a tiny caching reverse proxy (CDN edge)
├── load-balancing/
│   └── algorithms/         # request distribution algorithms
│       ├── round-robin/
│       ├── random/
│       ├── ip-hash/
│       ├── least-request/
│       ├── least-connection/   (bonus, not yet in course)
│       └── lor/                (bonus — Least Outstanding Requests)
├── synchronous-protocols/      # synchronous communication protocols
│   ├── rest/               # resource-oriented HTTP API (CRUD)
│   ├── polling/            # client repeatedly asks "done yet?"
│   ├── webhooks/           # server pushes events to a callback URL
│   ├── rpc/                # Go net/rpc: call a remote func like a local one
│   ├── grpc/               # gRPC + Protocol Buffers (nested module)
│   ├── websockets/         # full-duplex real-time connection (nested module)
│   └── graphql/            # client picks exactly the fields it needs (nested module)
└── asynchronous-protocols/
    ├── messaging-concepts/     # building blocks (stdlib, one concept each)
    │   ├── queue/              # FIFO — first in, first out
    │   ├── stack/              # LIFO — last in, first out
    │   ├── fanout/             # 1:N — copy one message to every queue
    │   ├── dead-letter-queue/  # park poison messages after N retries
    │   └── batch-processing/   # flush on a size or time trigger
    ├── kafka/                  # parallel consumption via a Kafka consumer group (nested module)
    ├── rabbitmq/               # AMQP exchange routing: direct / topic / fanout (nested module)
    └── mqtt/                   # IoT pub/sub: default vs shared subscriptions   (nested module)
```

## Topics & matching course lessons

| Topic | Folder | Course article |
|---|---|---|
| Concurrency & Parallelism | `concurrency-parallelism/` | [concorrencia-paralelismo](https://fidelissauro.dev/concorrencia-paralelismo/) |
| Caching | `cache/` | [caching](https://fidelissauro.dev/caching/) |
| Load Balancing | `load-balancing/` | [load-balancing](https://fidelissauro.dev/load-balancing/) |
| Synchronous Protocols | `synchronous-protocols/` | [padroes-de-comunicacao-sincronos](https://fidelissauro.dev/padroes-de-comunicacao-sincronos/) |
| Async Protocols / Messaging | `asynchronous-protocols/` | [mensageria-eventos-streaming](https://fidelissauro.dev/mensageria-eventos-streaming/) |

---

## Toolchain (mise)

The Go version for this repo is pinned with [**mise**](https://mise.jdx.dev/), the version manager I use across these projects. The pin lives in [`mise.toml`](mise.toml):

```toml
[tools]
go = "1.26.2"
```

With mise installed, run this once in the repo and you get the exact Go version automatically:

```bash
mise install      # downloads Go 1.26.2 if you don't have it
mise exec -- go run ./samples/load-balancing/algorithms/round-robin
```

If mise's shell hook is set up, `cd`-ing into the repo activates the pinned toolchain and plain `go` commands just work. Don't have mise? Any [Go](https://go.dev/dl/) **1.22+** install runs the samples fine — `mise.toml` is only there to make the version reproducible.

## Running the samples

Most examples are `package main` under the repo's single root module, so run any of them with:

```bash
# from the repo root
go run ./samples/concurrency-parallelism/concurrency
go run ./samples/load-balancing/algorithms/round-robin
```

The `cache/cdn` sample starts an HTTP server:

```bash
go run ./samples/cache/cdn
# then, in another terminal:
curl localhost:8080/        # first hit  -> "Cache miss" (fetched from origin)
curl localhost:8080/        # second hit -> "Cache hit"  (served from disk)
```

A few samples that need third-party libraries are **nested modules** with their own `go.mod` (`synchronous-protocols/grpc`, `synchronous-protocols/websockets`, `synchronous-protocols/graphql`, and the broker demos `asynchronous-protocols/kafka`, `asynchronous-protocols/rabbitmq`, `asynchronous-protocols/mqtt`). For those, `cd` into the folder first and run the commands in its README. The gRPC sample needs **Go 1.25+** (a transitive gRPC dependency requires it) and the MQTT sample needs **Go 1.24+**; everything else runs on Go 1.22+.

---

## What each sample teaches

### Concurrency & Parallelism (`samples/concurrency-parallelism/`)

| Sample | Idea in one line |
|---|---|
| `concurrency` | Start many tasks "at once" with goroutines; collect results through a channel, coordinate with a `WaitGroup`. |
| `parallelism` | Split a workload into one slice per CPU core so the work runs on different cores at the same instant. |
| `race_condition` | Unsynchronized `counter++` from many goroutines silently loses updates. Run with `go run -race .` to see it flagged. |
| `mutex` | A `sync.Mutex` serializes access to the shared counter — the race disappears. (Same program as `race_condition`, one lock added.) |
| `semaphore` | A buffered channel caps how many goroutines touch a resource at once (grill fits 3, 10 items queued). |
| `spinlock` | A lock that busy-waits on an atomic compare-and-swap instead of sleeping — useful only for very short waits. |

### Caching (`samples/cache/cdn/`)

A minimal CDN edge: a reverse proxy that hashes the request, serves a cached copy from disk when present (**cache hit**), and otherwise fetches from the origin, stores it, and returns it (**cache miss**). Demonstrates the core read-through caching pattern.

### Load Balancing (`samples/load-balancing/algorithms/`)

Each algorithm is a self-contained simulation that prints how requests are distributed:

- **round-robin** — cycle through servers in order; perfectly even split.
- **random** — pick a server at random; converges to even over many requests.
- **ip-hash** — hash the client IP so the same client always hits the same server (sticky sessions).
- **least-request** — route to the server with the fewest *total* requests so far.
- **least-connection** *(bonus)* — route to the server with the fewest *active* connections right now.
- **lor** *(bonus)* — Least Outstanding Requests: route by in-flight request count; automatically steers traffic away from slow backends (Envoy's HTTP/2 default).

### Synchronous Protocols (`samples/synchronous-protocols/`)

Ways for a client to talk to a server and (usually) wait for the answer. See the [folder README](samples/synchronous-protocols/README.md) for run instructions.

| Sample | Idea in one line |
|---|---|
| `rest` | Resource-oriented HTTP: the URL names the resource, the verb (GET/POST/PUT/DELETE) is the action. Pure stdlib, in-memory CRUD. |
| `polling` | No server push, so the client asks "done yet?" on a fixed interval — simple, but most requests are wasted. |
| `webhooks` | The inverse of polling: the client registers a URL and the server **POSTs** events to it when they happen. |
| `rpc` | Go's `net/rpc`: call `Calculator.Add` on a remote process almost like a local function. |
| `grpc` | RPC over HTTP/2 with Protocol Buffers — a typed `.proto` contract generates both server and client. *(nested module)* |
| `websockets` | One long-lived, full-duplex connection; either side can send anytime (chat + server push). *(nested module)* |
| `graphql` | One endpoint where the **client** picks exactly which fields come back — no over/under-fetching. *(nested module)* |

### Async Protocols (`samples/asynchronous-protocols/`)

Producers and consumers decoupled in time, handing work off through an intermediary. See the [folder README](samples/asynchronous-protocols/README.md) for run instructions.

**Messaging concepts** (`messaging-concepts/`) — the building blocks, each a tiny stdlib program. A queue is a *data structure* before it is ever a technology:

| Sample | Idea in one line |
|---|---|
| `queue` | **FIFO** — items are consumed in the order they were produced (enqueue/dequeue). |
| `stack` | **LIFO** — the newest item is served first (push/pop); the opposite ordering. |
| `fanout` | **1:N** — one production is copied to *every* bound queue; consumers don't compete for it. |
| `dead-letter-queue` | After N failed retries a "poison" message is moved aside so the main queue keeps flowing. |
| `batch-processing` | Accumulate messages, then flush the batch on a **size** or **time** trigger (throughput over latency). |

**Brokers** — end-to-end demos against a real broker, each a nested module run with `docker compose` (see each folder's README):

- **Kafka** (`kafka/`) — a producer publishes image-processing jobs to a 3-partition topic; consumers in a group split the partitions and process jobs in parallel. Scaling the consumer shows how a consumer group provides parallelism across instances.
- **RabbitMQ** (`rabbitmq/`) — three self-contained demos of AMQP exchange routing: `direct` (exact routing key), `topic` (wildcard patterns), and `fanout` (broadcast to every bound queue).
- **MQTT** (`mqtt/`) — lightweight IoT pub/sub. Scaling the subscriber shows the difference between a **default** subscription (every instance gets every message) and a **shared** subscription (the broker load-balances messages across the group). *(needs Go 1.24+)*

---

## Notes

- **Bonus** items (`least-connection`, `lor`) are valid algorithms the course hasn't covered yet — they're here as extra practice and clearly marked.
- These are study samples, optimized for clarity over completeness: no production error handling, configuration, or graceful shutdown.

## Credits

Concepts, structure, and the barbecue teaching analogy come from the LinuxTips **Descomplicando o System Design** course by [@fidelissauro](https://fidelissauro.dev/). The code here is my own implementation written while following along.
