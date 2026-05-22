package main

import (
	"log"
	"net"

	"github.com/trigold786/94-AI-Insurance-Design/shared/config"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	// Register ActuarialEngine service — implemented in Sprint 5-6

	log.Printf("actuarial-engine starting on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
