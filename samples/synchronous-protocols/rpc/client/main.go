// RPC client for the Calculator service. See ../server for the contract.
//
//	go run ./samples/synchronous-protocols/rpc/server   # start this first
//	go run ./samples/synchronous-protocols/rpc/client
package main

import (
	"log"

	"system-design-studies-samples/samples/synchronous-protocols/rpc/shared"

	"net/rpc"
)

func main() {
	client, err := rpc.Dial("tcp", shared.Addr)
	if err != nil {
		log.Fatalf("failed to connect (is the server running?): %v", err)
	}
	defer client.Close()

	// A remote call looks almost like a local one: name the method, pass the
	// args, and point at a variable to receive the reply.
	args := shared.CalcArgs{A: 7, B: 6}

	var sum float64
	if err := client.Call(shared.ServiceName+".Add", args, &sum); err != nil {
		log.Fatalf("Add call failed: %v", err)
	}
	log.Printf("remote Add(%v, %v) = %v", args.A, args.B, sum)

	var product float64
	if err := client.Call(shared.ServiceName+".Multiply", args, &product); err != nil {
		log.Fatalf("Multiply call failed: %v", err)
	}
	log.Printf("remote Multiply(%v, %v) = %v", args.A, args.B, product)
}
