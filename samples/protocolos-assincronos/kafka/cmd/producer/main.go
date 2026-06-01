// producer publishes N image-processing jobs to the "image-jobs" Kafka topic.
//
// Each job carries a filename and a filter operation (grayscale, blur, …).
// The topic has 3 partitions; jobs are routed by key so that related jobs land
// on the same partition — while unrelated jobs spread across all three,
// enabling parallel consumption.
//
// After all jobs are published the process exits cleanly (docker-compose marks
// it service_completed_successfully, which releases the consumer dependency).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
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
	count  := envInt("JOB_COUNT", 30)

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{broker},
		Topic:        "image-jobs",
		Balancer:     &kafka.Hash{}, // same key → same partition (deterministic)
		WriteTimeout: 10 * time.Second,
	})
	defer w.Close()

	log.Printf("producer: publishing %d jobs → topic 'image-jobs' on %s", count, broker)
	log.Printf("producer: topic has 3 partitions → up to 3 consumers work in parallel")

	filters := []string{"grayscale", "blur", "sharpen", "sepia", "contrast"}

	for i := range count {
		job := Job{
			ID:       i + 1,
			Filename: fmt.Sprintf("image_%03d.jpg", i+1),
			Filter:   filters[i%len(filters)],
		}
		payload, _ := json.Marshal(job)

		// Key controls which partition this message lands on.
		// Hash(key) % numPartitions = partition index.
		key := fmt.Sprintf("job-%d", job.ID)

		if err := w.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(key),
			Value: payload,
		}); err != nil {
			log.Printf("producer: failed to write job %d: %v", job.ID, err)
			continue
		}

		log.Printf("producer: published job %3d  %-25s  filter=%-10s",
			job.ID, job.Filename, job.Filter)

		time.Sleep(50 * time.Millisecond) // throttle so consumers visibly race
	}

	log.Printf("producer: all %d jobs published", count)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
