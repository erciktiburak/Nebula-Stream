package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nebula-stream/cli/internal/health"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "health":
		return runHealth(args[1:])
	case "deploy":
		return runDeploy(args[1:])
	case "trigger":
		return runTrigger(args[1:])
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runHealth(args []string) error {
	fs := flag.NewFlagSet("health", flag.ContinueOnError)
	var nodeID string
	fs.StringVar(&nodeID, "node", "", "node id to query")

	if err := fs.Parse(args); err != nil {
		return err
	}

	report := health.NodeHealthSummary(nodeID)
	fmt.Println(health.Render(report))
	return nil
}

func runDeploy(args []string) error {
	fs := flag.NewFlagSet("deploy", flag.ContinueOnError)
	var file string
	fs.StringVar(&file, "file", "", "workflow yaml path")
	fs.StringVar(&file, "f", "", "workflow yaml path (shorthand)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if file == "" {
		return errors.New("deploy requires -f <workflow.yaml>")
	}

	if err := validateWorkflowFile(file); err != nil {
		return err
	}

	fmt.Printf("workflow accepted: %s\n", file)
	fmt.Println("deployment request queued")
	return nil
}

func runTrigger(args []string) error {
	fs := flag.NewFlagSet("trigger", flag.ContinueOnError)
	var (
		natsURL string
		subject string
		topic   string
		payload string
	)

	fs.StringVar(&natsURL, "nats", "nats://127.0.0.1:4222", "nats server URL")
	fs.StringVar(&subject, "subject", "nebula.events.ingest", "ingestion subject")
	fs.StringVar(&topic, "topic", "workflow.hello", "event topic")
	fs.StringVar(&payload, "payload", "{}", "event payload as JSON string")

	if err := fs.Parse(args); err != nil {
		return err
	}

	conn, err := nats.Connect(natsURL)
	if err != nil {
		return fmt.Errorf("connect to nats: %w", err)
	}
	defer conn.Close()

	envelope := map[string]any{
		"id":         fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		"topic":      topic,
		"payload":    []byte(payload),
		"created_at": time.Now().UTC(),
		"meta": map[string]string{
			"source": "nebula-cli",
		},
	}

	raw, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if err := conn.Publish(subject, raw); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}

	if err := conn.FlushWithContext(context.Background()); err != nil {
		return fmt.Errorf("flush publish: %w", err)
	}

	fmt.Printf("event published subject=%s topic=%s\n", subject, topic)
	return nil
}

func validateWorkflowFile(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".yaml" && ext != ".yml" {
		return fmt.Errorf("workflow file must be .yaml or .yml: %s", path)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read workflow file: %w", err)
	}

	body := string(raw)
	if !strings.Contains(body, "steps:") || !strings.Contains(body, "triggers:") {
		return errors.New("invalid workflow yaml: missing triggers or steps section")
	}

	return nil
}

func printUsage() {
	fmt.Println("nebula-cli commands:")
	fmt.Println("  health [--node <id>]          Show health status for a node")
	fmt.Println("  deploy -f <workflow.yaml>     Validate and queue workflow deployment")
	fmt.Println("  trigger [flags]               Publish workflow event to NATS")
}
