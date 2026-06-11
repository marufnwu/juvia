import { useState, useEffect } from 'react'
import api from '../../lib/api'

type Tab = 'general' | 'email' | 'security' | 'updates' | 'users' | 'backups' | 'notifications' | 'audit' | 'export'

interface Settings {
  panel_port?: string
  panel_hostname?: string
  smtp_host?: string
  smtp_port?: string
  smtp_username?: string
  smtp_from?: string
  smtp_enabled?: string
  session_timeout?: string
  auto_update?: string
  update_channel?: string
  backup_remote_storage?: string
  backup_retention_days?: string
  alert_disk_usage_threshold?: string
  alert_ssl_expiry_days?: string
}

interface VersionInfo {
  version: string
  current_version: string
  latest_version: string
  download_url?: string
  changelog?: string
}

export default function Settings() {
  const [activeTab, setActiveTab] = useState<Tab>('general')
  const [settings, setSettings] = useState<Settings>({})
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const [versionInfo, setVersionInfo] = useState<VersionInfo | null>(null)

  useEffect(() => {
    fetchSettings()
  }, [])

  const fetchSettings = async () => {
    try {
      const res = await api.get('/settings')
      const data = res.data
      if (data.success) {
        setSettings(data.data)
      }
    } catch (err) {
      console.error('Failed to fetch settings:', err)
    } finally {
      setLoading(false)
    }
  }

  const saveSettings = async (newSettings: Partial<Settings>) => {
    setSaving(true)
    try {
      const res = await api.put('/settings', newSettings)
      const data = res.data
      if (data.success) {
        setMessage({ type: 'success', text: 'Settings saved successfully' })
        fetchSettings()
      } else {
        setMessage({ type: 'error', text: data.error?.message || 'Failed to save settings' })
      }
    } catch (err) {
      setMessage({ type: 'error', text: 'Failed to save settings' })
    } finally {
      setSaving(false)
    }
  }

  const checkForUpdates = async () => {
    try {
      const res = await api.get('/updates/check')
      const data = res.data
      if (data.success) {
        setVersionInfo(data.data)
      }
    } catch (err) {
      console.error('Failed to check for updates:', err)
    }
  }

  const tabs: { id: Tab; label: string; description: string }[] = [
    { id: 'general', label: 'General', description: 'Panel port, hostname, server name' },
    { id: 'email', label: 'Email', description: 'SMTP configuration for sending emails' },
    { id: 'security', label: 'Security', description: '2FA enforcement, session timeout' },
    { id: 'updates', label: 'Updates', description: 'Auto-update settings, check for updates' },
    { id: 'backups', label: 'Backups', description: 'Remote storage, retention settings' },
    { id: 'notifications', label: 'Notifications', description: 'Alert thresholds, email notifications' },
    { id: 'audit', label: 'Audit Log', description: 'View all actions taken on this server' },
    { id: 'export', label: 'Export/Import', description: 'Export or import full configuration' },
  ]

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-muted-foreground">Loading settings...</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Settings</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Configure your server control panel
        </p>
      </div>

      {message && (
        <div className={`p-3 rounded-lg ${message.type === 'success' ? 'bg-green-500/10 text-green-500' : 'bg-red-500/10 text-red-500'}`}>
          {message.text}
        </div>
      )}

      <div className="flex gap-6">
        <div className="w-48 shrink-0">
          <nav className="space-y-1">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                  activeTab === tab.id
                    ? 'bg-primary text-primary-foreground'
                    : 'hover:bg-accent'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        <div className="flex-1 border rounded-lg p-6">
          {activeTab === 'general' && (
            <div className="space-y-4">
              <h2 className="text-lg font-medium">General Settings</h2>
              <p className="text-sm text-muted-foreground">
                Basic panel configuration
              </p>
              <div className="space-y-4 max-w-md">
                <div>
                  <label className="text-sm font-medium">Panel Port</label>
                  <input
                    type="number"
                    value={settings.panel_port || '8080'}
                    onChange={(e) => setSettings({ ...settings, panel_port: e.target.value })}
                    className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
                  />
                  <p className="text-xs text-muted-foreground mt-1">
                    The port number the control panel listens on
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium">Hostname</label>
                  <input
                    type="text"
                    value={settings.panel_hostname || ''}
                    onChange={(e) => setSettings({ ...settings, panel_hostname: e.target.value })}
                    className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
                    placeholder="panel.example.com"
                  />
                  <p className="text-xs text-muted-foreground mt-1">
                    The hostname used to access this panel
                  </p>
                </div>
                <button
                  onClick={() => saveSettings(settings)}
                  disabled={saving}
                  className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50"
                >
                  {saving ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </div>
          )}

          {activeTab === 'email' && (
            <div className="space-y-4">
              <h2 className="text-lg font-medium">Email Settings</h2>
              <p className="text-sm text-muted-foreground">
                SMTP configuration for sending notifications and alerts
              </p>
              <div className="space-y-4 max-w-md">
                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    id="smtp_enabled"
                    checked={settings.smtp_enabled === 'true'}
                    onChange={(e) => setSettings({ ...settings, smtp_enabled: e.target.checked ? 'true' : 'false' })}
                  />
                  <label htmlFor="smtp_enabled" className="text-sm font-medium">Enable SMTP</label>
                </div>
                <div>
                  <label className="text-sm font-medium">SMTP Host</label>
                  <input
                    type="text"
                    value={settings.smtp_host || ''}
                    onChange={(e) => setSettings({ ...settings, smtp_host: e.target.value })}
                    className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
                    placeholder="smtp.example.com"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">SMTP Port</label>
                  <input
                    type="number"
                    value={settings.smtp_port || '587'}
                    onChange={(e) => setSettings({ ...settings, smtp_port: e.target.value })}
                    className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">Username</label>
                  <input
                    type="text"
                    value={settings.smtp_username || ''}
                    onChange={(e) => setSettings({ ...settings, smtp_username: e.target.value })}
                    className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">From Address</label>
                  <input
                    type="email"
                    value={settings.smtp_from || ''}
                    onChange={(e) => setSettings({ ...settings, smtp_from: e.target.value })}
                    className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
                    placeholder="noreply@example.com"
                  />
                </div>
                <button
                  onClick={() => saveSettings(settings)}
                  disabled={saving}
                  className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50"
                >
                  {saving ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </div>
          )}

          {activeTab === 'security' && (
            <div className="space-y-4">
              <h2 className="text-lg font-medium">Security Settings</h2>
              <p className="text-sm text-muted-foreground">
                Control panel security options
              </p>
              <div className="space-y-4 max-w-md">
                <div>
                  <label className="text-sm font-medium">Session Timeout (minutes)</label>
                  <input
                    type="number"
                    value={settings.session_timeout || '60'}
                    onChange={(e) => setSettings({ ...settings, session_timeout: e.target.value })}
                    className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
                  />
                  <p className="text-xs text-muted-foreground mt-1">
                    How long before inactive sessions expire
                  </p>
                </div>
                <button
                  onClick={() => saveSettings(settings)}
                  disabled={saving}
                  className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50"
                >
                  {saving ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </div>
          )}

          {activeTab === 'updates' && (
            <div className="space-y-4">
              <h2 className="text-lg font-medium">Update Settings</h2>
              <p className="text-sm text-muted-foreground">
                Panel update configuration
              </p>
              <div className="space-y-4 max-w-md">
                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    id="auto_update"
                    checked={settings.auto_update === 'true'}
                    onChange={(e) => setSettings({ ...settings, auto_update: e.target.checked ? 'true' : 'false' })}
                  />
                  <label htmlFor="auto_update" className="text-sm font-medium">Enable Auto-Update</label>
                </div>
                <div>
                  <label className="text-sm font-medium">Update Channel</label>
                  <select
                    value={settings.update_channel || 'stable'}
                    onChange={(e) => setSettings({ ...settings, update_channel: e.target.value })}
                    className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
                  >
                    <option value="stable">Stable</option>
                    <option value="beta">Beta</option>
                  </select>
                </div>
                <div className="p-4 border rounded-lg">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-medium">Current Version</p>
                      <p className="text-sm text-muted-foreground">{versionInfo?.version || '0.1.0'}</p>
                    </div>
                    <button
                      onClick={checkForUpdates}
                      className="px-4 py-2 border rounded-lg hover:bg-accent"
                    >
                      Check for Updates
                    </button>
                  </div>
                  {versionInfo && versionInfo.latest_version !== versionInfo.version && (
                    <div className="mt-4 p-3 bg-primary/10 rounded-lg">
                      <p className="text-sm">Version {versionInfo.latest_version} is available</p>
                    </div>
                  )}
                </div>
                <button
                  onClick={() => saveSettings(settings)}
                  disabled={saving}
                  className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50"
                >
                  {saving ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </div>
          )}

          {activeTab === 'backups' && (
            <div className="space-y-4">
              <h2 className="text-lg font-medium">Backup Settings</h2>
              <p className="text-sm text-muted-foreground">
                Configure backup storage and retention
              </p>
              <div className="space-y-4 max-w-md">
                <div>
                  <label className="text-sm font-medium">Remote Storage</label>
                  <select
                    value={settings.backup_remote_storage || 'local'}
                    onChange={(e) => setSettings({ ...settings, backup_remote_storage: e.target.value })}
                    className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
                  >
                    <option value="local">Local</option>
                    <option value="s3">Amazon S3</option>
                    <option value="r2">Cloudflare R2</option>
                    <option value="b2">Backblaze B2</option>
                  </select>
                </div>
                <div>
                  <label className="text-sm font-medium">Retention Days</label>
                  <input
                    type="number"
                    value={settings.backup_retention_days || '30'}
                    onChange={(e) => setSettings({ ...settings, backup_retention_days: e.target.value })}
                    className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
                  />
                  <p className="text-xs text-muted-foreground mt-1">
                    How long to keep backups before deletion
                  </p>
                </div>
                <button
                  onClick={() => saveSettings(settings)}
                  disabled={saving}
                  className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50"
                >
                  {saving ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </div>
          )}

          {activeTab === 'notifications' && (
            <div className="space-y-4">
              <h2 className="text-lg font-medium">Notification Settings</h2>
              <p className="text-sm text-muted-foreground">
                Configure alert thresholds
              </p>
              <div className="space-y-4 max-w-md">
                <div>
                  <label className="text-sm font-medium">Disk Usage Alert Threshold (%)</label>
                  <input
                    type="number"
                    value={settings.alert_disk_usage_threshold || '85'}
                    onChange={(e) => setSettings({ ...settings, alert_disk_usage_threshold: e.target.value })}
                    className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
                  />
                  <p className="text-xs text-muted-foreground mt-1">
                    Alert when disk usage exceeds this percentage
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium">SSL Expiry Alert (days)</label>
                  <input
                    type="number"
                    value={settings.alert_ssl_expiry_days || '30'}
                    onChange={(e) => setSettings({ ...settings, alert_ssl_expiry_days: e.target.value })}
                    className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
                  />
                  <p className="text-xs text-muted-foreground mt-1">
                    Alert when SSL certificates expire within this many days
                  </p>
                </div>
                <button
                  onClick={() => saveSettings(settings)}
                  disabled={saving}
                  className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50"
                >
                  {saving ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </div>
          )}

          {activeTab === 'audit' && (
            <div className="space-y-4">
              <h2 className="text-lg font-medium">Audit Log</h2>
              <p className="text-sm text-muted-foreground">
                All actions taken on this server
              </p>
              <AuditLog />
            </div>
          )}

          {activeTab === 'export' && (
            <div className="space-y-4">
              <h2 className="text-lg font-medium">Export / Import Configuration</h2>
              <p className="text-sm text-muted-foreground">
                Export your full configuration or import from a backup
              </p>
              <div className="space-y-4 max-w-md">
                <div className="p-4 border rounded-lg">
                  <h3 className="font-medium">Export Configuration</h3>
                  <p className="text-sm text-muted-foreground mt-1">
                    Download all websites, databases, email accounts, and settings as JSON
                  </p>
                  <button className="mt-3 px-4 py-2 border rounded-lg hover:bg-accent">
                    Export JSON
                  </button>
                </div>
                <div className="p-4 border rounded-lg">
                  <h3 className="font-medium">Import Configuration</h3>
                  <p className="text-sm text-muted-foreground mt-1">
                    Restore configuration from a previously exported JSON file
                  </p>
                  <input type="file" accept=".json" className="mt-3 block" />
                  <button className="mt-3 px-4 py-2 border rounded-lg hover:bg-accent">
                    Import
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function AuditLog() {
  const [entries, setEntries] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchAuditLog()
  }, [])

  const fetchAuditLog = async () => {
    try {
      const res = await api.get('/audit-log')
      const data = res.data
      if (data.success) {
        setEntries(data.data)
      }
    } catch (err) {
      console.error('Failed to fetch audit log:', err)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return <div className="text-muted-foreground">Loading audit log...</div>
  }

  if (entries.length === 0) {
    return <div className="text-muted-foreground">No audit log entries yet</div>
  }

  return (
    <div className="border rounded-lg">
      <table className="w-full">
        <thead>
          <tr className="border-b bg-muted/50">
            <th className="text-left p-3 text-sm font-medium">Timestamp</th>
            <th className="text-left p-3 text-sm font-medium">Action</th>
            <th className="text-left p-3 text-sm font-medium">IP Address</th>
            <th className="text-left p-3 text-sm font-medium">Details</th>
          </tr>
        </thead>
        <tbody>
          {entries.map((entry) => (
            <tr key={entry.id} className="border-b hover:bg-muted/30">
              <td className="p-3 text-sm">{new Date(entry.created_at).toLocaleString()}</td>
              <td className="p-3 text-sm">{entry.action}</td>
              <td className="p-3 text-sm font-mono text-xs">{entry.ip_address}</td>
              <td className="p-3 text-sm">{entry.details}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}