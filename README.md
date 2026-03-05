# Nebula-Stream

**The High-Performance Edge-to-Cloud Distributed AI Orchestrator**

Nebula-Stream is an event-driven orchestration platform for microservices, AI workers, and serverless functions. It is designed for low-latency execution and clear operational visibility.

The long-term goal is to let users define complex flows (for example: ingest video -> run AI inference -> notify Slack) through a single YAML workflow file.

## Vision

- High-throughput event mesh across distributed worker nodes
- WebAssembly-based isolated plugin runtime (multi-language support)
- gRPC and Protocol Buffers for efficient inter-service communication
- OpenTelemetry-native tracing and metrics
- Visual workflow builder and real-time telemetry dashboard

## Architecture (Planned)

Core components:

1. **Orchestrator Core (Go)**: workflow scheduling, routing, state transitions
2. **Event Bus (NATS/JetStream)**: pub/sub, retries, dead-letter queues
3. **WASM Runtime (Wasmtime)**: secure plugin execution sandbox
4. **AI Workers**: specialized nodes (ONNX and model-serving tasks)
5. **API Layer (gRPC + Gateway)**: external triggers and control plane APIs
6. **Web UI (Next.js + React Flow)**: visual workflow authoring and observability

## Repository Layout

```text
Nebula-Stream/
  backend/
    engine/           # Go orchestrator service
    cli/              # Go CLI client
  proto/              # Protocol Buffers definitions
  workflows/          # Sample YAML workflows
  plugins/            # WASM plugin examples
  docs/               # Architecture and design docs
  deploy/             # Local/dev deployment assets
  web/                # Dashboard app (Next.js)
```

## Current Status

This is the bootstrap phase. Initial repository scaffolding is in place to support a 50-commit implementation roadmap.

## Getting Started (Bootstrap)

Requirements:

- Go 1.22+
- Make

Run local bootstrap checks:

```bash
make check
```

## Roadmap

The implementation plan is organized in 3 stages:

- **Stage 1**: Core architecture and messaging mesh
- **Stage 2**: Workflow engine and WASM runtime
- **Stage 3**: UI, telemetry, and production hardening

Detailed milestone breakdown will be tracked in `docs/architecture.md` and upcoming project docs.

## License

TBD
# Nebula-Stream

## System Architecture

Nebula-Stream uses an event-driven control plane with distributed workers and WASM runtime sandboxes.
