package bus

import (
	"testing"
	"time"
)

func TestDefaultNATSURL(t *testing.T) {
	if DefaultNATSURL == "" {
		t.Fatal("default URL must not be empty")
	}
}

func TestEncodeDecode(t *testing.T) {
	raw, err := Encode(EventEnvelope{
		ID:      "evt-1",
		Topic:   "workflow.start",
		Payload: []byte(`{"foo":"bar"}`),
		Meta:    map[string]string{"source": "test"},
	})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	event, err := Decode(raw)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if event.ID != "evt-1" {
		t.Fatalf("unexpected id: %s", event.ID)
	}

	if event.CreatedAt.Before(time.Now().Add(-1 * time.Minute)) {
		t.Fatalf("expected auto-created timestamp, got: %v", event.CreatedAt)
	}
}
