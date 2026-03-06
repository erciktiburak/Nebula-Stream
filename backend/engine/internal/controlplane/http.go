package controlplane

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nebula-stream/engine/internal/bus"
	"github.com/nebula-stream/engine/internal/state"
	"github.com/nebula-stream/engine/internal/workflow"
)

type EventPublisher interface {
	Publish(ctx context.Context, subject string, event bus.EventEnvelope) error
}

type Server struct {
	registry *workflow.Registry
	store    state.Store
	pub      EventPublisher
	subject  string
}

func NewServer(registry *workflow.Registry, store state.Store, pub EventPublisher, subject string) *Server {
	return &Server{registry: registry, store: store, pub: pub, subject: subject}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/v1/workflows", s.handleWorkflows)
	mux.HandleFunc("/api/v1/workflows/active", s.handleActiveWorkflow)
	mux.HandleFunc("/api/v1/executions/latest", s.handleLatestExecution)
	mux.HandleFunc("/api/v1/executions/history", s.handleExecutionHistory)
	mux.HandleFunc("/api/v1/executions/", s.handleExecutionByID)
	mux.HandleFunc("/api/v1/triggers", s.handleTrigger)
	return mux
}

type activeWorkflowRequest struct {
	Workflow string `json:"workflow"`
}

func (s *Server) handleActiveWorkflow(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		active, ok := s.registry.Active()
		if !ok {
			writeErr(w, http.StatusNotFound, fmt.Errorf("no active workflow configured"))
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"workflow": active.Name})
	case http.MethodPost:
		var req activeWorkflowRequest
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, fmt.Errorf("decode active workflow request: %w", err))
			return
		}

		if req.Workflow == "" {
			writeErr(w, http.StatusBadRequest, fmt.Errorf("workflow is required"))
			return
		}

		if err := s.registry.SetActive(req.Workflow); err != nil {
			writeErr(w, http.StatusNotFound, err)
			return
		}

		writeJSON(w, http.StatusAccepted, map[string]string{
			"status":   "accepted",
			"workflow": req.Workflow,
		})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleWorkflows(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleWorkflowDeploy(w, r)
	case http.MethodGet:
		active, _ := s.registry.Active()
		writeJSON(w, http.StatusOK, map[string]any{
			"active":    active.Name,
			"workflows": s.registry.Names(),
		})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleWorkflowDeploy(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("read request body: %w", err))
		return
	}

	def, err := workflow.ParseYAML(raw)
	if err != nil {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("parse workflow yaml: %w", err))
		return
	}

	s.registry.Upsert(def)
	if r.URL.Query().Get("activate") != "false" {
		_ = s.registry.SetActive(def.Name)
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":   "accepted",
		"workflow": def.Name,
	})
}

func (s *Server) handleExecutionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/executions/")
	if id == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("execution id is required"))
		return
	}

	if s.store == nil {
		writeErr(w, http.StatusNotFound, fmt.Errorf("state store is not configured"))
		return
	}

	raw, err := s.store.Load("execution:" + id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

func (s *Server) handleLatestExecution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if s.store == nil {
		writeErr(w, http.StatusNotFound, fmt.Errorf("state store is not configured"))
		return
	}

	name := r.URL.Query().Get("workflow")
	if name == "" {
		active, ok := s.registry.Active()
		if !ok {
			writeErr(w, http.StatusNotFound, fmt.Errorf("no active workflow configured"))
			return
		}
		name = active.Name
	}

	raw, err := s.store.Load("workflow:" + name + ":latest")
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

func (s *Server) handleExecutionHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if s.store == nil {
		writeErr(w, http.StatusNotFound, fmt.Errorf("state store is not configured"))
		return
	}

	name := r.URL.Query().Get("workflow")
	if name == "" {
		active, ok := s.registry.Active()
		if !ok {
			writeErr(w, http.StatusNotFound, fmt.Errorf("no active workflow configured"))
			return
		}
		name = active.Name
	}

	limit := 10
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			writeErr(w, http.StatusBadRequest, fmt.Errorf("invalid limit value: %s", v))
			return
		}
		if n < 50 {
			limit = n
		}
	}

	raw, err := s.store.Load("workflow:" + name + ":history")
	if err != nil {
		if errors.Is(err, state.ErrNotFound) {
			writeJSON(w, http.StatusOK, []any{})
			return
		}
		writeErr(w, http.StatusNotFound, err)
		return
	}

	var records []map[string]any
	if err := json.Unmarshal(raw, &records); err != nil {
		writeErr(w, http.StatusInternalServerError, fmt.Errorf("decode execution history: %w", err))
		return
	}

	if len(records) > limit {
		records = records[:limit]
	}

	writeJSON(w, http.StatusOK, records)
}

type triggerRequest struct {
	Workflow string         `json:"workflow"`
	Topic    string         `json:"topic"`
	Payload  map[string]any `json:"payload"`
}

func (s *Server) handleTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if s.pub == nil {
		writeErr(w, http.StatusServiceUnavailable, fmt.Errorf("event publisher is not configured"))
		return
	}

	var req triggerRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("decode trigger request: %w", err))
		return
	}

	workflowName := req.Workflow
	if workflowName == "" {
		active, ok := s.registry.Active()
		if !ok {
			writeErr(w, http.StatusBadRequest, fmt.Errorf("workflow is required when no active workflow exists"))
			return
		}
		workflowName = active.Name
	}

	topic := req.Topic
	if topic == "" {
		topic = fmt.Sprintf("workflow.%s.run", workflowName)
	}

	rawPayload, err := json.Marshal(req.Payload)
	if err != nil {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("encode payload: %w", err))
		return
	}

	eventID := fmt.Sprintf("evt-%d", time.Now().UnixNano())
	event := bus.EventEnvelope{
		ID:        eventID,
		Topic:     topic,
		Payload:   rawPayload,
		CreatedAt: time.Now().UTC(),
		Meta: map[string]string{
			"source":   "controlplane",
			"workflow": workflowName,
		},
	}

	if err := s.pub.Publish(r.Context(), s.subject, event); err != nil {
		writeErr(w, http.StatusBadGateway, fmt.Errorf("publish trigger event: %w", err))
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":   "accepted",
		"event_id": eventID,
		"workflow": workflowName,
		"topic":    topic,
	})
}

func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
