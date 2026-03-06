package controlplane

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nebula-stream/engine/internal/bus"
	"github.com/nebula-stream/engine/internal/state"
	"github.com/nebula-stream/engine/internal/workflow"
)

type testPublisher struct {
	events []bus.EventEnvelope
}

func (p *testPublisher) Publish(_ context.Context, _ string, event bus.EventEnvelope) error {
	p.events = append(p.events, event)
	return nil
}

func TestWorkflowDeployEndpoint(t *testing.T) {
	registry := workflow.NewRegistry(workflow.Definition{Name: "hello", Version: "v1", Triggers: []workflow.Trigger{{Type: "manual"}}, Steps: []workflow.Step{{ID: "s1", Type: "builtin.log"}}})
	srv := NewServer(registry, state.NewMemoryStore(), &testPublisher{}, "nebula.events.ingest")

	body := `version: v1
name: uploaded
triggers:
  - type: manual
steps:
  - id: s1
    type: builtin.log
`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows", strings.NewReader(body))
	res := httptest.NewRecorder()

	srv.Handler().ServeHTTP(res, req)

	if res.Code != http.StatusAccepted {
		t.Fatalf("unexpected status code: %d", res.Code)
	}

	active, ok := registry.Active()
	if !ok || active.Name != "uploaded" {
		t.Fatalf("expected uploaded workflow to be active, got %+v", active)
	}
}

func TestExecutionEndpoint(t *testing.T) {
	store := state.NewMemoryStore()
	if err := store.Save("execution:evt-123", []byte(`{"event_id":"evt-123"}`)); err != nil {
		t.Fatalf("seed store: %v", err)
	}
	if err := store.Save("workflow:hello:latest", []byte(`{"event_id":"evt-999","workflow":"hello"}`)); err != nil {
		t.Fatalf("seed latest: %v", err)
	}
	if err := store.Save("workflow:hello:history", []byte(`[{"event_id":"evt-999"},{"event_id":"evt-998"}]`)); err != nil {
		t.Fatalf("seed history: %v", err)
	}

	registry := workflow.NewRegistry(workflow.Definition{Name: "hello", Version: "v1", Triggers: []workflow.Trigger{{Type: "manual"}}, Steps: []workflow.Step{{ID: "s1", Type: "builtin.log"}}})
	srv := NewServer(registry, store, &testPublisher{}, "nebula.events.ingest")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions/evt-123", nil)
	res := httptest.NewRecorder()
	srv.Handler().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", res.Code)
	}

	latestReq := httptest.NewRequest(http.MethodGet, "/api/v1/executions/latest?workflow=hello", nil)
	latestRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(latestRes, latestReq)
	if latestRes.Code != http.StatusOK {
		t.Fatalf("unexpected latest status code: %d", latestRes.Code)
	}

	historyReq := httptest.NewRequest(http.MethodGet, "/api/v1/executions/history?workflow=hello&limit=1", nil)
	historyRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(historyRes, historyReq)
	if historyRes.Code != http.StatusOK {
		t.Fatalf("unexpected history status code: %d", historyRes.Code)
	}
}

func TestTriggerEndpoint(t *testing.T) {
	pub := &testPublisher{}
	registry := workflow.NewRegistry(workflow.Definition{Name: "hello", Version: "v1", Triggers: []workflow.Trigger{{Type: "manual"}}, Steps: []workflow.Step{{ID: "s1", Type: "builtin.log"}}})
	srv := NewServer(registry, state.NewMemoryStore(), pub, "nebula.events.ingest")

	body := `{"workflow":"hello","payload":{"message":"from-ui"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/triggers", strings.NewReader(body))
	res := httptest.NewRecorder()

	srv.Handler().ServeHTTP(res, req)

	if res.Code != http.StatusAccepted {
		t.Fatalf("unexpected status code: %d", res.Code)
	}

	if len(pub.events) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(pub.events))
	}

	if pub.events[0].Meta["workflow"] != "hello" {
		t.Fatalf("unexpected workflow meta: %s", pub.events[0].Meta["workflow"])
	}

	var payload map[string]any
	if err := json.Unmarshal(pub.events[0].Payload, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	if payload["message"] != "from-ui" {
		t.Fatalf("unexpected payload message: %v", payload["message"])
	}
}

func TestActiveWorkflowEndpoint(t *testing.T) {
	pub := &testPublisher{}
	registry := workflow.NewRegistry(workflow.Definition{Name: "hello", Version: "v1", Triggers: []workflow.Trigger{{Type: "manual"}}, Steps: []workflow.Step{{ID: "s1", Type: "builtin.log"}}})
	registry.Upsert(workflow.Definition{Name: "image", Version: "v1", Triggers: []workflow.Trigger{{Type: "manual"}}, Steps: []workflow.Step{{ID: "s1", Type: "builtin.log"}}})
	srv := NewServer(registry, state.NewMemoryStore(), pub, "nebula.events.ingest")

	setReq := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/active", strings.NewReader(`{"workflow":"image"}`))
	setRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(setRes, setReq)
	if setRes.Code != http.StatusAccepted {
		t.Fatalf("unexpected set active status code: %d", setRes.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/workflows/active", nil)
	getRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(getRes, getReq)
	if getRes.Code != http.StatusOK {
		t.Fatalf("unexpected get active status code: %d", getRes.Code)
	}
}
