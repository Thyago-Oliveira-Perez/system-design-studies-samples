// gRPC client for the BMIService. See ../server for the service definition.
//
//	go run ./server      # start this first
//	go run ./client
package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"grpc-bmi/gen/bmi"
)

const addr = "localhost:50051"

func main() {
	// Dial the server. insecure credentials are fine for a local study sample;
	// real services use TLS.
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := bmi.NewBMIServiceClient(conn)

	// Calls carry a deadline — the request is cancelled if the server is slow.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req := &bmi.BMIRequest{WeightKg: 90.5, HeightM: 1.77}
	resp, err := client.Calculate(ctx, req)
	if err != nil {
		log.Fatalf("Calculate failed (is the server running?): %v", err)
	}

	log.Printf("weight %.1fkg, height %.2fm -> BMI %.2f (%s)",
		req.WeightKg, req.HeightM, resp.Bmi, resp.Category)
}
