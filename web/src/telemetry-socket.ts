'use client'

import { useEffect, useState } from 'react'

import type { EdgeMessage } from '@/message-flow'
import type { NodeLog, NodeLogLevel } from '@/node-logs'

type TelemetrySnapshot = {
  throughput: number
  activeNodes: number
  latencyMs: number
  logs: NodeLog[]
  messages: EdgeMessage[]
  mode: 'mock' | 'live'
  activeWorkflow: string
}

const pipelineEdges = [
  ['trigger', 'orchestrator'],
  ['orchestrator', 'nats'],
  ['nats', 'wasm'],
  ['nats', 'ai'],
  ['wasm', 'state'],
  ['ai', 'state'],
] as const

const logLevels: NodeLogLevel[] = ['info', 'warn', 'error']

function nextEdge(index: number): EdgeMessage {
  const [from, to] = pipelineEdges[index % pipelineEdges.length]
  const now = Date.now()
  return {
    id: `${from}-${to}-${now}`,
    from,
    to,
    at: now,
  }
}

function nextLog(index: number): NodeLog {
  const level = logLevels[index % logLevels.length]
  const now = Date.now()
  const nodeId = index % 2 === 0 ? 'wasm-node-a1' : 'ai-node-b2'
  const messageByLevel: Record<NodeLogLevel, string> = {
    info: 'step completed in expected latency budget',
    warn: 'queue depth increased, applying soft backpressure',
    error: 'transient transport error, retrying on secondary route',
  }

  return {
    id: `${nodeId}-${now}`,
    nodeId,
    level,
    message: messageByLevel[level],
    at: now,
  }
}

type WorkflowResponse = {
  active: string
  workflows: string[]
}

type ExecutionResult = {
  id: string
  type: string
  output?: Record<string, unknown>
}

type ExecutionRecord = {
  event_id: string
  workflow: string
  topic: string
  duration_ms: number
  step_count: number
  results: ExecutionResult[]
}

function toNode(stepType: string): string {
  if (stepType.startsWith('wasm')) {
    return 'wasm'
  }
  if (stepType.startsWith('ai')) {
    return 'ai'
  }
  return 'orchestrator'
}

function fromExecution(record: ExecutionRecord): Pick<TelemetrySnapshot, 'throughput' | 'activeNodes' | 'latencyMs' | 'logs' | 'messages' | 'activeWorkflow'> {
  const now = Date.now()

  const logs: NodeLog[] = record.results.map((step, index) => ({
    id: `${record.event_id}-${step.id}`,
    nodeId: toNode(step.type),
    level: 'info',
    message: `${step.id} (${step.type}) executed`,
    at: now - index,
  }))

  const messages: EdgeMessage[] = [
    { id: `${record.event_id}-trigger-orch`, from: 'trigger', to: 'orchestrator', at: now },
    { id: `${record.event_id}-orch-nats`, from: 'orchestrator', to: 'nats', at: now },
  ]

  record.results.forEach((step, index) => {
    const node = toNode(step.type)
    messages.push({ id: `${record.event_id}-nats-${index}`, from: 'nats', to: node, at: now + index })
    if (node !== 'orchestrator') {
      messages.push({ id: `${record.event_id}-state-${index}`, from: node, to: 'state', at: now + index + 1 })
    }
  })

  return {
    throughput: Math.max(1, Math.floor(1000 / Math.max(1, record.duration_ms))),
    activeNodes: 6,
    latencyMs: record.duration_ms,
    logs,
    messages,
    activeWorkflow: record.workflow,
  }
}

async function fetchLiveSnapshot(engineURL: string): Promise<Pick<TelemetrySnapshot, 'throughput' | 'activeNodes' | 'latencyMs' | 'logs' | 'messages' | 'activeWorkflow'> | null> {
  const workflowsRes = await fetch(`${engineURL}/api/v1/workflows`, { cache: 'no-store' })
  if (!workflowsRes.ok) {
    return null
  }

  const workflows = (await workflowsRes.json()) as WorkflowResponse
  if (!workflows.active) {
    return null
  }

  const latestRes = await fetch(`${engineURL}/api/v1/executions/latest?workflow=${encodeURIComponent(workflows.active)}`, { cache: 'no-store' })
  if (!latestRes.ok) {
    return null
  }

  const latest = (await latestRes.json()) as ExecutionRecord
  return fromExecution(latest)
}

export function connectTelemetrySocket(
  onTick: (snapshot: TelemetrySnapshot) => void,
  engineURL: string,
): () => void {
  let cursor = 0

  const interval = setInterval(() => {
    cursor += 1
    void fetchLiveSnapshot(engineURL)
      .then((live) => {
        if (live) {
          onTick({ ...live, mode: 'live' })
          return
        }

        onTick({
          throughput: 46800 + (cursor % 12) * 310,
          activeNodes: 7 + (cursor % 3),
          latencyMs: 12 + (cursor % 8),
          logs: [nextLog(cursor)],
          messages: [nextEdge(cursor), nextEdge(cursor + 1)],
          mode: 'mock',
          activeWorkflow: 'hello-world',
        })
      })
      .catch(() => {
        onTick({
          throughput: 46800 + (cursor % 12) * 310,
          activeNodes: 7 + (cursor % 3),
          latencyMs: 12 + (cursor % 8),
          logs: [nextLog(cursor)],
          messages: [nextEdge(cursor), nextEdge(cursor + 1)],
          mode: 'mock',
          activeWorkflow: 'hello-world',
        })
      })
  }, 1500)

  return () => clearInterval(interval)
}

export function useTelemetryFeed() {
  const engineURL = process.env.NEXT_PUBLIC_ENGINE_URL || 'http://127.0.0.1:8080'
  const [data, setData] = useState<TelemetrySnapshot>({
    throughput: 47000,
    activeNodes: 8,
    latencyMs: 14,
    logs: [
      {
        id: 'boot-log-1',
        nodeId: 'orchestrator',
        level: 'info',
        message: 'telemetry stream initialized',
        at: Date.now(),
      },
    ],
    messages: [
      {
        id: 'boot-edge-1',
        from: 'trigger',
        to: 'orchestrator',
        at: Date.now(),
      },
      {
        id: 'boot-edge-2',
        from: 'orchestrator',
        to: 'nats',
        at: Date.now(),
      },
    ],
    mode: 'mock',
    activeWorkflow: 'hello-world',
  })

  useEffect(() => {
    const unsubscribe = connectTelemetrySocket((snapshot) => {
      setData((prev) => ({
        throughput: snapshot.throughput,
        activeNodes: snapshot.activeNodes,
        latencyMs: snapshot.latencyMs,
        logs: [...snapshot.logs, ...prev.logs].slice(0, 16),
        messages: [...snapshot.messages, ...prev.messages].slice(0, 12),
        mode: snapshot.mode,
        activeWorkflow: snapshot.activeWorkflow,
      }))
    }, engineURL)

    return unsubscribe
  }, [engineURL])

  return data
}
