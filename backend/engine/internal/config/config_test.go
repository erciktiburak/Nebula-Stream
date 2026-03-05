package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("NEBULA_NODE_ID", "")
	t.Setenv("NEBULA_NATS_URL", "")
	t.Setenv("NEBULA_HEARTBEAT_SECONDS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.NodeID == "" || cfg.NATSURL == "" || cfg.HeartbeatSecs <= 0 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}

	if cfg.WorkflowPath == "" || cfg.IngestSubject == "" {
		t.Fatalf("expected workflow and subject defaults: %+v", cfg)
	}
}

func TestLoadInvalidHeartbeat(t *testing.T) {
	t.Setenv("NEBULA_HEARTBEAT_SECONDS", "abc")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid heartbeat error")
	}
}
