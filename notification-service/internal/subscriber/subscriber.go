package subscriber

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"notification-service/internal/jobqueue"
	"time"

	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	conn          *nats.Conn
	subscriptions []*nats.Subscription
	jobQueue      *jobqueue.JobQueue
}

func Connect(natsURL string, jq *jobqueue.JobQueue) (*Subscriber, error) {
	var conn *nats.Conn
	var err error

	delays := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second}
	for counter, delay := range delays {
		conn, err = nats.Connect(natsURL)
		if err == nil {
			break
		}
		fmt.Printf("WARN: attempt %d failed to connect to NATS: %v. Retrying in %v...\n", counter+1, err, delay)
		time.Sleep(delay)
	}
	if err != nil {
		return nil, err
	}

	fmt.Println("Notification Service connected to NATS")
	return &Subscriber{conn: conn, jobQueue: jq}, nil
}

func (subscriber *Subscriber) Subscribe() error {
	subjects := []string{"doctors.created", "appointments.created", "appointments.status_updated"}

	for _, subject := range subjects {
		subj := subject
		sub, err := subscriber.conn.Subscribe(subj, func(msg *nats.Msg) {
			subscriber.handleMessage(subj, msg.Data)
		})
		if err != nil {
			return err
		}
		subscriber.subscriptions = append(subscriber.subscriptions, sub)
		fmt.Printf("Subscribed to subject: %s\n", subj)
	}
	return nil
}

func (subscriber *Subscriber) handleMessage(subject string, data []byte) {
	var payload map[string]any
	err := json.Unmarshal(data, &payload)
	if err != nil {
		fmt.Printf("ERROR: failed to deserialize message on %s: %v\n", subject, err)
		return
	}

	logEntry := map[string]any{
		"time":    time.Now().UTC().Format(time.RFC3339),
		"subject": subject,
		"event":   payload,
	}
	logBytes, _ := json.Marshal(logEntry)
	fmt.Println(string(logBytes))

	if subject == "appointments.status_updated" {
		newStatus, _ := payload["new_status"].(string)
		if newStatus == "done" {
			subscriber.enqueueJob(payload)
		}
	}
}

func (subscriber *Subscriber) enqueueJob(payload map[string]any) {
	id, _ := payload["id"].(string)
	doctorID, _ := payload["doctor_id"].(string)
	occurredAt, _ := payload["occurred_at"].(string)
	eventType, _ := payload["event_type"].(string)

	raw := eventType + id + occurredAt
	hash := sha256.Sum256([]byte(raw))
	idemKey := fmt.Sprintf("%x", hash)

	job := jobqueue.Job{
		IdempotencyKey: idemKey,
		AppointmentID:  id,
		DoctorID:       doctorID,
		OccurredAt:     occurredAt,
		Channel:        "email",
		Recipient:      "patient@clinic.kz",
		Message:        fmt.Sprintf("Your appointment %s with doctor %s is complete.", id, doctorID),
	}

	subscriber.jobQueue.Enqueue(job)
}

func (subscriber *Subscriber) Drain() {
	if subscriber.conn != nil {
		err := subscriber.conn.Drain()
		if err != nil {
			fmt.Printf("WARN: error draining NATS connection: %v\n", err)
		}
		fmt.Println("NATS connection drained and closed")
	}
}
