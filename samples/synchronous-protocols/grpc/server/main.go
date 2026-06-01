// gRPC server for the BMIService defined in proto/bmi/bmi.proto.
//
// gRPC = RPC over HTTP/2 with Protocol Buffers as the message format. Compared
// to the plain net/rpc sample (../../rpc) it adds: a language-neutral schema
// (the .proto), compact binary encoding, streaming, deadlines, and generated
// clients for many languages.
//
// The generated code (gen/bmi) gives us the BMIServiceServer interface to
// implement and RegisterBMIServiceServer to wire it up. We only write the
// business logic.
//
// Regenerate the gen/ code after editing the .proto with:  buf generate
//
// Run the server, then the client in another terminal:
//
//	go run ./server
//	go run ./client
package main

import (
	"context"
	"log"
	"math"
	"net"

	"google.golang.org/grpc"

	"grpc-bmi/gen/bmi"
)

const addr = "localhost:50051"

// server implements the generated bmi.BMIServiceServer interface.
// Embedding UnimplementedBMIServiceServer keeps us forward-compatible if new
// methods are added to the service later.
type server struct {
	bmi.UnimplementedBMIServiceServer
}

func (s *server) Calculate(ctx context.Context, req *bmi.BMIRequest) (*bmi.BMIResponse, error) {
	value := req.WeightKg / (req.HeightM * req.HeightM)
	value = math.Round(value*100) / 100

	log.Printf("Calculate(weight=%.1fkg, height=%.2fm) -> BMI %.2f", req.WeightKg, req.HeightM, value)
	return &bmi.BMIResponse{Bmi: value, Category: categorize(value)}, nil
}

func categorize(v float64) string {
	switch {
	case v < 18.5:
		return "underweight"
	case v < 25:
		return "normal"
	case v < 30:
		return "overweight"
	default:
		return "obese"
	}
}

func main() {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	bmi.RegisterBMIServiceServer(s, &server{})

	log.Printf("gRPC server listening on %s", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
