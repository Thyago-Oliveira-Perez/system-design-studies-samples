# Synchronous Communication Protocols

Different ways for a client and a server to talk. Each sample isolates one
protocol/pattern and is small enough to read in a couple of minutes.

A rough mental model:

| | Who starts the message | Connection | Shape of data |
|---|---|---|---|
| **REST** | client | one request/response | server-defined (per endpoint) |
| **Polling** | client (repeatedly) | many request/responses | server-defined |
| **Webhooks** | server (on event) | one request/response, reversed | server-defined |
| **RPC** | client | one call/return | method args/return |
| **gRPC** | client (also streaming) | HTTP/2, long-lived | typed protobuf messages |
| **WebSockets** | either side, anytime | one long-lived, full-duplex | anything (text/binary) |
| **GraphQL** | client | one request/response | **client picks the fields** |

Polling and webhooks are two answers to the same problem ("how does the client
learn about an event?"): pull vs. push.

## Running each sample

The first four use only the Go standard library and live in the repo's root
module — run them from the **repo root**:

```bash
# REST — start the server, then curl it (see the file header for more)
go run ./samples/synchronous-protocols/rest

# Polling — self-contained, prints the whole pull loop and exits
go run ./samples/synchronous-protocols/polling

# Webhooks — self-contained, shows register -> push -> ack and exits
go run ./samples/synchronous-protocols/webhooks

# RPC — two terminals
go run ./samples/synchronous-protocols/rpc/server
go run ./samples/synchronous-protocols/rpc/client
```

The last three need third-party libraries and are **nested modules** with their
own `go.mod` — `cd` into the folder first:

```bash
# gRPC (needs Go 1.25+)
cd samples/synchronous-protocols/grpc
go run ./server      # terminal 1
go run ./client      # terminal 2

# WebSockets — run several clients to see broadcast
cd samples/synchronous-protocols/websockets
go run ./server          # terminal 1
go run ./client alice    # terminal 2
go run ./client bob      # terminal 3

# GraphQL — serves + runs two demo queries, then stays up for your own curls
cd samples/synchronous-protocols/graphql
go run .
```

## gRPC code generation

The Go code under `grpc/gen/` is generated from `grpc/proto/bmi/bmi.proto` and
**committed** so the sample runs without any extra tooling. To regenerate after
editing the `.proto` you need [`buf`](https://buf.build/) and the two Go plugins:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/bufbuild/buf/cmd/buf@latest

cd samples/synchronous-protocols/grpc
buf generate
```

`buf` ships its own protobuf compiler, so a separate `protoc` install isn't
needed.
