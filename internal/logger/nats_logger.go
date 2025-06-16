package logger

import (
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
)

type Event struct {
	ID        int       `json:"id"`
	ProjectID int       `json:"project_id"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}

type Logger interface {
	Publish(event Event) error
}

type NatsLogger struct {
	conn  *nats.Conn
	topic string
}

func NewNatsLogger(url, topic string) (*NatsLogger, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return &NatsLogger{
		conn:  nc,
		topic: topic,
	}, nil
}

func (l *NatsLogger) Publish(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return l.conn.Publish(l.topic, data)
}
