// consumer is a single member of the "image-processors" Kafka consumer group.
//
// External parallelism through consumer groups:
//   - The topic "image-jobs" has 3 partitions.
//   - Kafka's group coordinator assigns non-overlapping partitions to members:
//       1 consumer  → owns all 3 partitions → processes all jobs sequentially.
//       3 consumers → 1 partition each     → process jobs in parallel (3× throughput).
//
// Scale with: docker compose up --scale consumer=3
//
// Each container gets a different hostname — printed in every log line so you
// can see which consumer handles which partition's jobs.
package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type Job struct {
	ID       int    `json:"id"`
	Filename string `json:"filename"`
	Filter   string `json:"filter"`
}

func main() {
	broker := env("KAFKA_BROKER", "localhost:9092")
	group  := env("CONSUMER_GROUP", "image-processors")

	hostname, _ := os.Hostname()
	id := "consumer@" + hostname

	log.Printf("%s: joining group '%s'", id, group)

	processed := 0
	for {
		n, err := consume(id, broker, group)
		processed += n
		if err == nil {
			break
		}
		// __consumer_offsets is created lazily; the coordinator returns this
		// error for a brief window on a fresh cluster. Create a new reader and retry.
		if strings.Contains(err.Error(), "Group Coordinator Not Available") {
			log.Printf("%s: coordinator not ready — retrying in 3 s", id)
			time.Sleep(3 * time.Second)
			continue
		}
		log.Printf("%s: stopping (%v)", id, err)
		break
	}

	log.Printf("%s: done — processed %d jobs", id, processed)
}

// consume opens a reader, processes messages until an error, and returns the
// count of jobs processed along with the error (nil only if context cancelled).
func consume(id, broker, group string) (int, error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{broker},
		Topic:          "image-jobs",
		GroupID:        group,       // same group = Kafka coordinates partition assignment
		MinBytes:       1,
		MaxBytes:       1e6,
		CommitInterval: time.Second,
		// Read from the beginning when no committed offset exists (first run).
		// Without this, the default LastOffset makes the consumer miss messages
		// published before it joined (producer completes before consumer starts).
		StartOffset: kafka.FirstOffset,
	})
	defer r.Close()

	log.Printf("%s: waiting for partition assignment", id)

	processed := 0
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			return processed, err
		}

		var job Job
		if err := json.Unmarshal(m.Value, &job); err != nil {
			log.Printf("%s: unreadable message: %v", id, err)
			continue
		}

		processJob(id, job, m.Partition)
		processed++
	}
}

// processJob simulates CPU-bound image filtering.
// In a real system this would call an image library and write the result to S3.
func processJob(workerID string, job Job, partition int) {
	duration := time.Duration(200+rand.Intn(400)) * time.Millisecond

	log.Printf("%s [partition=%d] → start  job %3d  %s  filter=%s",
		workerID, partition, job.ID, job.Filename, job.Filter)

	time.Sleep(duration)

	log.Printf("%s [partition=%d] ✓ done   job %3d  in %s",
		workerID, partition, job.ID, duration)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
