# Nebula Dashboard

This package contains a running Next.js demo for Nebula-Stream workflow visualization.

## Features in this iteration

- React Flow based pipeline graph
- Mock real-time telemetry stream
- Live control-plane polling for workflow and latest execution
- Execution history timeline from control-plane (`/api/v1/executions/history`)
- Manual trigger button (`POST /api/v1/triggers`)
- Workflow selector for targeting specific deployed workflows
- Set active workflow button (`POST /api/v1/workflows/active`)
- Throughput, node count, latency, and error widgets
- Live node log feed panel

## Run locally

```bash
npm install
npm run dev
```

Open `http://localhost:3000`.

To connect dashboard to a running engine API:

```bash
NEXT_PUBLIC_ENGINE_URL=http://127.0.0.1:8080 npm run dev
```
