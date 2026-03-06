package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nebula-stream/engine/internal/bus"
	"github.com/nebula-stream/engine/internal/engine"
	"github.com/nebula-stream/engine/internal/state"
	"github.com/nebula-stream/engine/internal/workflow"
)

type Service struct {
	busClient *bus.Client
	registry  *workflow.Registry
	runner    *engine.Runner
	store     state.Store
}

type stepExecution struct {
	ID     string         `json:"id"`
	Type   string         `json:"type"`
	Output map[string]any `json:"output,omitempty"`
}

type executionRecord struct {
	EventID    string          `json:"event_id"`
	Workflow   string          `json:"workflow"`
	Topic      string          `json:"topic"`
	Payload    map[string]any  `json:"payload,omitempty"`
	StartedAt  time.Time       `json:"started_at"`
	DurationMs int64           `json:"duration_ms"`
	StepCount  int             `json:"step_count"`
	Results    []stepExecution `json:"results"`
}

func NewService(busClient *bus.Client, registry *workflow.Registry, store state.Store) *Service {
	return &Service{
		busClient: busClient,
		registry:  registry,
		runner:    engine.NewRunner(),
		store:     store,
	}
}

func (s *Service) Start(ctx context.Context, subject string) error {
	if s == nil || s.busClient == nil {
		return fmt.Errorf("ingestion service requires an initialized bus client")
	}

	if _, err := s.busClient.Subscribe(subject, func(event bus.EventEnvelope) error {
		return s.handle(event)
	}); err != nil {
		return err
	}

	active, ok := s.registry.Active()
	if ok {
		log.Printf("ingestion subscribed subject=%s workflow=%s", subject, active.Name)
	} else {
		log.Printf("ingestion subscribed subject=%s workflow=<none>", subject)
	}

	<-ctx.Done()
	return nil
}

func (s *Service) handle(event bus.EventEnvelope) error {
	log.Printf("event received id=%s topic=%s payload=%dB", event.ID, event.Topic, len(event.Payload))

	def, err := s.resolveWorkflow(event)
	if err != nil {
		return err
	}

	started := time.Now().UTC()
	results, err := s.runner.Execute(context.Background(), def, event)
	if err != nil {
		return err
	}
	durationMs := time.Since(started).Milliseconds()

	if err := s.persistExecution(def, event, started, durationMs, results); err != nil {
		return err
	}

	log.Printf("workflow executed name=%s steps=%d duration_ms=%d", def.Name, len(results), durationMs)

	return nil
}

func (s *Service) persistExecution(def workflow.Definition, event bus.EventEnvelope, startedAt time.Time, durationMs int64, results map[string]engine.StepResult) error {
	if s.store == nil {
		return nil
	}

	stepResults := make([]map[string]any, 0, len(def.Steps))
	execSteps := make([]stepExecution, 0, len(def.Steps))
	for _, step := range def.Steps {
		stepResult, ok := results[step.ID]
		if !ok {
			continue
		}

		stepResults = append(stepResults, map[string]any{
			"id":     step.ID,
			"type":   step.Type,
			"output": stepResult.Output,
		})

		execSteps = append(execSteps, stepExecution{
			ID:     step.ID,
			Type:   step.Type,
			Output: stepResult.Output,
		})
	}

	record := executionRecord{
		EventID:    event.ID,
		Workflow:   def.Name,
		Topic:      event.Topic,
		Payload:    decodeEventPayload(event.Payload),
		StartedAt:  startedAt,
		DurationMs: durationMs,
		StepCount:  len(stepResults),
		Results:    execSteps,
	}

	raw, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal execution record: %w", err)
	}

	if err := s.store.Save(executionKey(event.ID), raw); err != nil {
		return fmt.Errorf("save execution record: %w", err)
	}

	if err := s.store.Save(latestExecutionKey(def.Name), raw); err != nil {
		return fmt.Errorf("save latest execution record: %w", err)
	}

	history, err := s.loadHistory(def.Name)
	if err != nil {
		return err
	}

	history = append([]executionRecord{record}, history...)
	if len(history) > 20 {
		history = history[:20]
	}

	rawHistory, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("marshal execution history: %w", err)
	}

	if err := s.store.Save(historyExecutionKey(def.Name), rawHistory); err != nil {
		return fmt.Errorf("save execution history: %w", err)
	}

	return nil
}

func (s *Service) loadHistory(workflowName string) ([]executionRecord, error) {
	raw, err := s.store.Load(historyExecutionKey(workflowName))
	if err != nil {
		if errors.Is(err, state.ErrNotFound) {
			return []executionRecord{}, nil
		}
		return nil, fmt.Errorf("load execution history: %w", err)
	}

	var records []executionRecord
	if err := json.Unmarshal(raw, &records); err != nil {
		return nil, fmt.Errorf("decode execution history: %w", err)
	}

	return records, nil
}

func executionKey(eventID string) string {
	return fmt.Sprintf("execution:%s", eventID)
}

func latestExecutionKey(workflowName string) string {
	return fmt.Sprintf("workflow:%s:latest", workflowName)
}

func historyExecutionKey(workflowName string) string {
	return fmt.Sprintf("workflow:%s:history", workflowName)
}

func decodeEventPayload(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return map[string]any{}
	}

	return payload
}

func (s *Service) resolveWorkflow(event bus.EventEnvelope) (workflow.Definition, error) {
	if s.registry == nil {
		return workflow.Definition{}, fmt.Errorf("workflow registry is not initialized")
	}

	if name := workflowNameFromEvent(event); name != "" {
		if def, ok := s.registry.Get(name); ok {
			return def, nil
		}
		return workflow.Definition{}, fmt.Errorf("workflow %q not found", name)
	}

	def, ok := s.registry.Active()
	if !ok {
		return workflow.Definition{}, fmt.Errorf("no active workflow configured")
	}

	return def, nil
}

func workflowNameFromEvent(event bus.EventEnvelope) string {
	if name := event.Meta["workflow"]; name != "" {
		return name
	}

	parts := strings.Split(event.Topic, ".")
	if len(parts) >= 2 && parts[0] == "workflow" {
		return parts[1]
	}

	return ""
}
