package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"notification-service/internal/subscriber"
)

func main() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	sub, err := subscriber.Connect(natsURL)
	if err != nil {
		log.Fatalf("ERROR: failed to connect to NATS after retries: %v", err)
	}

	if err := sub.Subscribe(); err != nil {
		log.Fatalf("ERROR: failed to subscribe: %v", err)
	}

	log.Println("Notification Service is running. Waiting for events...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Notification Service...")
	sub.Drain()
	log.Println("Notification Service stopped cleanly")
	os.Exit(0)
}
