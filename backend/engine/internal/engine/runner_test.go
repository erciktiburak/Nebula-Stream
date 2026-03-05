package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/nebula-stream/engine/internal/bus"
	"github.com/nebula-stream/engine/internal/workflow"
)

func TestRunnerExecute(t *testing.T) {
	runner := NewRunner()

	def := workflow.Definition{
		Version: "v1",
		Name:    "demo",
		Triggers: []workflow.Trigger{
			{Type: "manual"},
		},
		Steps: []workflow.Step{
			{ID: "s1", Type: "builtin.log", Input: map[string]any{"message": "hello"}},
			{ID: "s2", Type: "wasm.transform", Input: map[string]any{"module": "transform.wasm"}},
			{ID: "s3", Type: "ai.onnx", Input: map[string]any{"model": "resnet50"}},
		},
	}

	results, err := runner.Execute(context.Background(), def, bus.EventEnvelope{ID: "evt-1", Topic: "workflow.hello", Payload: []byte("{}")})
	if err != nil {
		t.Fatalf("execute workflow: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 step results, got %d", len(results))
	}
}

func TestRunnerUnknownStepType(t *testing.T) {
	runner := NewRunner()
	def := workflow.Definition{
		Version: "v1",
		Name:    "demo",
		Triggers: []workflow.Trigger{
			{Type: "manual"},
		},
		Steps: []workflow.Step{
			{ID: "s1", Type: "custom.step"},
		},
	}

	_, err := runner.Execute(context.Background(), def, bus.EventEnvelope{ID: "evt-2", Topic: "workflow.hello"})
	if !errors.Is(err, ErrUnknownStepType) {
		t.Fatalf("expected ErrUnknownStepType, got %v", err)
	}
}
