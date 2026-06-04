# MQTT — Telemetry, Default vs Shared Subscriptions

**MQTT** (Message Queuing Telemetry Transport) is a lightweight publish/subscribe
protocol over TCP, built for IoT and edge devices with limited CPU and bandwidth.
A publisher sends to a **topic** on a broker without knowing who is subscribed;
subscribers receive messages from topics they subscribe to.

This demo has a publisher emitting one telemetry reading per second, and a
subscriber you **scale to several instances** to compare the two subscription
styles:

| Mode | Topic filter | Behaviour with 3 subscribers |
|---|---|---|
| **default** | `sensors/telemetry` | each instance gets **its own copy** of every reading |
| **shared** | `$share/processors/sensors/telemetry` | the broker **load-balances** readings across the group — each goes to exactly one instance |

Default is classic pub/sub (every subscriber sees everything — e.g. dashboards).
Shared subscription is closer to a load balancer / work queue, for scaling out
processing of a high-volume topic.

This is a nested Go module (depends on
[`eclipse/paho.mqtt.golang`](https://github.com/eclipse/paho.mqtt.golang)),
separate from the repo's root module. It builds with **Go 1.24+** (the MQTT
client requires it).

## Run it

Requires Docker. From this folder:

```bash
# default subscription — each of the 3 subscribers prints EVERY reading
docker compose up --build --scale subscriber=3

# shared subscription — the 3 subscribers SPLIT the readings (~1/3 each)
SUBSCRIPTION=shared docker compose up --build --scale subscriber=3
```

Watch the logs: in `default` mode every `subscriber@<host>` line shows the same
readings; in `shared` mode each reading appears under only one subscriber.

Stop with `Ctrl-C`, then `docker compose down`.

## Run the Go binaries directly (without Docker)

If you already have a broker on `localhost:1883` (e.g. `eclipse-mosquitto`):

```bash
go run ./cmd/subscriber                 # default mode
SUBSCRIPTION=shared go run ./cmd/subscriber
go run ./cmd/publisher                  # in another terminal
```

### Environment variables

| Variable | Default | Used by |
|---|---|---|
| `MQTT_BROKER` | `tcp://localhost:1883` | publisher, subscriber |
| `SUBSCRIPTION` | `default` | subscriber (`default` or `shared`) |

## Notes

- **QoS 1** (at least once) is used so the broker acknowledges delivery; it may
  duplicate a message but won't silently drop it.
- Shared subscriptions are an MQTT v5 feature; Mosquitto also honours the
  `$share/<group>/<topic>` form for the v3.1.1 client used here.
