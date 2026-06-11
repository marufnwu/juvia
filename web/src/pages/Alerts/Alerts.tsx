import { useEffect, useState } from 'react'
import { Bell, CheckCircle, AlertTriangle, XCircle } from 'lucide-react'
import api from '../../lib/api'

interface Alert {
  id: number
  type: string
  severity: string
  message: string
  resource_id: number
  resource_type: string
  acknowledged: boolean
  created_at: string
}

export default function Alerts() {
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState<'all' | 'unacknowledged'>('all')

  useEffect(() => {
    loadAlerts()
  }, [filter])

  const loadAlerts = async () => {
    setLoading(true)
    try {
      const params = filter === 'unacknowledged' ? '?acknowledged=false' : ''
      const res = await api.get(`/alerts${params}`)
      setAlerts(res.data.data || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const acknowledgeAlert = async (id: number) => {
    try {
      await api.post(`/alerts/${id}/acknowledge`)
      loadAlerts()
    } catch (err) {
      console.error(err)
    }
  }

  const deleteAlert = async (id: number) => {
    try {
      await api.delete(`/alerts/${id}`)
      loadAlerts()
    } catch (err) {
      console.error(err)
    }
  }

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'critical':
        return <XCircle size={16} className="text-red-600" />
      case 'warning':
        return <AlertTriangle size={16} className="text-yellow-600" />
      default:
        return <Bell size={16} className="text-blue-600" />
    }
  }

  const getSeverityClass = (severity: string): string => {
    switch (severity) {
      case 'critical':
        return 'border-l-red-500 bg-red-50'
      case 'warning':
        return 'border-l-yellow-500 bg-yellow-50'
      default:
        return 'border-l-blue-500 bg-blue-50'
    }
  }

  const getAlertTypeLabel = (type: string): string => {
    const labels: Record<string, string> = {
      'disk_usage': 'Disk Usage',
      'website_down': 'Website Down',
      'ssl_expiry': 'SSL Expiring',
      'backup_overdue': 'Backup Overdue',
      'service_down': 'Service Down',
    }
    const prefix = type.split(':')[0]
    return labels[prefix] || type
  }

  const formatDescription = (alert: Alert): string => {
    switch (alert.type) {
      case 'disk_usage':
        return 'Disk Usage · Your server is running low on disk space'
      case 'website_down':
        return 'Website Down · Your website is not responding'
      case 'ssl_expiry':
        return 'SSL Expiring · Your SSL certificate is about to expire'
      case 'backup_overdue':
        return 'Backup Overdue · No backup has been made recently'
      case 'service_down':
        return 'Service Down · A system service is not running'
      default:
        return alert.message
    }
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Alerts</h1>
          <p className="text-sm text-muted-foreground">
            Alert · A notification about something that needs your attention
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value as 'all' | 'unacknowledged')}
            className="border border-border px-3 py-2 text-sm"
          >
            <option value="all">All Alerts</option>
            <option value="unacknowledged">Unacknowledged</option>
          </select>
          <button onClick={loadAlerts} className="border border-border px-3 py-2 text-sm hover:bg-muted">
            Refresh
          </button>
        </div>
      </div>

      {loading ? (
        <div className="text-muted-foreground">Loading alerts...</div>
      ) : alerts.length === 0 ? (
        <div className="text-center py-12 border border-border">
          <CheckCircle size={48} className="mx-auto text-green-500 mb-4" />
          <p className="text-muted-foreground">No alerts</p>
        </div>
      ) : (
        <div className="space-y-3">
          {alerts.map((alert) => (
            <div
              key={alert.id}
              className={`border border-l-4 ${getSeverityClass(alert.severity)} p-4 ${alert.acknowledged ? 'opacity-60' : ''}`}
            >
              <div className="flex items-start gap-3">
                {getSeverityIcon(alert.severity)}
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="font-medium text-sm">{getAlertTypeLabel(alert.type)}</span>
                    <span className={`text-xs px-2 py-0.5 rounded ${alert.severity === 'critical' ? 'bg-red-200 text-red-800' : 'bg-yellow-200 text-yellow-800'}`}>
                      {alert.severity}
                    </span>
                    {alert.acknowledged && (
                      <span className="text-xs text-muted-foreground">Acknowledged</span>
                    )}
                  </div>
                  <p className="text-sm">{formatDescription(alert)}</p>
                  <p className="text-xs text-muted-foreground mt-1">
                    {new Date(alert.created_at).toLocaleString()}
                  </p>
                </div>
                <div className="flex gap-2">
                  {!alert.acknowledged && (
                    <button
                      onClick={() => acknowledgeAlert(alert.id)}
                      className="text-primary hover:underline text-xs"
                    >
                      Acknowledge
                    </button>
                  )}
                  <button onClick={() => deleteAlert(alert.id)} className="text-red-600 hover:underline text-xs">
                    Delete
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}