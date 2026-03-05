package bus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

const DefaultNATSURL = "nats://127.0.0.1:4222"

const SerializationFormat = "protobuf"

const AutoReconnect = true

type EventEnvelope struct {
	ID        string            `json:"id"`
	Topic     string            `json:"topic"`
	Payload   []byte            `json:"payload"`
	CreatedAt time.Time         `json:"created_at"`
	Meta      map[string]string `json:"meta,omitempty"`
}

type Client struct {
	conn *nats.Conn
}

func Connect(url string) (*Client, error) {
	if url == "" {
		url = DefaultNATSURL
	}

	conn, err := nats.Connect(
		url,
		nats.Name("nebula-engine"),
		nats.RetryOnFailedConnect(AutoReconnect),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to nats: %w", err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Close() {
	if c == nil || c.conn == nil {
		return
	}
	c.conn.Close()
}

func (c *Client) Publish(ctx context.Context, subject string, event EventEnvelope) error {
	if c == nil || c.conn == nil {
		return errors.New("nats connection is not initialized")
	}
	if subject == "" {
		return errors.New("subject is required")
	}

	raw, err := Encode(event)
	if err != nil {
		return err
	}

	if err := c.conn.Publish(subject, raw); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}

	if err := c.conn.FlushWithContext(ctx); err != nil {
		return fmt.Errorf("flush publish: %w", err)
	}

	return nil
}

func (c *Client) Subscribe(subject string, handler func(EventEnvelope) error) (*nats.Subscription, error) {
	if c == nil || c.conn == nil {
		return nil, errors.New("nats connection is not initialized")
	}
	if subject == "" {
		return nil, errors.New("subject is required")
	}

	sub, err := c.conn.Subscribe(subject, func(msg *nats.Msg) {
		event, decodeErr := Decode(msg.Data)
		if decodeErr != nil {
			return
		}
		_ = handler(event)
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe: %w", err)
	}

	return sub, nil
}

func Encode(event EventEnvelope) ([]byte, error) {
	if event.ID == "" {
		return nil, errors.New("event id is required")
	}
	if event.Topic == "" {
		return nil, errors.New("event topic is required")
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}

	raw, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("marshal event: %w", err)
	}

	return raw, nil
}

func Decode(raw []byte) (EventEnvelope, error) {
	var event EventEnvelope
	if err := json.Unmarshal(raw, &event); err != nil {
		return EventEnvelope{}, fmt.Errorf("unmarshal event: %w", err)
	}

	if event.ID == "" || event.Topic == "" {
		return EventEnvelope{}, errors.New("event id and topic are required")
	}

	return event, nil
}
