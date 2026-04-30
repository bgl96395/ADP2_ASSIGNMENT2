package subscriber

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	conn          *nats.Conn
	subscriptions []*nats.Subscription
}

func Connect(natsURL string) (*Subscriber, error) {
	var conn *nats.Conn
	var err error

	delays := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second}
	for counter, delay := range delays {
		conn, err = nats.Connect(natsURL)
		if err == nil {
			break
		}
		fmt.Printf("WARN: attempt %d failed to connect to NATS: %v. Retrying in %v...", counter+1, err, delay)
		time.Sleep(delay)
	}
	if err != nil {
		return nil, err
	}

	fmt.Println("Notification Service connected to NATS")
	return &Subscriber{conn: conn}, nil
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
		fmt.Printf("Subscribed to subject: %s", subj)
	}
	return nil
}

func (subscriber *Subscriber) handleMessage(subject string, data []byte) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		fmt.Printf("ERROR: failed to deserialize message on %s: %v", subject, err)
		return
	}

	logEntry := map[string]any{
		"time":    time.Now().UTC().Format(time.RFC3339),
		"subject": subject,
		"event":   payload,
	}

	logBytes, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Printf("ERROR: failed to marshal log entry: %v", err)
		return
	}

	fmt.Println(string(logBytes))
}

func (subscriber *Subscriber) Drain() {
	if subscriber.conn != nil {
		if err := subscriber.conn.Drain(); err != nil {
			fmt.Printf("WARN: error draining NATS connection: %v", err)
		}
		fmt.Println("NATS connection drained and closed")
	}
}
