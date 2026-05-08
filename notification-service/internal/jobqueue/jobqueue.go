package jobqueue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type Job struct {
	IdempotencyKey string
	AppointmentID  string
	DoctorID       string
	OccurredAt     string
	Channel        string
	Recipient      string
	Message        string
}

type JobQueue struct {
	jobs    chan Job
	rdb     *redis.Client
	gateway string
}

func NewJobQueue(rdb *redis.Client, gateway string, poolSize int) *JobQueue {
	jq := &JobQueue{
		jobs:    make(chan Job, 100),
		rdb:     rdb,
		gateway: gateway,
	}
	for i := 0; i < poolSize; i++ {
		go jq.worker()
	}
	return jq
}

func (jq *JobQueue) Enqueue(job Job) {
	ctx := context.Background()

	val, err := jq.rdb.Get(ctx, "idem:"+job.IdempotencyKey).Result()
	if err == nil && val == "done" {
		jq.logLine(os.Stdout, "info", job.IdempotencyKey, 0, "dropped_duplicate", "")
		return
	}

	jq.logLine(os.Stdout, "info", job.IdempotencyKey, 0, "enqueued", "")
	jq.jobs <- job
}

func (jq *JobQueue) worker() {
	for job := range jq.jobs {
		jq.process(job)
	}
}

func (jq *JobQueue) process(job Job) {
	backoffs := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	for attempt := 1; attempt <= 3; attempt++ {
		jq.logLine(os.Stdout, "info", job.IdempotencyKey, attempt, "processing", "")

		err := jq.callGateway(job)
		if err == nil {
			ctx := context.Background()
			jq.rdb.Set(ctx, "idem:"+job.IdempotencyKey, "done", 24*time.Hour)
			jq.logLine(os.Stdout, "info", job.IdempotencyKey, attempt, "success", "")
			return
		}

		jq.logLine(os.Stdout, "warn", job.IdempotencyKey, attempt, "retry", err.Error())

		if attempt < 3 {
			time.Sleep(backoffs[attempt-1])
		}
	}

	jq.logLine(os.Stderr, "error", job.IdempotencyKey, 3, "dead_letter", "max retries exceeded")
}

type notifyRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
	Channel        string `json:"channel"`
	Recipient      string `json:"recipient"`
	Message        string `json:"message"`
}

func (jq *JobQueue) callGateway(job Job) error {
	payload := notifyRequest{
		IdempotencyKey: job.IdempotencyKey,
		Channel:        job.Channel,
		Recipient:      job.Recipient,
		Message:        job.Message,
	}
	data, _ := json.Marshal(payload)

	resp, err := http.Post(jq.gateway+"/notify", "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusServiceUnavailable {
		return fmt.Errorf("gateway returned 503")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gateway returned %d", resp.StatusCode)
	}
	return nil
}

type logEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	JobID   string `json:"job_id"`
	Attempt int    `json:"attempt,omitempty"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

func (jq *JobQueue) logLine(out *os.File, level, jobID string, attempt int, status, errMsg string) {
	entry := logEntry{
		Time:    time.Now().UTC().Format(time.RFC3339),
		Level:   level,
		JobID:   jobID,
		Attempt: attempt,
		Status:  status,
		Error:   errMsg,
	}
	data, _ := json.Marshal(entry)
	fmt.Fprintln(out, string(data))
	_ = log.Writer()
}
