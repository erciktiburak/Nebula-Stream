'use client'

import { useState } from 'react'
import { WorkflowCanvas } from '@/react-flow'
import { triggerWorkflow, useTelemetryFeed } from '@/telemetry-socket'

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <article className="stat-card">
      <p>{label}</p>
      <h3>{value}</h3>
    </article>
  )
}

export default function Page() {
  const telemetry = useTelemetryFeed()
  const [triggerStatus, setTriggerStatus] = useState('')

  const handleTrigger = async () => {
    const engineURL = process.env.NEXT_PUBLIC_ENGINE_URL || 'http://127.0.0.1:8080'
    try {
      const result = await triggerWorkflow(engineURL, telemetry.activeWorkflow, 'dashboard manual trigger')
      setTriggerStatus(`triggered event ${result.event_id}`)
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'unknown trigger error'
      setTriggerStatus(msg)
    }
  }

  return (
    <main className="page-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">Nebula-Stream</p>
          <h1>Live Workflow Studio</h1>
        </div>
        <span className="chip">{telemetry.mode === 'live' ? 'Live Engine Data' : 'Mock Telemetry'}</span>
      </header>

      <section className="panel trigger-panel">
        <div className="panel-head">
          <h2>Manual Trigger</h2>
          <p>Send one event to the active workflow via control-plane API</p>
        </div>
        <div className="trigger-actions">
          <button onClick={handleTrigger} type="button" className="trigger-btn">
            Trigger {telemetry.activeWorkflow}
          </button>
          <span>{triggerStatus || 'waiting'}</span>
        </div>
      </section>

      <section className="stat-grid">
        <StatCard label="Event Throughput" value={`${telemetry.throughput}/sec`} />
        <StatCard label="Active Nodes" value={`${telemetry.activeNodes}`} />
        <StatCard label="Pipeline Latency" value={`${telemetry.latencyMs} ms`} />
        <StatCard label="Workflow" value={telemetry.activeWorkflow} />
      </section>

      <section className="content-grid">
        <article className="panel graph-panel">
          <div className="panel-head">
            <h2>Workflow Graph</h2>
            <p>Real-time edge highlights from streaming events</p>
          </div>
          <WorkflowCanvas messages={telemetry.messages} />
        </article>

        <article className="panel log-panel">
          <div className="panel-head">
            <h2>Node Logs</h2>
            <p>Last 8 telemetry records</p>
          </div>
          <ul>
            {telemetry.logs.slice(0, 8).map((log) => (
              <li key={log.id}>
                <span>{log.nodeId}</span>
                <strong className={`level-${log.level}`}>{log.level}</strong>
                <p>{log.message}</p>
              </li>
            ))}
          </ul>
        </article>
      </section>

      <section className="panel history-panel">
        <div className="panel-head">
          <h2>Execution History</h2>
          <p>Recent runs from control-plane API</p>
        </div>
        <ul className="history-list">
          {telemetry.history.length === 0 ? (
            <li>no execution history yet</li>
          ) : (
            telemetry.history.map((item) => (
              <li key={item.eventId}>
                <strong>{item.workflow}</strong>
                <span>{item.eventId}</span>
                <p>{item.durationMs} ms</p>
              </li>
            ))
          )}
        </ul>
      </section>
    </main>
  )
}
