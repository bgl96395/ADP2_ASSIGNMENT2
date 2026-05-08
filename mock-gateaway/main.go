package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

type NotifyRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
	Channel        string `json:"channel"`
	Recipient      string `json:"recipient"`
	Message        string `json:"message"`
}

type NotifyResponse struct {
	Status string `json:"status"`
}

var (
	processed = make(map[string]bool)
	mu        sync.Mutex
)

func main() {
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/notify", handleNotify)

	log.Printf("Mock Notification Gateway running on port :%s", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Failed to start gateway: %v", err)
	}
}

func handleNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req NotifyRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	logEntry := map[string]any{
		"time":            time.Now().UTC().Format(time.RFC3339),
		"idempotency_key": req.IdempotencyKey,
		"channel":         req.Channel,
		"recipient":       req.Recipient,
		"message":         req.Message,
	}
	logBytes, _ := json.Marshal(logEntry)
	log.Println(string(logBytes))

	if rand.Float32() < 0.20 {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	mu.Lock()
	defer mu.Unlock()

	if processed[req.IdempotencyKey] {
		json.NewEncoder(w).Encode(NotifyResponse{Status: "duplicate"})
		return
	}

	processed[req.IdempotencyKey] = true
	json.NewEncoder(w).Encode(NotifyResponse{Status: "accepted"})
}
