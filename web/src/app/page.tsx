'use client'

import { useEffect, useState } from 'react'
import { WorkflowCanvas } from '@/react-flow'
import {
  fetchExecutionByID,
  setActiveWorkflow,
  triggerWorkflow,
  useTelemetryFeed,
  type ExecutionDetail,
} from '@/telemetry-socket'

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
  const [selectedWorkflow, setSelectedWorkflow] = useState('hello-world')
  const [triggerStatus, setTriggerStatus] = useState('')
  const [activeStatus, setActiveStatus] = useState('')
  const [executionDetail, setExecutionDetail] = useState<ExecutionDetail | null>(null)
  const [detailStatus, setDetailStatus] = useState('')

  useEffect(() => {
    if (!telemetry.workflows.length) {
      return
    }

    if (!telemetry.workflows.includes(selectedWorkflow)) {
      setSelectedWorkflow(telemetry.activeWorkflow || telemetry.workflows[0])
    }
  }, [telemetry.activeWorkflow, telemetry.workflows, selectedWorkflow])

  const handleTrigger = async () => {
    const engineURL = process.env.NEXT_PUBLIC_ENGINE_URL || 'http://127.0.0.1:8080'
    const workflow = selectedWorkflow || telemetry.activeWorkflow
    try {
      const result = await triggerWorkflow(engineURL, workflow, 'dashboard manual trigger')
      setTriggerStatus(`triggered event ${result.event_id}`)
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'unknown trigger error'
      setTriggerStatus(msg)
    }
  }

  const handleSetActive = async () => {
    const engineURL = process.env.NEXT_PUBLIC_ENGINE_URL || 'http://127.0.0.1:8080'
    const workflow = selectedWorkflow || telemetry.activeWorkflow
    try {
      const result = await setActiveWorkflow(engineURL, workflow)
      setActiveStatus(`active workflow set to ${result.workflow}`)
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'unknown active workflow error'
      setActiveStatus(msg)
    }
  }

  const handleHistoryClick = async (eventID: string) => {
    const engineURL = process.env.NEXT_PUBLIC_ENGINE_URL || 'http://127.0.0.1:8080'
    setDetailStatus('loading execution detail...')
    try {
      const detail = await fetchExecutionByID(engineURL, eventID)
      setExecutionDetail(detail)
      setDetailStatus('')
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'unknown detail error'
      setDetailStatus(msg)
    }
  }

  const handleCopyDetail = async () => {
    if (!executionDetail) {
      return
    }
    try {
      await navigator.clipboard.writeText(JSON.stringify(executionDetail, null, 2))
      setDetailStatus('execution JSON copied to clipboard')
    } catch {
      setDetailStatus('clipboard write failed')
    }
  }

  const handleRetrigger = async () => {
    if (!executionDetail) {
      return
    }

    const engineURL = process.env.NEXT_PUBLIC_ENGINE_URL || 'http://127.0.0.1:8080'
    try {
      const message =
        typeof executionDetail.payload.message === 'string'
          ? executionDetail.payload.message
          : `retrigger from ${executionDetail.eventId}`
      const result = await triggerWorkflow(engineURL, executionDetail.workflow, message)
      setDetailStatus(`re-triggered as ${result.event_id}`)
      setTriggerStatus(`re-triggered event ${result.event_id}`)
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'unknown re-trigger error'
      setDetailStatus(msg)
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
          <label className="workflow-select-label">
            Workflow
            <select
              className="workflow-select"
              value={selectedWorkflow}
              onChange={(e) => setSelectedWorkflow(e.target.value)}
            >
              {telemetry.workflows.map((workflow) => (
                <option key={workflow} value={workflow}>
                  {workflow}
                </option>
              ))}
            </select>
          </label>
          <button onClick={handleTrigger} type="button" className="trigger-btn">
            Trigger {selectedWorkflow}
          </button>
          <button onClick={handleSetActive} type="button" className="trigger-btn secondary-btn">
            Set Active
          </button>
          <span>{triggerStatus || 'waiting'}</span>
        </div>
        <div className="trigger-status-line">{activeStatus || `current active: ${telemetry.activeWorkflow}`}</div>
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
              <li key={item.eventId} onClick={() => handleHistoryClick(item.eventId)} className="history-item">
                <strong>{item.workflow}</strong>
                <span>{item.eventId}</span>
                <p>{item.durationMs} ms</p>
              </li>
            ))
          )}
        </ul>
        <div className="detail-panel">
          {executionDetail ? (
            <>
              <h3>Execution Detail</h3>
              <p>
                {executionDetail.workflow} · {executionDetail.eventId} · {executionDetail.durationMs} ms
              </p>
              <div className="detail-actions">
                <button type="button" className="trigger-btn" onClick={handleCopyDetail}>
                  Copy JSON
                </button>
                <button type="button" className="trigger-btn secondary-btn" onClick={handleRetrigger}>
                  Re-trigger Payload
                </button>
              </div>
              <ul className="step-detail-list">
                {executionDetail.results.map((step) => (
                  <li key={step.id}>
                    <strong>{step.id}</strong>
                    <span>{step.type}</span>
                    <pre>{JSON.stringify(step.output ?? {}, null, 2)}</pre>
                  </li>
                ))}
              </ul>
              <p>{detailStatus || 'ready'}</p>
            </>
          ) : (
            <p>{detailStatus || 'click an execution card to inspect step outputs'}</p>
          )}
        </div>
      </section>
    </main>
  )
}
