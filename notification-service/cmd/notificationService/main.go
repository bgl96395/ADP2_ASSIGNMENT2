package main

import (
	"log"
	"notification-service/internal/subscriber"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	natsURL := os.Getenv("NATS_URL")

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
