'use client'

import { WorkflowCanvas } from '@/react-flow'
import { useTelemetryFeed } from '@/telemetry-socket'

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

  return (
    <main className="page-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">Nebula-Stream</p>
          <h1>Live Workflow Studio</h1>
        </div>
        <span className="chip">{telemetry.mode === 'live' ? 'Live Engine Data' : 'Mock Telemetry'}</span>
      </header>

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
    </main>
  )
}
