# Nebula-Stream Architecture

This document tracks high-level architecture decisions during the 50-commit roadmap.

## Initial Boundaries

- `backend/engine`: orchestration core and runtime integration
- `backend/cli`: operational tooling and deployment commands
- `proto/`: event and control plane schemas
- `workflows/`: YAML-based workflow definitions
- `plugins/`: WebAssembly plugin artifacts and examples
- `web/`: visual workflow and telemetry dashboard

## Decision Log

### ADR-0001: Repository Bootstrap

- Date: 2026-03-05
- Status: Accepted
- Context: Need a scalable monorepo layout for distributed systems components.
- Decision: Start with Go workspace + multi-module backend skeleton.
- Consequence: Enables independent service versioning while preserving monorepo cohesion.
