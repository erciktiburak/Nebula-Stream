# Nebula-Stream

The High-Performance Edge-to-Cloud Distributed AI Orchestrator.

[![Go Version](https://img.shields.io/badge/go-1.24-00ADD8.svg)](https://go.dev/)
[![CI](https://img.shields.io/badge/ci-github_actions-24292e.svg)](./.github/workflows/ci.yml)
[![Release](https://img.shields.io/badge/release-v1.0.0--beta-2ea44f.svg)](./CHANGELOG.md)

Nebula-Stream is an event-driven orchestration platform for microservices, AI workers, and serverless functions. It combines Go concurrency, NATS messaging, and WebAssembly sandboxing to execute workflow pipelines with low latency and strong isolation.

## Why Nebula-Stream

- Build multi-step automations using one YAML workflow definition.
- Run untrusted plugin logic in WASM sandboxes (Rust and TinyGo targets).
- Route events across distributed worker nodes using NATS/JetStream patterns.
- Observe execution paths with OpenTelemetry-first design.
- Extend to visual flow editing and real-time operations dashboards.

## System Architecture

Nebula-Stream follows an event-driven control plane with distributed execution workers.

```text
Trigger/API -> Orchestrator Core (Go) -> NATS Event Bus -> Worker Mesh
                                                |              |
                                                |              +--> AI Workers (ONNX)
                                                |
                                                +--> WASM Runtime (Wasmtime)

Telemetry: OpenTelemetry traces + Prometheus metrics + Dashboard streams
```

Core components:

1. **Orchestrator Core (Go):** workflow scheduling, routing, state transitions.
2. **Event Bus (NATS/JetStream):** pub/sub transport, retries, dead-letter handling.
3. **WASM Runtime (Wasmtime):** secure plugin execution boundary.
4. **AI Workers:** model-specific processing nodes (ONNX-oriented).
5. **API Layer (gRPC):** external event triggers and control-plane surfaces.
6. **Web UI (Next.js + React Flow):** workflow authoring and node telemetry views.

## Project Layout

```text
Nebula-Stream/
  backend/
    engine/                # Orchestrator and runtime integration
    cli/                   # nebula-cli operational commands
  proto/nebula/v1/         # Internal gRPC and event schemas
  workflows/examples/      # Workflow YAML samples
  plugins/                 # WASM plugin examples and notes
  docs/                    # Architecture, benchmark, security notes
  deploy/                  # Local infrastructure definitions
  web/                     # Dashboard app surface
```

## Quick Start

Requirements:

- Go 1.24+
- Make
- Docker (for local NATS)

Bootstrap and validate:

```bash
make check
```

Start local messaging dependency:

```bash
docker compose -f deploy/docker-compose.yml up -d
```

Run entrypoints:

```bash
go run ./backend/engine/cmd/engine
go run ./backend/cli/cmd/nebula-cli
```

## Example Workflow

See `workflows/examples/hello-world.yaml` for a minimal trigger + step definition.

## Performance Target

- Synthetic benchmark target: **50,000 events/second**.
- Details: `docs/benchmark.md`.

## Roadmap

The implementation was organized in three phases:

- Core architecture and messaging mesh
- Workflow engine and WASM runtime
- UI, telemetry, and production hardening

Commit-by-commit progress log: `docs/roadmap-progress.md`.

## Release

- Current milestone: **v1.0.0-beta - The First Constellation**
- Changelog: `CHANGELOG.md`

## License

TBD
