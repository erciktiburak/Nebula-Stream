package main

import (
	"log"

	"github.com/nebula-stream/engine/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	log.Printf("nebula-engine bootstrap started node=%s nats=%s heartbeat=%ds", cfg.NodeID, cfg.NATSURL, cfg.HeartbeatSecs)
}
