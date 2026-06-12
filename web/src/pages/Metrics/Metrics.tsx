import { useEffect, useState } from 'react'
import { Activity, Cpu, HardDrive, MemoryStick, Network, Clock, RefreshCw } from 'lucide-react'
import { Card, CardHeader, CardTitle } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { PageHeader } from '../../components/ui/Misc'
import { ProgressBar } from '../../components/ui/Misc'
import { WSClient } from '../../lib/ws'
import { formatBytes, formatPercentage } from '../../lib/utils'
import { LineChart, Line, AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'
import { cn } from '../../lib/utils'

interface MetricsSnapshot {
  cpu: number
  ram_used: number
  ram_total: number
  disk_used: number
  disk_total: number
  load_avg: number[]
  processes: number
  uptime: number
  network_in: number
  network_out: number
}

interface DataPoint {
  time: string
  value: number
}

export default function Metrics() {
  const [metrics, setMetrics] = useState<MetricsSnapshot | null>(null)
  const [cpuHistory, setCpuHistory] = useState<DataPoint[]>([])
  const [ramHistory, setRamHistory] = useState<DataPoint[]>([])
  const [timeRange, setTimeRange] = useState<'1h' | '6h' | '24h' | '7d'>('1h')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const ws = new WSClient(`ws://${window.location.host}/ws/v1/metrics`)
    ws.on('metrics', (data) => {
      const m = data as MetricsSnapshot
      setMetrics(m)
      setLoading(false)

      const time = new Date().toLocaleTimeString()
      setCpuHistory((prev) => [...prev.slice(-59), { time, value: m.cpu }])
      setRamHistory((prev) => [...prev.slice(-59), {
        time,
        value: Math.round((m.ram_used / m.ram_total) * 100),
      }])
    })
    ws.connect()
    return () => ws.close()
  }, [])

  const cpuPercent = metrics?.cpu ?? 0
  const ramUsed = metrics?.ram_used ?? 0
  const ramTotal = metrics?.ram_total ?? 1
  const diskUsed = metrics?.disk_used ?? 0
  const diskTotal = metrics?.disk_total ?? 1

  const getVariant = (percent: number) => {
    if (percent >= 90) return 'danger'
    if (percent >= 70) return 'warning'
    return 'primary'
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Server Metrics"
        description="Real-time server performance monitoring"
        breadcrumbs={[{ label: 'Metrics' }]}
        actions={
          <div className="flex items-center gap-2">
            {['1h', '6h', '24h', '7d'].map((range) => (
              <button
                key={range}
                onClick={() => setTimeRange(range as any)}
                className={cn(
                  'px-3 py-1.5 text-xs font-medium rounded transition-colors',
                  timeRange === range
                    ? 'bg-primary text-white'
                    : 'bg-accent text-text-secondary hover:text-foreground'
                )}
              >
                {range}
              </button>
            ))}
            <Button variant="outline" size="sm" onClick={() => {}}>
              <RefreshCw className="w-4 h-4 mr-1" />
              Refresh
            </Button>
          </div>
        }
      />

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="relative overflow-hidden">
          <div className="flex items-start justify-between mb-2">
            <div className="p-2 bg-accent/50 rounded">
              <Cpu className="w-4 h-4 text-text-secondary" />
            </div>
            <span className={cn('text-xs font-medium', getVariant(cpuPercent) === 'danger' ? 'text-danger' : getVariant(cpuPercent) === 'warning' ? 'text-warning' : 'text-success')}>
              {getVariant(cpuPercent) === 'danger' ? 'Critical' : getVariant(cpuPercent) === 'warning' ? 'Warning' : 'Normal'}
            </span>
          </div>
          <p className="text-2xl font-semibold">{formatPercentage(cpuPercent)}</p>
          <p className="text-xs text-text-secondary">CPU Usage</p>
          <div className="mt-2">
            <ProgressBar value={cpuPercent} variant={getVariant(cpuPercent)} showLabel />
          </div>
        </Card>

        <Card>
          <div className="flex items-start justify-between mb-2">
            <div className="p-2 bg-accent/50 rounded">
              <MemoryStick className="w-4 h-4 text-text-secondary" />
            </div>
          </div>
          <p className="text-2xl font-semibold">{formatBytes(ramUsed)}</p>
          <p className="text-xs text-text-secondary">of {formatBytes(ramTotal)}</p>
          <div className="mt-2">
            <ProgressBar value={(ramUsed / ramTotal) * 100} variant={getVariant((ramUsed / ramTotal) * 100)} showLabel />
          </div>
        </Card>

        <Card>
          <div className="flex items-start justify-between mb-2">
            <div className="p-2 bg-accent/50 rounded">
              <HardDrive className="w-4 h-4 text-text-secondary" />
            </div>
          </div>
          <p className="text-2xl font-semibold">{formatBytes(diskUsed)}</p>
          <p className="text-xs text-text-secondary">of {formatBytes(diskTotal)}</p>
          <div className="mt-2">
            <ProgressBar value={(diskUsed / diskTotal) * 100} variant={getVariant((diskUsed / diskTotal) * 100)} showLabel />
          </div>
        </Card>

        <Card>
          <div className="flex items-start justify-between mb-2">
            <div className="p-2 bg-accent/50 rounded">
              <Activity className="w-4 h-4 text-text-secondary" />
            </div>
          </div>
          <p className="text-2xl font-semibold">{metrics?.processes ?? '—'}</p>
          <p className="text-xs text-text-secondary">Processes</p>
          {metrics?.load_avg && (
            <p className="text-xs text-text-secondary mt-1">
              Load: {metrics.load_avg.join(', ')}
            </p>
          )}
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>CPU Usage Over Time</CardTitle>
          </CardHeader>
          <div className="h-48">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={cpuHistory}>
                <defs>
                  <linearGradient id="cpuGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#2563EB" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#2563EB" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <XAxis dataKey="time" tick={{ fontSize: 10, fill: '#94A3B8' }} />
                <YAxis domain={[0, 100]} tick={{ fontSize: 10, fill: '#94A3B8' }} tickFormatter={(v) => `${v}%`} />
                <Tooltip
                  contentStyle={{ background: '#1E293B', border: '1px solid #334155', borderRadius: '6px', fontSize: '12px' }}
                  labelStyle={{ color: '#F8FAFC' }}
                />
                <Area type="monotone" dataKey="value" stroke="#2563EB" fill="url(#cpuGradient)" strokeWidth={2} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Memory Usage Over Time</CardTitle>
          </CardHeader>
          <div className="h-48">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={ramHistory}>
                <defs>
                  <linearGradient id="ramGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#16A34A" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#16A34A" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <XAxis dataKey="time" tick={{ fontSize: 10, fill: '#94A3B8' }} />
                <YAxis domain={[0, 100]} tick={{ fontSize: 10, fill: '#94A3B8' }} tickFormatter={(v) => `${v}%`} />
                <Tooltip
                  contentStyle={{ background: '#1E293B', border: '1px solid #334155', borderRadius: '6px', fontSize: '12px' }}
                  labelStyle={{ color: '#F8FAFC' }}
                />
                <Area type="monotone" dataKey="value" stroke="#16A34A" fill="url(#ramGradient)" strokeWidth={2} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Network Traffic</CardTitle>
        </CardHeader>
        <div className="grid grid-cols-2 gap-6">
          <div className="flex items-center gap-3">
            <div className="p-3 bg-primary/10 rounded">
              <Network className="w-5 h-5 text-primary" />
            </div>
            <div>
              <p className="text-sm text-text-secondary">Inbound</p>
              <p className="text-xl font-semibold">{metrics?.network_in ? formatBytes(metrics.network_in) + '/s' : '—'}</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <div className="p-3 bg-success/10 rounded">
              <Network className="w-5 h-5 text-success rotate-180" />
            </div>
            <div>
              <p className="text-sm text-text-secondary">Outbound</p>
              <p className="text-xl font-semibold">{metrics?.network_out ? formatBytes(metrics.network_out) + '/s' : '—'}</p>
            </div>
          </div>
        </div>
      </Card>
    </div>
  )
}