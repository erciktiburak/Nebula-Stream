package engine

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/nebula-stream/engine/internal/bus"
	"github.com/nebula-stream/engine/internal/workflow"
)

var ErrUnknownStepType = errors.New("unknown workflow step type")

type StepContext struct {
	Event bus.EventEnvelope
	Step  workflow.Step
	State map[string]any
}

type StepResult struct {
	Output map[string]any
}

type StepHandler func(context.Context, StepContext) (StepResult, error)

type Runner struct {
	handlers map[string]StepHandler
}

func NewRunner() *Runner {
	r := &Runner{handlers: make(map[string]StepHandler)}
	r.Register("builtin.log", builtinLogHandler)
	r.Register("wasm", wasmPlaceholderHandler)
	r.Register("ai", aiPlaceholderHandler)
	return r
}

func (r *Runner) Register(kind string, handler StepHandler) {
	r.handlers[kind] = handler
}

func (r *Runner) Execute(ctx context.Context, def workflow.Definition, event bus.EventEnvelope) (map[string]StepResult, error) {
	results := make(map[string]StepResult, len(def.Steps))
	state := make(map[string]any)

	for _, step := range def.Steps {
		handler, ok := r.selectHandler(step.Type)
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownStepType, step.Type)
		}

		result, err := handler(ctx, StepContext{Event: event, Step: step, State: state})
		if err != nil {
			return nil, fmt.Errorf("run step %s (%s): %w", step.ID, step.Type, err)
		}

		results[step.ID] = result
		state[step.ID] = result.Output
	}

	return results, nil
}

func (r *Runner) selectHandler(stepType string) (StepHandler, bool) {
	handler, ok := r.handlers[stepType]
	if ok {
		return handler, true
	}

	parts := strings.Split(stepType, ".")
	if len(parts) < 2 {
		return nil, false
	}

	handler, ok = r.handlers[parts[0]]
	return handler, ok
}

func builtinLogHandler(_ context.Context, stepCtx StepContext) (StepResult, error) {
	message, _ := stepCtx.Step.Input["message"].(string)
	if message == "" {
		message = fmt.Sprintf("event topic=%s payload=%dB", stepCtx.Event.Topic, len(stepCtx.Event.Payload))
	}

	log.Printf("builtin.log step=%s message=%s", stepCtx.Step.ID, message)

	return StepResult{Output: map[string]any{"message": message}}, nil
}

func wasmPlaceholderHandler(_ context.Context, stepCtx StepContext) (StepResult, error) {
	module, _ := stepCtx.Step.Input["module"].(string)
	if module == "" {
		module = "default.wasm"
	}

	log.Printf("wasm step=%s module=%s", stepCtx.Step.ID, module)
	return StepResult{Output: map[string]any{"module": module, "status": "ok"}}, nil
}

func aiPlaceholderHandler(_ context.Context, stepCtx StepContext) (StepResult, error) {
	model, _ := stepCtx.Step.Input["model"].(string)
	if model == "" {
		model = "onnx-default"
	}

	log.Printf("ai step=%s model=%s", stepCtx.Step.ID, model)
	return StepResult{Output: map[string]any{"model": model, "result": "unknown"}}, nil
}
