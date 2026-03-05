package ingestion

import (
	"context"
	"fmt"
	"log"

	"github.com/nebula-stream/engine/internal/bus"
	"github.com/nebula-stream/engine/internal/workflow"
)

type Service struct {
	busClient *bus.Client
	workflow  workflow.Definition
}

func NewService(busClient *bus.Client, def workflow.Definition) *Service {
	return &Service{busClient: busClient, workflow: def}
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

	log.Printf("ingestion subscribed subject=%s workflow=%s", subject, s.workflow.Name)

	<-ctx.Done()
	return nil
}

func (s *Service) handle(event bus.EventEnvelope) error {
	log.Printf("event received id=%s topic=%s payload=%dB", event.ID, event.Topic, len(event.Payload))

	for _, step := range s.workflow.Steps {
		log.Printf("workflow step id=%s type=%s", step.ID, step.Type)
	}

	return nil
}
