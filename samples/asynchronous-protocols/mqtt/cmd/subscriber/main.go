// subscriber consumes telemetry from "sensors/telemetry" using one of two
// subscription styles, selected by the SUBSCRIPTION env var. The contrast only
// shows up when you run several subscriber instances at once.
//
// default (the classic pub/sub):
//
//	Every subscriber that subscribes to the topic gets its OWN copy of every
//	message. Run 3 subscribers → each prints all readings. Use when all
//	consumers must see everything (e.g. every dashboard shows every sensor).
//
// shared (load balancing, MQTT v5 / Mosquitto's "$share/<group>/<topic>"):
//
//	Subscribers in the same share group split the messages between them — each
//	reading goes to exactly ONE member. Run 3 subscribers → readings are spread
//	≈1/3 each. Use to scale out processing of a high-volume topic.
//
// Scale with: docker compose up --build --scale subscriber=3
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const baseTopic = "sensors/telemetry"

func main() {
	broker := env("MQTT_BROKER", "tcp://localhost:1883")
	mode := env("SUBSCRIPTION", "default") // "default" or "shared"

	hostname, _ := os.Hostname()
	id := "subscriber@" + hostname

	// A shared subscription is just a specially-formed topic filter:
	//   $share/<group>/<topic>
	// The broker load-balances matching messages across the group's members.
	topic := baseTopic
	if mode == "shared" {
		topic = "$share/processors/" + baseTopic
	}

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(fmt.Sprintf("subscriber-%d", time.Now().UnixNano())).
		SetAutoReconnect(true).
		SetCleanSession(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("%s: connect: %v", id, token.Error())
	}

	handler := func(_ mqtt.Client, msg mqtt.Message) {
		log.Printf("%s: got [%s]", id, msg.Payload())
	}
	if token := client.Subscribe(topic, 1, handler); token.Wait() && token.Error() != nil {
		log.Fatalf("%s: subscribe: %v", id, token.Error())
	}
	log.Printf("%s: subscribed (%s) to %q", id, mode, topic)

	// Block until interrupted (Ctrl-C / docker stop).
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	log.Printf("%s: shutting down", id)
	client.Disconnect(250)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
