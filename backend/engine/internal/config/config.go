package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	NodeID        string
	NATSURL       string
	HeartbeatSecs int
	WorkflowPath  string
	IngestSubject string
}

func Load() (Config, error) {
	cfg := Config{
		NodeID:        getEnv("NEBULA_NODE_ID", "engine-local"),
		NATSURL:       getEnv("NEBULA_NATS_URL", "nats://127.0.0.1:4222"),
		HeartbeatSecs: 5,
		WorkflowPath:  getEnv("NEBULA_WORKFLOW_PATH", "workflows/examples/hello-world.yaml"),
		IngestSubject: getEnv("NEBULA_INGEST_SUBJECT", "nebula.events.ingest"),
	}

	if v := os.Getenv("NEBULA_HEARTBEAT_SECONDS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("invalid NEBULA_HEARTBEAT_SECONDS: %q", v)
		}
		cfg.HeartbeatSecs = n
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
