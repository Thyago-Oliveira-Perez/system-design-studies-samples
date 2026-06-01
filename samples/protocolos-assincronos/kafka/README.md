# Kafka — Parallel Consumption with a Consumer Group

A small end-to-end demo of **external parallelism** through Apache Kafka:

- A **producer** publishes 30 image-processing jobs to the topic `image-jobs`.
- The topic has **3 partitions**.
- One or more **consumers** join the group `image-processors`. Kafka assigns
  non-overlapping partitions to each member:
  - **1 consumer** → owns all 3 partitions → processes jobs sequentially.
  - **3 consumers** → 1 partition each → process jobs **in parallel** (≈3× throughput).

This is a self-contained Go module (it depends on
[`segmentio/kafka-go`](https://github.com/segmentio/kafka-go)), separate from
the repo's root module.

## Run it

Requires Docker. From this folder:

```bash
# 1 consumer — jobs processed one partition at a time
docker compose up --build

# 3 consumers — partitions split across instances, processed in parallel
docker compose up --build --scale consumer=3
```

What happens on `up`:

1. `kafka` broker starts (KRaft mode, no ZooKeeper).
2. `init-kafka` pre-creates the `image-jobs` topic with exactly 3 partitions.
3. `producer` publishes 30 jobs (keyed so related jobs land on the same
   partition) then exits cleanly.
4. `consumer` instances join the group and process jobs, logging which
   partition each job came from.

A Kafka UI is available at **http://localhost:8080** while the stack is up.

## Run the Go binaries directly (without Docker)

If you already have a broker on `localhost:9092`:

```bash
go run ./cmd/producer     # publishes jobs (honors JOB_COUNT, default 30)
go run ./cmd/consumer     # joins the group and processes jobs
```

### Environment variables

| Variable | Default | Used by |
|---|---|---|
| `KAFKA_BROKER` | `localhost:9092` | producer, consumer |
| `JOB_COUNT` | `30` | producer |
| `CONSUMER_GROUP` | `image-processors` | consumer |
