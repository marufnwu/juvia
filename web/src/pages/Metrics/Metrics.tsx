import { useEffect, useState } from 'react'
import { WSClient } from '../../lib/ws'
import api from '../../lib/api'

interface MetricsSnapshot {
  cpu: number
  ram_used: number
  ram_total: number
  disk_used: number
  disk_total: number
  load_avg: number[]
  network_rx: number
  network_tx: number
  processes: number
  uptime: number
}

interface Process {
  pid: number
  name: string
  cpu: number
  ram: number
  status: string
  command: string
}

function formatBytes(bytes: number) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function formatUptime(seconds: number) {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}

export default function Metrics() {
  const [metrics, setMetrics] = useState<MetricsSnapshot | null>(null)
  const [processes, setProcesses] = useState<Process[]>([])
  const [history, setHistory] = useState<{ cpu: number; ram: number; time: string }[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchCurrentMetrics()

    const ws = new WSClient(`ws://${window.location.host}/ws/v1/metrics`)
    ws.on('metrics', (data) => {
      const m = data as MetricsSnapshot
      setMetrics(m)
      setLoading(false)
      setHistory((prev) => {
        const newPoint = {
          cpu: m.cpu,
          ram: m.ram_total > 0 ? (m.ram_used / m.ram_total) * 100 : 0,
          time: new Date().toLocaleTimeString(),
        }
        const updated = [...prev, newPoint].slice(-30)
        return updated
      })
    })
    ws.connect()

    return () => ws.close()
  }, [])

  const fetchCurrentMetrics = async () => {
    try {
      const res = await api.get('/metrics/current')
      const data = res.data
      if (data.success) {
        setMetrics(data.data)
        if (data.data.processes) {
          setProcesses(data.data.processes.slice(0, 20))
        }
      }
    } catch (err) {
      console.error('Failed to fetch metrics:', err)
    } finally {
      setLoading(false)
    }
  }

  const cpuPercent = metrics ? metrics.cpu.toFixed(1) : '0'
  const ramPercent = metrics && metrics.ram_total > 0
    ? ((metrics.ram_used / metrics.ram_total) * 100).toFixed(1)
    : '0'
  const diskPercent = metrics && metrics.disk_total > 0
    ? ((metrics.disk_used / metrics.disk_total) * 100).toFixed(1)
    : '0'

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Metrics</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Server performance and resource usage
        </p>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
        <MetricCard
          label="CPU Usage"
          value={metrics ? `${metrics.cpu.toFixed(1)}%` : '—'}
          sub={metrics ? `Load: ${metrics.load_avg.map((l) => l.toFixed(2)).join(', ')}` : ''}
        />
        <MetricCard
          label="Memory"
          value={metrics ? `${formatBytes(metrics.ram_used)}` : '—'}
          sub={metrics ? `of ${formatBytes(metrics.ram_total)}` : ''}
        />
        <MetricCard
          label="Disk"
          value={metrics ? `${formatBytes(metrics.disk_used)}` : '—'}
          sub={metrics ? `of ${formatBytes(metrics.disk_total)}` : ''}
        />
        <MetricCard
          label="Network RX"
          value={metrics ? `${formatBytes(metrics.network_rx)}` : '—'}
        />
        <MetricCard
          label="Network TX"
          value={metrics ? `${formatBytes(metrics.network_tx)}` : '—'}
        />
        <MetricCard
          label="Uptime"
          value={metrics ? formatUptime(metrics.uptime) : '—'}
        />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <UsageBar label="CPU" percent={parseFloat(cpuPercent)} color="bg-blue-500" />
        <UsageBar label="Memory" percent={parseFloat(ramPercent)} color="bg-green-500" />
        <UsageBar label="Disk" percent={parseFloat(diskPercent)} color="bg-orange-500" />
      </div>

      <div className="border rounded-lg p-4">
        <h2 className="text-sm font-medium mb-3">CPU & Memory (last 30 samples)</h2>
        {history.length > 1 ? (
          <div className="h-32 flex items-end gap-1">
            {history.map((point, i) => (
              <div key={i} className="flex-1 flex flex-col justify-end gap-px">
                <div
                  className="bg-blue-500 rounded-t"
                  style={{ height: `${point.cpu}%` }}
                />
                <div
                  className="bg-green-500 rounded-t"
                  style={{ height: `${point.ram}%` }}
                />
              </div>
            ))}
          </div>
        ) : (
          <div className="h-32 flex items-center justify-center text-muted-foreground text-sm">
            Collecting data...
          </div>
        )}
        <div className="flex gap-4 mt-2 text-xs text-muted-foreground">
          <span className="flex items-center gap-1">
            <span className="w-2 h-2 bg-blue-500 rounded-full" /> CPU
          </span>
          <span className="flex items-center gap-1">
            <span className="w-2 h-2 bg-green-500 rounded-full" /> Memory
          </span>
        </div>
      </div>

      <div className="border rounded-lg">
        <div className="p-4 border-b">
          <h2 className="text-sm font-medium">Top Processes</h2>
        </div>
        {loading ? (
          <div className="p-8 text-center text-muted-foreground">Loading...</div>
        ) : processes.length === 0 ? (
          <div className="p-8 text-center text-muted-foreground">No process data available</div>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="text-left p-3 text-xs font-medium">PID</th>
                <th className="text-left p-3 text-xs font-medium">Name</th>
                <th className="text-left p-3 text-xs font-medium">CPU</th>
                <th className="text-left p-3 text-xs font-medium">Memory</th>
                <th className="text-left p-3 text-xs font-medium">Status</th>
              </tr>
            </thead>
            <tbody>
              {processes.map((proc) => (
                <tr key={proc.pid} className="border-b hover:bg-muted/30">
                  <td className="p-3 text-sm font-mono">{proc.pid}</td>
                  <td className="p-3 text-sm">{proc.name}</td>
                  <td className="p-3 text-sm">{proc.cpu.toFixed(1)}%</td>
                  <td className="p-3 text-sm">{formatBytes(proc.ram)}</td>
                  <td className="p-3 text-sm">
                    <span className={`text-xs px-2 py-0.5 rounded ${
                      proc.status === 'running' ? 'bg-green-500/10 text-green-500' : 'bg-yellow-500/10 text-yellow-500'
                    }`}>
                      {proc.status}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}

function MetricCard({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <div className="border border-border bg-card p-4">
      <div className="text-xs uppercase text-muted-foreground font-medium mb-1">
        {label}
      </div>
      <div className="text-xl font-semibold">{value}</div>
      {sub && <div className="text-xs text-muted-foreground mt-1">{sub}</div>}
    </div>
  )
}

function UsageBar({ label, percent, color }: { label: string; percent: number; color: string }) {
  return (
    <div className="border border-border bg-card p-4">
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs font-medium">{label}</span>
        <span className="text-sm font-mono">{percent.toFixed(1)}%</span>
      </div>
      <div className="h-2 bg-muted rounded-full overflow-hidden">
        <div
          className={`h-full ${color} transition-all duration-300`}
          style={{ width: `${Math.min(percent, 100)}%` }}
        />
      </div>
    </div>
  )
}