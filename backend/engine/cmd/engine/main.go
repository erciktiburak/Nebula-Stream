package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/nebula-stream/engine/internal/bus"
	"github.com/nebula-stream/engine/internal/config"
	"github.com/nebula-stream/engine/internal/ingestion"
	"github.com/nebula-stream/engine/internal/workflow"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	log.Printf("nebula-engine bootstrap started node=%s nats=%s heartbeat=%ds", cfg.NodeID, cfg.NATSURL, cfg.HeartbeatSecs)

	def, err := workflow.ParseFile(cfg.WorkflowPath)
	if err != nil {
		log.Fatalf("parse workflow file %q: %v", cfg.WorkflowPath, err)
	}

	busClient, err := bus.Connect(cfg.NATSURL)
	if err != nil {
		log.Fatalf("connect bus: %v", err)
	}
	defer busClient.Close()

	svc := ingestion.NewService(busClient, def)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := svc.Start(ctx, cfg.IngestSubject); err != nil {
		log.Fatalf("start ingestion: %v", err)
	}

	log.Println("engine shutdown complete")
}
