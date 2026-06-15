import { useEffect, useState } from 'react'
import { AlertTriangle, Bell, CheckCircle, Trash2, Settings, RefreshCw } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { Table } from '../../components/ui/Table'
import { PageHeader } from '../../components/ui/Misc'
import { Modal } from '../../components/ui/Modal'
import { Input, Label, Select, FormGroup, Switch } from '../../components/ui/Input'
import api from '../../lib/api'
import { formatDate, timeAgo } from '../../lib/utils'
import { cn } from '../../lib/utils'
import { useApiError } from '../../hooks/useToast'

interface Alert {
  id: number
  type: string
  message: string
  severity: 'critical' | 'warning' | 'info'
  source: string
  status: 'active' | 'acknowledged'
  created_at: string
}

export default function Alerts() {
  const showError = useApiError()
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [loading, setLoading] = useState(true)
  const [showSettingsModal, setShowSettingsModal] = useState(false)
  const [saving, setSaving] = useState(false)
  const [alertSettings, setAlertSettings] = useState({
    website_down_enabled: true,
    website_down_threshold: 5,
    ssl_expiry_enabled: true,
    ssl_expiry_days: 14,
    disk_usage_enabled: true,
    disk_usage_threshold: 80,
    backup_overdue_enabled: true,
    backup_overdue_days: 7,
    service_down_enabled: true,
  })

  useEffect(() => {
    loadAlerts()
    loadSettings()
  }, [])

  const loadAlerts = async () => {
    try {
      const res = await api.get('/alerts')
      setAlerts(res.data.data || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const loadSettings = async () => {
    try {
      const res = await api.get('/alerts/settings')
      const data = res.data.data || {}
      setAlertSettings({
        website_down_enabled: data.website_down_enabled ?? true,
        website_down_threshold: data.website_down_threshold ?? 5,
        ssl_expiry_enabled: data.ssl_expiry_enabled ?? true,
        ssl_expiry_days: data.ssl_expiry_days ?? 14,
        disk_usage_enabled: data.disk_usage_enabled ?? true,
        disk_usage_threshold: data.disk_usage_threshold ?? 80,
        backup_overdue_enabled: data.backup_overdue_enabled ?? true,
        backup_overdue_days: data.backup_overdue_days ?? 7,
        service_down_enabled: data.service_down_enabled ?? true,
      })
    } catch (err) {
      console.error(err)
    }
  }

  const handleAcknowledge = async (id: number) => {
    try {
      await api.post(`/alerts/${id}/acknowledge`)
      setAlerts(alerts.map((a) => a.id === id ? { ...a, status: 'acknowledged' } : a))
    } catch (err) {
      console.error(err)
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await api.delete(`/alerts/${id}`)
      setAlerts(alerts.filter((a) => a.id !== id))
    } catch (err) {
      console.error(err)
    }
  }

  const saveSettings = async () => {
    setSaving(true)
    try {
      await api.put('/alerts/settings', alertSettings)
      setShowSettingsModal(false)
    } catch (err: any) {
      showError(err, 'Failed to save alert settings')
    } finally {
      setSaving(false)
    }
  }

  const columns = [
    {
      key: 'severity',
      header: 'Severity',
      width: '100px',
      render: (alert: Alert) => (
        <span className={cn('w-2 h-2 rounded-full',
          alert.severity === 'critical' ? 'bg-danger' :
          alert.severity === 'warning' ? 'bg-warning' : 'bg-primary')} />
      ),
    },
    {
      key: 'message',
      header: 'Alert',
      render: (alert: Alert) => (
        <div>
          <p className="font-medium text-sm">{alert.message}</p>
          <p className="text-xs text-text-secondary mt-0.5">
            {alert.type} · {alert.source}
          </p>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (alert: Alert) => (
        <Badge variant={alert.status === 'active' ? 'danger' : 'neutral'}>
          {alert.status}
        </Badge>
      ),
    },
    {
      key: 'created',
      header: 'Time',
      render: (alert: Alert) => (
        <span className="text-xs text-text-secondary">{timeAgo(alert.created_at)}</span>
      ),
    },
    {
      key: 'actions',
      header: '',
      width: '160px',
      render: (alert: Alert) => (
        <div className="flex items-center gap-2">
          {alert.status === 'active' && (
            <button
              onClick={() => handleAcknowledge(alert.id)}
              className="px-3 py-1.5 text-xs font-medium border border-border rounded hover:bg-accent transition-colors"
            >
              <CheckCircle className="w-3.5 h-3.5 inline mr-1" />
              Acknowledge
            </button>
          )}
          <button
            onClick={() => handleDelete(alert.id)}
            className="p-1.5 text-text-secondary hover:text-danger hover:bg-danger/10 rounded transition-colors"
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <PageHeader
        title="Alerts"
        description="Server monitoring alerts and notifications"
        breadcrumbs={[{ label: 'Alerts' }]}
        actions={
          <Button variant="outline" onClick={() => setShowSettingsModal(true)}>
            <Settings className="w-4 h-4 mr-2" />
            Alert Settings
          </Button>
        }
      />

      <Card padding="none">
        <div className="p-4 flex items-center justify-between border-b border-border">
          <div className="flex items-center gap-3">
            <div className={cn('p-2 rounded',
              alerts.filter(a => a.status === 'active').length > 0 ? 'bg-danger/10' : 'bg-success/10')}>
              <Bell className={cn('w-5 h-5',
                alerts.filter(a => a.status === 'active').length > 0 ? 'text-danger' : 'text-success')} />
            </div>
            <div>
              <p className="font-medium">
                {alerts.filter(a => a.status === 'active').length} active alerts
              </p>
              <p className="text-xs text-text-secondary">Real-time monitoring</p>
            </div>
          </div>
          <Button variant="outline" size="sm" onClick={loadAlerts}>
            <RefreshCw className="w-3.5 h-3.5 mr-1" />
            Refresh
          </Button>
        </div>
        <Table
          columns={columns}
          data={alerts}
          keyField="id"
          loading={loading}
          emptyMessage="No alerts — everything looks good!"
        />
      </Card>

      <Modal
        open={showSettingsModal}
        onClose={() => setShowSettingsModal(false)}
        title="Alert Settings"
        size="md"
      >
        <div className="space-y-6">
          <FormGroup>
            <Label>Website Down Alert · Notify when site stops responding</Label>
            <div className="flex items-center gap-3">
              <Switch
                checked={alertSettings.website_down_enabled}
                onChange={(e) => setAlertSettings({ ...alertSettings, website_down_enabled: e.target.checked })}
              />
              <span className="text-sm">Enabled</span>
              <div className="flex items-center gap-2 ml-4">
                <span className="text-sm text-text-secondary">Threshold:</span>
                <Input
                  type="number"
                  value={alertSettings.website_down_threshold}
                  onChange={(e) => setAlertSettings({ ...alertSettings, website_down_threshold: parseInt(e.target.value) || 5 })}
                  className="w-20"
                />
                <span className="text-sm text-text-secondary">checks</span>
              </div>
            </div>
          </FormGroup>
          <FormGroup>
            <Label>SSL Expiry Alert · Warn before cert expires</Label>
            <div className="flex items-center gap-3">
              <Switch
                checked={alertSettings.ssl_expiry_enabled}
                onChange={(e) => setAlertSettings({ ...alertSettings, ssl_expiry_enabled: e.target.checked })}
              />
              <span className="text-sm">Enabled</span>
              <div className="flex items-center gap-2 ml-4">
                <span className="text-sm text-text-secondary">Warn</span>
                <Input
                  type="number"
                  value={alertSettings.ssl_expiry_days}
                  onChange={(e) => setAlertSettings({ ...alertSettings, ssl_expiry_days: parseInt(e.target.value) || 14 })}
                  className="w-20"
                />
                <span className="text-sm text-text-secondary">days before expiry</span>
              </div>
            </div>
          </FormGroup>
          <FormGroup>
            <Label>Disk Usage Alert · Alert when storage fills up</Label>
            <div className="flex items-center gap-3">
              <Switch
                checked={alertSettings.disk_usage_enabled}
                onChange={(e) => setAlertSettings({ ...alertSettings, disk_usage_enabled: e.target.checked })}
              />
              <span className="text-sm">Enabled</span>
              <div className="flex items-center gap-2 ml-4">
                <span className="text-sm text-text-secondary">Threshold:</span>
                <input
                  type="range"
                  min="50"
                  max="100"
                  value={alertSettings.disk_usage_threshold}
                  onChange={(e) => setAlertSettings({ ...alertSettings, disk_usage_threshold: parseInt(e.target.value) })}
                  className="flex-1"
                />
                <span className="text-sm font-mono w-12">{alertSettings.disk_usage_threshold}%</span>
              </div>
            </div>
          </FormGroup>
          <FormGroup>
            <Label>Backup Overdue Alert · Alert when no recent backup</Label>
            <div className="flex items-center gap-3">
              <Switch
                checked={alertSettings.backup_overdue_enabled}
                onChange={(e) => setAlertSettings({ ...alertSettings, backup_overdue_enabled: e.target.checked })}
              />
              <span className="text-sm">Enabled</span>
              <div className="flex items-center gap-2 ml-4">
                <span className="text-sm text-text-secondary">After</span>
                <Input
                  type="number"
                  value={alertSettings.backup_overdue_days}
                  onChange={(e) => setAlertSettings({ ...alertSettings, backup_overdue_days: parseInt(e.target.value) || 7 })}
                  className="w-20"
                />
                <span className="text-sm text-text-secondary">days without backup</span>
              </div>
            </div>
          </FormGroup>
          <FormGroup>
            <Label>Service Down Alert · Alert when service stops</Label>
            <div className="flex items-center gap-3">
              <Switch
                checked={alertSettings.service_down_enabled}
                onChange={(e) => setAlertSettings({ ...alertSettings, service_down_enabled: e.target.checked })}
              />
              <span className="text-sm">Alert when a system service stops</span>
            </div>
          </FormGroup>
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="outline" onClick={() => setShowSettingsModal(false)}>Cancel</Button>
            <Button onClick={saveSettings} loading={saving}>Save Settings</Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
