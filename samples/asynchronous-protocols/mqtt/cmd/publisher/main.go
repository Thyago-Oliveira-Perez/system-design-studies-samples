// publisher emits IoT telemetry readings to the MQTT topic "sensors/telemetry".
//
// MQTT (Message Queuing Telemetry Transport) is a lightweight publish/subscribe
// protocol over TCP, designed for IoT and edge devices with limited CPU and
// bandwidth. A publisher never knows who (if anyone) is subscribed — it just
// sends to a topic on the broker.
//
// This publisher loops forever, sending one reading per second so you can watch
// how subscribers receive them (see cmd/subscriber for default vs shared
// subscriptions).
package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const topic = "sensors/telemetry"

func main() {
	broker := env("MQTT_BROKER", "tcp://localhost:1883")
	clientID := fmt.Sprintf("publisher-%d", time.Now().UnixNano())

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetAutoReconnect(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("publisher: connect: %v", token.Error())
	}
	log.Printf("publisher: connected to %s, publishing to %q", broker, topic)

	// Give subscribers a moment to come up and subscribe.
	time.Sleep(3 * time.Second)

	for reading := 1; ; reading++ {
		payload := fmt.Sprintf("reading #%d temp=%.1f°C", reading, 18+rand.Float64()*10)

		// QoS 1 = at least once: the broker acknowledges delivery (may duplicate).
		token := client.Publish(topic, 1, false, payload)
		token.Wait()
		if err := token.Error(); err != nil {
			log.Printf("publisher: publish error: %v", err)
		} else {
			log.Printf("publisher: sent %s", payload)
		}
		time.Sleep(1 * time.Second)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
