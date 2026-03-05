package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
}
