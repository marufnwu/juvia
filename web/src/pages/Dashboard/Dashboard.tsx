import { useEffect, useState } from 'react'
import { WSClient } from '../../lib/ws'

interface MetricsSnapshot {
  cpu: number
  ram_used: number
  ram_total: number
  disk_used: number
  disk_total: number
  load_avg: number[]
  processes: number
  uptime: number
}

function formatBytes(bytes: number) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

export default function Dashboard() {
  const [metrics, setMetrics] = useState<MetricsSnapshot | null>(null)

  useEffect(() => {
    const ws = new WSClient(`ws://${window.location.host}/ws/v1/metrics`)
    ws.on('metrics', (data) => {
      setMetrics(data as MetricsSnapshot)
    })
    ws.connect()
    return () => ws.close()
  }, [])

  return (
    <div>
      <h1 className="text-lg font-semibold mb-6">Dashboard</h1>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard
          label="CPU Usage"
          value={metrics ? `${metrics.cpu.toFixed(1)}%` : '—'}
        />
        <MetricCard
          label="Memory"
          value={
            metrics
              ? `${formatBytes(metrics.ram_used)} / ${formatBytes(metrics.ram_total)}`
              : '—'
          }
        />
        <MetricCard
          label="Disk"
          value={
            metrics
              ? `${formatBytes(metrics.disk_used)} / ${formatBytes(metrics.disk_total)}`
              : '—'
          }
        />
        <MetricCard
          label="Processes"
          value={metrics ? `${metrics.processes}` : '—'}
        />
      </div>
    </div>
  )
}

function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="border border-border bg-card p-4">
      <div className="text-xs uppercase text-muted-foreground font-medium mb-1">
        {label}
      </div>
      <div className="text-xl font-semibold">{value}</div>
    </div>
  )
}
