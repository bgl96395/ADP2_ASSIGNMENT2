package event

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type EventPublisher interface {
	Publish(subject string, payload any) error
}

type NATSPublisher struct {
	conn *nats.Conn
}

func NewNATSPublisher(natsURL string) (*NATSPublisher, error) {
	conn, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}
	return &NATSPublisher{conn: conn}, nil
}

func (publisher *NATSPublisher) Publish(subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return publisher.conn.Publish(subject, data)
}

func (publisher *NATSPublisher) Close() {
	publisher.conn.Close()
}

type NoopPublisher struct{}

func (noopPublisher *NoopPublisher) Publish(subject string, payload any) error {
	log.Printf("WARN: broker unavailable, event %s dropped", subject)
	return nil
}

type DoctorCreatedEvent struct {
	EventType      string    `json:"event_type"`
	OccurredAt     time.Time `json:"occurred_at"`
	ID             string    `json:"id"`
	FullName       string    `json:"full_name"`
	Specialization string    `json:"specialization"`
	Email          string    `json:"email"`
}
