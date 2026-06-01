// RPC server using Go's standard net/rpc package.
//
// RPC (Remote Procedure Call) makes calling a function on another process look
// almost like a local call: the client invokes Calculator.Add(a, b) and gets a
// result, while net/rpc handles connecting, serializing the arguments (gob by
// default), and shipping them over TCP.
//
// To expose methods over RPC they must follow net/rpc's shape:
//
//	func (t *T) Method(args ArgType, reply *ReplyType) error
//
// Run the server, then run the client in another terminal:
//
//	go run ./samples/synchronous-protocols/rpc/server
//	go run ./samples/synchronous-protocols/rpc/client
package main

import (
	"log"
	"net"
	"net/rpc"

	"system-design-studies-samples/samples/synchronous-protocols/rpc/shared"
)

// Calculator is the type whose methods we publish over RPC.
type Calculator struct{}

func (c *Calculator) Add(args shared.CalcArgs, reply *float64) error {
	*reply = args.A + args.B
	log.Printf("Add(%v, %v) = %v", args.A, args.B, *reply)
	return nil
}

func (c *Calculator) Multiply(args shared.CalcArgs, reply *float64) error {
	*reply = args.A * args.B
	log.Printf("Multiply(%v, %v) = %v", args.A, args.B, *reply)
	return nil
}

func main() {
	// Register makes Calculator's eligible methods callable as "Calculator.Add" etc.
	if err := rpc.Register(new(Calculator)); err != nil {
		log.Fatal(err)
	}

	ln, err := net.Listen("tcp", shared.Addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("RPC server listening on %s", shared.Addr)

	// One connection per client; serve each in its own goroutine.
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
