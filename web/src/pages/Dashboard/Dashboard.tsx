import { useEffect, useState, useRef } from 'react'
import { Link } from 'react-router-dom'
import {
  Activity, Cpu, HardDrive, MemoryStick, Globe, Database, Mail,
  Server, Clock, AlertTriangle, CheckCircle, XCircle, ArrowRight,
  Plus, Terminal, Archive, Shield, RefreshCw, Copy
} from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Badge, StatusBadge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { ProgressBar } from '../../components/ui/Misc'
import { PageHeader } from '../../components/ui/Misc'
import { ConfirmModal } from '../../components/ui/Modal'
import api from '../../lib/api'
import { formatBytes, formatPercentage, timeAgo } from '../../lib/utils'
import { cn } from '../../lib/utils'
import { WSClient } from '../../lib/ws'
import { useApiError } from '../../hooks/useToast'

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

interface Service {
  name: string
  description: string
  status: string
  installed: boolean
  error_detail?: string
  fix_hint?: string
}

interface Alert {
  id: number
  type: string
  message: string
  severity: string
  created_at: string
}

interface Website {
  id: number
  domain: string
  status: string
  ssl_enabled: boolean
}

export default function Dashboard() {
  const showError = useApiError()
  const [metrics, setMetrics] = useState<MetricsSnapshot | null>(null)
  const [services, setServices] = useState<Service[]>([])
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [websites, setWebsites] = useState<Website[]>([])
  const [loading, setLoading] = useState(true)
  const [serverIP, setServerIP] = useState<string>('')
  const [nameservers, setNameservers] = useState<string[]>([])
  const [restarting, setRestarting] = useState<string | null>(null)
  const [confirmRestart, setConfirmRestart] = useState<string | null>(null)
  const loadDataRef = useRef<() => void>(() => {})

  useEffect(() => {
    const ws = new WSClient(`ws://${window.location.host}/ws/v1/metrics`)
    ws.on('metrics', (data) => {
      setMetrics(data as MetricsSnapshot)
    })
    ws.connect()

    const loadData = async () => {
      try {
        const [metricsRes, servicesRes, alertsRes, websitesRes, ipRes, nsRes] = await Promise.allSettled([
          api.get('/metrics/current'),
          api.get('/services'),
          api.get('/alerts'),
          api.get('/websites'),
          api.get('/server/ip'),
          api.get('/dns/nameservers'),
        ])

        if (metricsRes.status === 'fulfilled') {
          setMetrics(metricsRes.value.data.data)
        }
        if (servicesRes.status === 'fulfilled') {
          setServices(servicesRes.value.data.data || [])
        }
        if (alertsRes.status === 'fulfilled') {
          setAlerts(alertsRes.value.data.data?.slice(0, 5) || [])
        }
        if (websitesRes.status === 'fulfilled') {
          setWebsites(websitesRes.value.data.data?.slice(0, 5) || [])
        }
        if (ipRes.status === 'fulfilled') {
          setServerIP(ipRes.value.data.data?.ip || '')
        }
        if (nsRes.status === 'fulfilled') {
          const nsData = nsRes.value.data.data
          setNameservers([nsData?.ns1_hostname, nsData?.ns2_hostname].filter(Boolean))
        }
      } catch (err) {
        console.error('Failed to load dashboard data:', err)
      } finally {
        setLoading(false)
      }
    }

    loadDataRef.current = loadData
    loadData()
    const interval = setInterval(loadData, 30000)
    return () => {
      ws.close()
      clearInterval(interval)
    }
  }, [])

  const handleRestart = async (serviceName: string) => {
    setConfirmRestart(serviceName)
  }

  const doRestart = async () => {
    if (!confirmRestart) return
    const serviceName = confirmRestart
    setConfirmRestart(null)
    setRestarting(serviceName)
    try {
      await api.post('/services/' + serviceName + '/restart')
      setTimeout(() => loadDataRef.current(), 2000)
    } catch (err: any) {
      showError(err, 'Failed to restart ' + serviceName)
    } finally {
      setRestarting(null)
    }
  }

  const cpuPercent = metrics?.cpu ?? 0
  const ramUsed = metrics?.ram_used ?? 0
  const ramTotal = metrics?.ram_total ?? 1
  const diskUsed = metrics?.disk_used ?? 0
  const diskTotal = metrics?.disk_total ?? 1

  const getMetricVariant = (percent: number) => {
    if (percent >= 90) return 'danger'
    if (percent >= 70) return 'warning'
    return 'primary'
  }

  const getServiceStatus = (status: string) => {
    if (status === 'running' || status === 'active') return 'healthy'
    if (status === 'stopped') return 'critical'
    return 'unknown'
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Dashboard"
        description="Server overview and quick actions"
        breadcrumbs={[{ label: 'Dashboard' }]}
        actions={
          <div className="flex items-center gap-2">
            <Link
              to="/websites/create"
              className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-white text-sm font-medium rounded hover:bg-primary/90 transition-colors"
            >
              <Plus className="w-4 h-4" />
              Create Website
            </Link>
            <Link
              to="/terminal"
              className="inline-flex items-center gap-2 px-4 py-2 border border-border text-foreground text-sm font-medium rounded hover:bg-accent transition-colors"
            >
              <Terminal className="w-4 h-4" />
              Terminal
            </Link>
          </div>
        }
      />

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard
          title="CPU Usage · How busy your server's processor is"
          icon={<Cpu className="w-4 h-4" />}
          value={formatPercentage(cpuPercent)}
          subValue={metrics ? `Load: ${metrics.load_avg?.join(', ') || 'N/A'}` : undefined}
          variant={getMetricVariant(cpuPercent)}
          loading={loading}
        />
        <MetricCard
          title="Memory (RAM) · Temporary working space your server uses to run programs"
          icon={<MemoryStick className="w-4 h-4" />}
          value={formatBytes(ramUsed)}
          subValue={metrics ? `of ${formatBytes(ramTotal)}` : undefined}
          variant={getMetricVariant((ramUsed / ramTotal) * 100)}
          loading={loading}
        />
        <MetricCard
          title="Disk · Permanent storage space on your server"
          icon={<HardDrive className="w-4 h-4" />}
          value={formatBytes(diskUsed)}
          subValue={metrics ? `of ${formatBytes(diskTotal)}` : undefined}
          variant={getMetricVariant((diskUsed / diskTotal) * 100)}
          loading={loading}
        />
        <MetricCard
          title="Processes · The number of programs currently running"
          icon={<Activity className="w-4 h-4" />}
          value={metrics?.processes?.toString() ?? '—'}
          subValue={metrics?.uptime ? `Uptime · How long since your server last restarted: ${formatUptime(metrics.uptime)}` : undefined}
          variant="primary"
          loading={loading}
        />
      </div>

      {serverIP && (
        <Card className="bg-gradient-to-r from-primary/5 to-transparent border-primary/20">
          <div className="p-4">
            <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
              <div className="space-y-1">
                <h3 className="text-sm font-medium text-foreground flex items-center gap-2">
                  <Server className="w-4 h-4 text-primary" />
                  Server IP · Address to point your domains to
                </h3>
                <p className="text-2xl font-mono font-semibold text-foreground">{serverIP}</p>
                {nameservers.length > 0 && (
                  <p className="text-sm text-text-secondary">
                    Nameservers: <span className="font-mono">{nameservers.join(' / ')}</span>
                  </p>
                )}
              </div>
              <div className="flex items-center gap-2">
                <Button variant="outline" size="sm" onClick={() => navigator.clipboard.writeText(serverIP)}>
                  <Copy className="w-3.5 h-3.5 mr-1.5" />
                  Copy IP
                </Button>
                <Link to="/settings">
                  <Button variant="primary" size="sm">Configure Nameservers</Button>
                </Link>
              </div>
            </div>
          </div>
        </Card>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <Card padding="none">
            <div className="p-4 border-b border-border">
              <div className="flex items-center justify-between">
                <CardTitle>Services Status · Whether key programs are running</CardTitle>
                <Link to="/settings" className="text-xs text-primary hover:underline">View all</Link>
              </div>
            </div>
            <div className="divide-y divide-border">
              {loading ? (
                [...Array(6)].map((_, i) => (
                  <div key={i} className="flex items-center justify-between p-4">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded skeleton" />
                      <div className="space-y-1">
                        <div className="h-4 w-24 skeleton rounded" />
                        <div className="h-3 w-16 skeleton rounded" />
                      </div>
                    </div>
                    <div className="h-5 w-16 skeleton rounded" />
                  </div>
                ))
              ) : services.length === 0 ? (
                <div className="p-8 text-center text-text-secondary text-sm">
                  No services configured
                </div>
              ) : (
                services.map((service) => (
                  <div key={service.name} className="flex items-center justify-between p-4 hover:bg-accent/30 transition-colors">
                    <div className="flex items-center gap-3 min-w-0">
                      <div className={`w-8 h-8 rounded flex items-center justify-center ${
                        !service.installed ? 'bg-warning/10' :
                        service.status === 'failed' ? 'bg-danger/10' :
                        service.status === 'inactive' ? 'bg-accent/50' :
                        'bg-success/10'
                      }`}>
                        <Server className={`w-4 h-4 ${
                          !service.installed ? 'text-warning' :
                          service.status === 'failed' ? 'text-danger' :
                          service.status === 'inactive' ? 'text-text-secondary' :
                          'text-success'
                        }`} />
                      </div>
                      <div className="min-w-0">
                        <p className="text-sm font-medium">{service.description || service.name}</p>
                        <div className="flex items-center gap-2">
                          <p className="text-xs text-text-secondary capitalize">{service.name}</p>
                          {service.error_detail && (
                            <span className="text-xs text-warning/80 truncate">· {service.error_detail}</span>
                          )}
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-2 shrink-0">
                      <StatusBadge status={service.status} />
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleRestart(service.name)}
                        disabled={restarting === service.name || !service.installed}
                        loading={restarting === service.name}
                        title={!service.installed ? 'Install first' : 'Restart'}
                      >
                        <RefreshCw className="w-3.5 h-3.5" />
                      </Button>
                    </div>
                  </div>
                ))
              )}
            </div>
          </Card>

          <Card padding="none">
            <div className="p-4 border-b border-border">
              <div className="flex items-center justify-between">
                <CardTitle>Recent Websites</CardTitle>
                <Link to="/websites" className="text-xs text-primary hover:underline flex items-center gap-1">
                  View all <ArrowRight className="w-3 h-3" />
                </Link>
              </div>
            </div>
            <div className="divide-y divide-border">
              {loading ? (
                [...Array(3)].map((_, i) => (
                  <div key={i} className="flex items-center justify-between p-4">
                    <div className="space-y-1">
                      <div className="h-4 w-32 skeleton rounded" />
                      <div className="h-3 w-20 skeleton rounded" />
                    </div>
                    <div className="h-5 w-16 skeleton rounded" />
                  </div>
                ))
              ) : websites.length === 0 ? (
                <div className="p-8 text-center">
                  <Globe className="w-8 h-8 mx-auto text-text-secondary/50 mb-2" />
                  <p className="text-sm text-text-secondary mb-3">No websites yet</p>
                  <Link to="/websites/create" className="text-xs text-primary hover:underline">
                    Create your first website
                  </Link>
                </div>
              ) : (
                websites.map((site) => (
                  <Link
                    key={site.id}
                    to={`/websites/${site.id}`}
                    className="flex items-center justify-between p-4 hover:bg-accent/30 transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded bg-accent/50 flex items-center justify-center">
                        <Globe className="w-4 h-4 text-text-secondary" />
                      </div>
                      <div>
                        <p className="text-sm font-medium">{site.domain}</p>
                        <p className="text-xs text-text-secondary">
                          {site.ssl_enabled ? (
                            <span className="text-success">SSL · Active</span>
                          ) : (
                            <span>No SSL · Security not configured</span>
                          )}
                        </p>
                      </div>
                    </div>
                    <StatusBadge status={site.status} />
                  </Link>
                ))
              )}
            </div>
          </Card>
        </div>

        <div className="space-y-6">
          <Card padding="none">
            <div className="p-4 border-b border-border">
              <CardTitle className="flex items-center gap-2">
                <AlertTriangle className="w-4 h-4 text-warning" />
                Recent Alerts
              </CardTitle>
            </div>
            <div className="divide-y divide-border">
              {loading ? (
                [...Array(3)].map((_, i) => (
                  <div key={i} className="p-4 space-y-1">
                    <div className="h-4 w-full skeleton rounded" />
                    <div className="h-3 w-24 skeleton rounded" />
                  </div>
                ))
              ) : alerts.length === 0 ? (
                <div className="p-6 text-center">
                  <CheckCircle className="w-8 h-8 mx-auto text-success/50 mb-2" />
                  <p className="text-sm text-text-secondary">No active alerts</p>
                </div>
              ) : (
                alerts.map((alert) => (
                  <div key={alert.id} className="p-4">
                    <div className="flex items-start gap-2">
                      <span
                        className={`mt-0.5 w-1.5 h-1.5 rounded-full ${
                          alert.severity === 'critical'
                            ? 'bg-danger'
                            : alert.severity === 'warning'
                            ? 'bg-warning'
                            : 'bg-primary'
                        }`}
                      />
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium truncate">{alert.message}</p>
                        <p className="text-xs text-text-secondary mt-0.5">
                          {alert.type} · {timeAgo(alert.created_at)}
                        </p>
                      </div>
                    </div>
                  </div>
                ))
              )}
              {alerts.length > 0 && (
                <Link
                  to="/alerts"
                  className="block p-3 text-center text-xs text-primary hover:bg-accent/50 border-t border-border"
                >
                  View all alerts
                </Link>
              )}
            </div>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Quick Actions</CardTitle>
            </CardHeader>
            <div className="grid grid-cols-2 gap-2">
              <Link
                to="/websites/create"
                className="flex items-center gap-2 p-3 bg-accent/50 rounded hover:bg-accent transition-colors"
              >
                <Globe className="w-4 h-4 text-primary" />
                <span className="text-xs font-medium">Website</span>
              </Link>
              <Link
                to="/databases/create"
                className="flex items-center gap-2 p-3 bg-accent/50 rounded hover:bg-accent transition-colors"
              >
                <Database className="w-4 h-4 text-primary" />
                <span className="text-xs font-medium">Database · Structured storage for website data</span>
              </Link>
              <Link
                to="/email"
                className="flex items-center gap-2 p-3 bg-accent/50 rounded hover:bg-accent transition-colors"
              >
                <Mail className="w-4 h-4 text-primary" />
                <span className="text-xs font-medium">Mailbox · An email account</span>
              </Link>
              <Link
                to="/backups"
                className="flex items-center gap-2 p-3 bg-accent/50 rounded hover:bg-accent transition-colors"
              >
                <Archive className="w-4 h-4 text-primary" />
                <span className="text-xs font-medium">Backup</span>
              </Link>
              <Link
                to="/firewall"
                className="flex items-center gap-2 p-3 bg-accent/50 rounded hover:bg-accent transition-colors"
              >
                <Shield className="w-4 h-4 text-primary" />
                <span className="text-xs font-medium">Firewall · Traffic rules</span>
              </Link>
              <Link
                to="/cron"
                className="flex items-center gap-2 p-3 bg-accent/50 rounded hover:bg-accent transition-colors"
              >
                <Clock className="w-4 h-4 text-primary" />
                <span className="text-xs font-medium">Cron · Scheduled tasks</span>
              </Link>
            </div>
          </Card>
        </div>
      </div>

      <ConfirmModal
        open={!!confirmRestart}
        onClose={() => setConfirmRestart(null)}
        onConfirm={doRestart}
        title="Restart Service · A background program that keeps your server running"
        description={confirmRestart ? `Restart the ${confirmRestart} service? This may briefly interrupt dependent websites.` : ''}
        confirmLabel="Restart"
        variant="primary"
        loading={!!restarting}
      />
    </div>
  )
}

function MetricCard({
  title,
  icon,
  value,
  subValue,
  variant = 'primary',
  loading,
}: {
  title: string
  icon: React.ReactNode
  value: string
  subValue?: string
  variant?: 'primary' | 'warning' | 'danger'
  loading?: boolean
}) {
  const variantColors = {
    primary: 'text-primary',
    warning: 'text-warning',
    danger: 'text-danger',
  }

  return (
    <Card className="relative overflow-hidden">
      <div className="flex items-start justify-between mb-3">
        <div className="p-2 bg-accent/50 rounded">
          <span className="text-text-secondary">{icon}</span>
        </div>
      </div>
      <div className="space-y-1">
        {loading ? (
          <>
            <div className="h-8 w-20 skeleton rounded" />
            <div className="h-3 w-16 skeleton rounded" />
          </>
        ) : (
          <>
            <p className={`text-2xl font-semibold ${variantColors[variant]}`}>{value}</p>
            <p className="text-xs text-text-secondary">{title}</p>
            {subValue && <p className="text-xs text-text-secondary">{subValue}</p>}
          </>
        )}
      </div>
    </Card>
  )
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
}
