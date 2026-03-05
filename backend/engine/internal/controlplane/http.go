package controlplane

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nebula-stream/engine/internal/state"
	"github.com/nebula-stream/engine/internal/workflow"
)

type Server struct {
	registry *workflow.Registry
	store    state.Store
}

func NewServer(registry *workflow.Registry, store state.Store) *Server {
	return &Server{registry: registry, store: store}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/v1/workflows", s.handleWorkflows)
	mux.HandleFunc("/api/v1/executions/latest", s.handleLatestExecution)
	mux.HandleFunc("/api/v1/executions/", s.handleExecutionByID)
	return mux
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

func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
