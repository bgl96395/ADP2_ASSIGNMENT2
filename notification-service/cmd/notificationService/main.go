package main

import (
	"log"
	"notification-service/internal/jobqueue"
	"notification-service/internal/subscriber"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"context"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	redisURL := os.Getenv("REDIS_URL")
	var rdb *redis.Client
	opts, err := redis.ParseURL(redisURL)
	if err == nil {
		rdb = redis.NewClient(opts)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err := rdb.Ping(ctx).Err()
		if err != nil {
			log.Printf("WARN: Redis unavailable: %v. Idempotency will not work.", err)
			rdb = nil
		} else {
			log.Println("Connected to Redis")
		}
	} else {
		log.Printf("WARN: Invalid REDIS_URL: %v", err)
	}

	poolSize := 3
	val := os.Getenv("WORKER_POOL_SIZE")
	if val != "" {
		n, err := strconv.Atoi(val)
		if err == nil {
			poolSize = n
		}
	}

	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8080"
	}

	jq := jobqueue.NewJobQueue(rdb, gatewayURL, poolSize)

	natsURL := os.Getenv("NATS_URL")
	sub, err := subscriber.Connect(natsURL, jq)
	if err != nil {
		log.Fatalf("ERROR: failed to connect to NATS after retries: %v", err)
	}

	err = sub.Subscribe()
	if err != nil {
		log.Fatalf("ERROR: failed to subscribe: %v", err)
	}

	log.Println("Notification Service is running. Waiting for events...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Notification Service...")
	sub.Drain()
	log.Println("Notification Service stopped cleanly")
}
