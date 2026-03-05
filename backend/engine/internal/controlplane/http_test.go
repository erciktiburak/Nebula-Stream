package controlplane

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nebula-stream/engine/internal/state"
	"github.com/nebula-stream/engine/internal/workflow"
)

func TestWorkflowDeployEndpoint(t *testing.T) {
	registry := workflow.NewRegistry(workflow.Definition{Name: "hello", Version: "v1", Triggers: []workflow.Trigger{{Type: "manual"}}, Steps: []workflow.Step{{ID: "s1", Type: "builtin.log"}}})
	srv := NewServer(registry, state.NewMemoryStore())

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

	registry := workflow.NewRegistry(workflow.Definition{Name: "hello", Version: "v1", Triggers: []workflow.Trigger{{Type: "manual"}}, Steps: []workflow.Step{{ID: "s1", Type: "builtin.log"}}})
	srv := NewServer(registry, store)

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
}
