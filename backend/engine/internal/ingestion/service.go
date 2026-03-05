package ingestion

import (
	"context"
	"fmt"
	"log"

	"github.com/nebula-stream/engine/internal/bus"
	"github.com/nebula-stream/engine/internal/engine"
	"github.com/nebula-stream/engine/internal/workflow"
)

type Service struct {
	busClient *bus.Client
	workflow  workflow.Definition
	runner    *engine.Runner
}

func NewService(busClient *bus.Client, def workflow.Definition) *Service {
	return &Service{
		busClient: busClient,
		workflow:  def,
		runner:    engine.NewRunner(),
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

	log.Printf("ingestion subscribed subject=%s workflow=%s", subject, s.workflow.Name)

	<-ctx.Done()
	return nil
}

func (s *Service) handle(event bus.EventEnvelope) error {
	log.Printf("event received id=%s topic=%s payload=%dB", event.ID, event.Topic, len(event.Payload))

	results, err := s.runner.Execute(context.Background(), s.workflow, event)
	if err != nil {
		return err
	}

	log.Printf("workflow executed name=%s steps=%d", s.workflow.Name, len(results))

	return nil
}
