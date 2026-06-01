# System Design Studies — Go Samples

Small, focused Go programs I wrote while studying system design. Each one isolates **a single concept** so it can be read in a couple of minutes and run with one command.

The folder layout mirrors the [LinuxTips *Descomplicando o System Design*](https://github.com/) course repo (the `exemplos/` tree), so each sample sits next to the lesson it belongs to. As the course advances I add the missing topics in the same structure.

> Learning approach: many of these examples use a **barbecue** analogy — cooks, a grill, friends helping out — to make abstract concurrency ideas concrete. It comes from the course and I kept it because it sticks.

---

## Repository layout

```
exemplos/
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
│   └── algoritmos/         # request distribution algorithms
│       ├── round-robin/
│       ├── random/
│       ├── ip-hash/
│       ├── least-request/
│       ├── least-connection/   (bonus, not yet in course)
│       └── lor/                (bonus — Least Outstanding Requests)
└── protocolos-assincronos/
    └── kafka/              # parallel consumption via a Kafka consumer group
```

## Topics & matching course lessons

| Topic | Folder | Course article |
|---|---|---|
| Concurrency & Parallelism | `concurrency-parallelism/` | [concorrencia-paralelismo](https://fidelissauro.dev/concorrencia-paralelismo/) |
| Caching | `cache/` | [caching](https://fidelissauro.dev/caching/) |
| Load Balancing | `load-balancing/` | [load-balancing](https://fidelissauro.dev/load-balancing/) |
| Async Protocols / Messaging | `protocolos-assincronos/` | [mensageria-eventos-streaming](https://fidelissauro.dev/mensageria-eventos-streaming/) |

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
mise exec -- go run ./exemplos/load-balancing/algoritmos/round-robin
```

If mise's shell hook is set up, `cd`-ing into the repo activates the pinned toolchain and plain `go` commands just work. Don't have mise? Any [Go](https://go.dev/dl/) **1.22+** install runs the samples fine — `mise.toml` is only there to make the version reproducible.

## Running the samples

Every standalone example lives in its own directory as `package main` under a single module, so run any of them with:

```bash
# from the repo root
go run ./exemplos/concurrency-parallelism/concurrency
go run ./exemplos/load-balancing/algoritmos/round-robin
```

The `cache/cdn` sample starts an HTTP server:

```bash
go run ./exemplos/cache/cdn
# then, in another terminal:
curl localhost:8080/        # first hit  -> "Cache miss" (fetched from origin)
curl localhost:8080/        # second hit -> "Cache hit"  (served from disk)
```

The `protocolos-assincronos/kafka` sample is a self-contained module run with Docker (see its own README).

---

## What each sample teaches

### Concurrency & Parallelism (`exemplos/concurrency-parallelism/`)

| Sample | Idea in one line |
|---|---|
| `concurrency` | Start many tasks "at once" with goroutines; collect results through a channel, coordinate with a `WaitGroup`. |
| `parallelism` | Split a workload into one slice per CPU core so the work runs on different cores at the same instant. |
| `race_condition` | Unsynchronized `counter++` from many goroutines silently loses updates. Run with `go run -race .` to see it flagged. |
| `mutex` | A `sync.Mutex` serializes access to the shared counter — the race disappears. (Same program as `race_condition`, one lock added.) |
| `semaphore` | A buffered channel caps how many goroutines touch a resource at once (grill fits 3, 10 items queued). |
| `spinlock` | A lock that busy-waits on an atomic compare-and-swap instead of sleeping — useful only for very short waits. |

### Caching (`exemplos/cache/cdn/`)

A minimal CDN edge: a reverse proxy that hashes the request, serves a cached copy from disk when present (**cache hit**), and otherwise fetches from the origin, stores it, and returns it (**cache miss**). Demonstrates the core read-through caching pattern.

### Load Balancing (`exemplos/load-balancing/algoritmos/`)

Each algorithm is a self-contained simulation that prints how requests are distributed:

- **round-robin** — cycle through servers in order; perfectly even split.
- **random** — pick a server at random; converges to even over many requests.
- **ip-hash** — hash the client IP so the same client always hits the same server (sticky sessions).
- **least-request** — route to the server with the fewest *total* requests so far.
- **least-connection** *(bonus)* — route to the server with the fewest *active* connections right now.
- **lor** *(bonus)* — Least Outstanding Requests: route by in-flight request count; automatically steers traffic away from slow backends (Envoy's HTTP/2 default).

### Async Protocols — Kafka (`exemplos/protocolos-assincronos/kafka/`)

A producer publishes image-processing jobs to a 3-partition topic; consumers in a group split the partitions and process jobs in parallel. Scaling the consumer demonstrates how a Kafka consumer group provides parallelism across instances. Runs via `docker compose` — see the folder's own README.

---

## Notes

- **Bonus** items (`least-connection`, `lor`) are valid algorithms the course hasn't covered yet — they're here as extra practice and clearly marked.
- These are study samples, optimized for clarity over completeness: no production error handling, configuration, or graceful shutdown.

## Credits

Concepts, structure, and the barbecue teaching analogy come from the LinuxTips **Descomplicando o System Design** course by [@fidelissauro](https://fidelissauro.dev/). The code here is my own implementation written while following along.
