import { useEffect, useRef, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { Settings, Shield, Bell, Database, Globe, RefreshCw, Save, Download, Upload, Copy, CheckCircle2, X, AlertTriangle } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Input, Label, Select, FormGroup, Switch } from '../../components/ui/Input'
import { Tabs, TabPanel } from '../../components/ui/Tabs'
import { PageHeader } from '../../components/ui/Misc'
import { Modal } from '../../components/ui/Modal'
import api from '../../lib/api'
import { cn } from '../../lib/utils'
import { useApiError } from '../../hooks/useToast'
import { useAuthStore } from '../../stores/authStore'

export default function SettingsPage() {
  const showError = useApiError()
  const user = useAuthStore((s) => s.user)
  const refreshUser = useAuthStore((s) => s.refreshUser)
  const location = useLocation()
  const [activeTab, setActiveTab] = useState((location.state as { activeTab?: string } | null)?.activeTab || 'general')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [settings, setSettings] = useState<Record<string, string>>({})

  const [generalForm, setGeneralForm] = useState({
    server_name: '',
    hostname: '',
    timezone: 'UTC',
  })

  const [smtpForm, setSmtpForm] = useState({
    smtp_host: '',
    smtp_port: '587',
    smtp_username: '',
    smtp_password: '',
  })

  const [backupForm, setBackupForm] = useState({
    backup_storage: 'local',
    retention_days: '30',
  })

  const [alertsForm, setAlertsForm] = useState({
    email_on_alert: false,
    ssl_expiry_warnings: false,
    backup_failures: false,
    website_down_alerts: false,
  })

  const [twoFAState, setTwoFAState] = useState<'idle' | 'setup' | 'verify' | 'done'>('idle')
  const [twoFASecret, setTwoFASecret] = useState('')
  const [twoFAURI, setTwoFAURI] = useState('')
  const [twoFACode, setTwoFACode] = useState('')
  const [twoFARecoveryCodes, setTwoFARecoveryCodes] = useState<string[]>([])
  const [twoFALoading, setTwoFALoading] = useState(false)
  const [twoFAError, setTwoFAError] = useState('')
  const [showRecoveryModal, setShowRecoveryModal] = useState(false)
  const sessions = useAuthStore((s) => s.sessions)
  const [revokingSession, setRevokingSession] = useState<string | null>(null)
  const [updateInfo, setUpdateInfo] = useState<{ current_version: string; latest_version: string; download_url: string; changelog: string; sig_url: string } | null>(null)
  const [checkingUpdate, setCheckingUpdate] = useState(false)
  const [downloadingUpdate, setDownloadingUpdate] = useState(false)
  const [applyingUpdate, setApplyingUpdate] = useState(false)
  const [updateMessage, setUpdateMessage] = useState('')
  const [updateDownloadPath, setUpdateDownloadPath] = useState('')

  const [sshPort, setSshPort] = useState('22')
  const [sshRootLogin, setSshRootLogin] = useState(false)
  const [savingSSH, setSavingSSH] = useState(false)

  const [autoUpdate, setAutoUpdate] = useState(false)
  const [updateChannel, setUpdateChannel] = useState('stable')

  const [nsBrandDomain, setNsBrandDomain] = useState('')
  const [ns1Hostname, setNs1Hostname] = useState('')
  const [ns2Hostname, setNs2Hostname] = useState('')
  const [serverIP, setServerIP] = useState('')
  const [savingNS, setSavingNS] = useState(false)
  const [refreshingIP, setRefreshingIP] = useState(false)

  const [importing, setImporting] = useState(false)
  const [showImportConfirm, setShowImportConfirm] = useState(false)
  const [pendingImportData, setPendingImportData] = useState<string>('')
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    loadSettings()
    loadSSHSettings()
    refreshUser()
  }, [])

  const loadSettings = async () => {
    try {
      const res = await api.get('/settings')
      const data = res.data.data || {}
      setSettings(data)
      setGeneralForm({
        server_name: data.server_name || '',
        hostname: data.hostname || '',
        timezone: data.timezone || 'UTC',
      })
      setSmtpForm({
        smtp_host: data.smtp_host || '',
        smtp_port: data.smtp_port || '587',
        smtp_username: data.smtp_username || '',
        smtp_password: '',
      })
      setBackupForm({
        backup_storage: data.backup_storage || 'local',
        retention_days: data.retention_days || '30',
      })
      setAlertsForm({
        email_on_alert: data.email_on_alert === 'true',
        ssl_expiry_warnings: data.ssl_expiry_warnings === 'true',
        backup_failures: data.backup_failures === 'true',
        website_down_alerts: data.website_down_alerts === 'true',
      })
      setAutoUpdate(data.auto_update === 'true')
      setUpdateChannel(data.update_channel || 'stable')
      setNsBrandDomain(data.ns_brand_domain || '')
      setNs1Hostname(data.ns1_hostname || '')
      setNs2Hostname(data.ns2_hostname || '')
      setServerIP(data.server_ip || '')
    } catch (err: any) {
      showError(err, 'Failed to load settings')
    } finally {
      setLoading(false)
    }
  }

  const saveGeneral = async () => {
    setSaving(true)
    try {
      await api.put('/settings', generalForm)
    } catch (err: any) {
      showError(err, 'Failed to save settings')
    } finally {
      setSaving(false)
    }
  }

  const saveSMTP = async () => {
    setSaving(true)
    try {
      await api.put('/settings', smtpForm)
    } catch (err: any) {
      showError(err, 'Failed to save SMTP settings')
    } finally {
      setSaving(false)
    }
  }

  const saveBackup = async () => {
    setSaving(true)
    try {
      await api.put('/settings', backupForm)
    } catch (err: any) {
      showError(err, 'Failed to save backup settings')
    } finally {
      setSaving(false)
    }
  }

  const saveAlerts = async () => {
    setSaving(true)
    try {
      await api.put('/settings/alerts', alertsForm)
    } catch (err: any) {
      showError(err, 'Failed to save alert settings')
    } finally {
      setSaving(false)
    }
  }

  const saveAutoUpdate = async () => {
    setSaving(true)
    try {
      await api.put('/settings', { auto_update: autoUpdate, update_channel: updateChannel })
    } catch (err: any) {
      showError(err, 'Failed to save auto-update settings')
    } finally {
      setSaving(false)
    }
  }

  const start2FA = async () => {
    setTwoFALoading(true)
    setTwoFAError('')
    try {
      const res = await api.post('/auth/2fa/enable')
      setTwoFASecret(res.data.data.secret)
      setTwoFAURI(res.data.data.uri)
      setTwoFAState('setup')
    } catch (err: any) {
      showError(err, 'Failed to start 2FA setup')
    } finally {
      setTwoFALoading(false)
    }
  }

  const verify2FA = async (e: React.FormEvent) => {
    e.preventDefault()
    setTwoFALoading(true)
    setTwoFAError('')
    try {
      const res = await api.post('/auth/2fa/verify', {
        secret: twoFASecret,
        code: twoFACode,
      })
      setTwoFARecoveryCodes(res.data.data.recovery_codes || [])
      setTwoFAState('done')
      setShowRecoveryModal(true)
      refreshUser()
    } catch (err: any) {
      setTwoFAError(err.response?.data?.error?.user_message || 'Invalid code. Please try again.')
    } finally {
      setTwoFALoading(false)
    }
  }

  const close2FAModal = () => {
    setShowRecoveryModal(false)
    setTwoFAState('idle')
    setTwoFASecret('')
    setTwoFAURI('')
    setTwoFACode('')
    setTwoFARecoveryCodes([])
    setTwoFAError('')
  }

  const revokeSession = async (sessionId: string) => {
    setRevokingSession(sessionId)
    try {
      await api.delete(`/auth/sessions/${sessionId}`)
      refreshUser()
    } catch (err: any) {
      showError(err, 'Failed to revoke session')
    } finally {
      setRevokingSession(null)
    }
  }

  const checkForUpdates = async () => {
    setCheckingUpdate(true)
    setUpdateMessage('')
    try {
      const res = await api.get('/updates/check')
      setUpdateInfo(res.data.data)
      if (res.data.data.latest_version === res.data.data.current_version) {
        setUpdateMessage('You are up to date')
      } else {
        setUpdateMessage(`Update available: v${res.data.data.latest_version}`)
      }
    } catch (err: any) {
      showError(err, 'Failed to check for updates')
    } finally {
      setCheckingUpdate(false)
    }
  }

  const downloadUpdate = async () => {
    if (!updateInfo?.download_url) return
    setDownloadingUpdate(true)
    setUpdateMessage('')
    try {
      const res = await api.post('/updates/download', { url: updateInfo.download_url })
      setUpdateDownloadPath(res.data.data.path)
      setUpdateMessage('Update downloaded. Ready to apply.')
    } catch (err: any) {
      showError(err, 'Failed to download update')
    } finally {
      setDownloadingUpdate(false)
    }
  }

  const applyUpdate = async () => {
    if (!updateDownloadPath) return
    setApplyingUpdate(true)
    setUpdateMessage('')
    try {
      await api.post('/updates/apply', {
        path: updateDownloadPath,
        sig_url: updateInfo?.sig_url || '',
      })
      setUpdateMessage('Update applied. Panel is restarting...')
    } catch (err: any) {
      showError(err, 'Failed to apply update')
    } finally {
      setApplyingUpdate(false)
    }
  }

  const copySecret = () => {
    navigator.clipboard.writeText(twoFASecret)
  }

  const loadSSHSettings = async () => {
    try {
      const res = await api.get('/settings/ssh')
      const data = res.data.data || {}
      setSshPort(String(data.port || 22))
      setSshRootLogin(data.root_login || false)
    } catch (err: any) {
      showError(err, 'Failed to load SSH settings')
    }
  }

  const saveSSH = async () => {
    setSavingSSH(true)
    try {
      await api.put('/settings/ssh', { port: parseInt(sshPort), root_login: sshRootLogin })
    } catch (err: any) {
      showError(err, 'Failed to save SSH settings')
    } finally {
      setSavingSSH(false)
    }
  }

  const saveNameservers = async () => {
    setSavingNS(true)
    try {
      await api.put('/settings', {
        ns_brand_domain: nsBrandDomain,
        ns1_hostname: ns1Hostname,
        ns2_hostname: ns2Hostname,
      })
    } catch (err: any) {
      showError(err, 'Failed to save nameserver settings')
    } finally {
      setSavingNS(false)
    }
  }

  const refreshServerIP = async () => {
    setRefreshingIP(true)
    try {
      const res = await api.post('/server/ip/refresh', {})
      setServerIP(res.data.data?.ip || '')
    } catch (err: any) {
      showError(err, 'Failed to refresh server IP')
    } finally {
      setRefreshingIP(false)
    }
  }

  const handleExportConfig = async () => {
    try {
      const res = await api.get('/settings/export')
      const blob = new Blob([JSON.stringify(res.data.data, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'juvia-config.json'
      a.click()
      URL.revokeObjectURL(url)
    } catch (err: any) {
      showError(err, 'Failed to export config')
    }
  }

  const handleImportConfig = () => {
    fileInputRef.current?.click()
  }

  const onFileSelected = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = () => {
      const content = reader.result as string
      try {
        JSON.parse(content)
        setPendingImportData(content)
        setShowImportConfirm(true)
      } catch {
        showError(new Error('Invalid JSON'), 'Invalid config file')
      }
    }
    reader.readAsText(file)
    e.target.value = ''
  }

  const confirmImport = async () => {
    setImporting(true)
    try {
      await api.post('/settings/import', JSON.parse(pendingImportData))
      setShowImportConfirm(false)
      setPendingImportData('')
      await loadSettings()
      await loadSSHSettings()
    } catch (err: any) {
      showError(err, 'Failed to import config')
    } finally {
      setImporting(false)
    }
  }

  const tabs = [
    { id: 'general', label: 'General' },
    { id: 'security', label: 'Security' },
    { id: 'email', label: 'Email' },
    { id: 'updates', label: 'Updates' },
    { id: 'backup', label: 'Backup' },
    { id: 'notifications', label: 'Notifications' },
  ]

  if (loading) {
    return (
      <div className="space-y-6">
        <PageHeader title="Settings" description="Panel configuration and preferences" breadcrumbs={[{ label: 'Settings' }]} />
        <div className="flex items-center justify-center h-64">
          <RefreshCw className="w-6 h-6 animate-spin text-text-secondary" />
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Settings"
        description="Panel configuration and preferences"
        breadcrumbs={[{ label: 'Settings' }]}
      />

      <Tabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />

      {activeTab === 'general' && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Server Identity · Basic server information</CardTitle>
            </CardHeader>
            <div className="space-y-4">
              <FormGroup>
                <Label>Server Name</Label>
                <Input
                  value={generalForm.server_name}
                  onChange={(e) => setGeneralForm({ ...generalForm, server_name: e.target.value })}
                />
              </FormGroup>
              <FormGroup>
                <Label>Hostname · Your server's network name</Label>
                <Input
                  value={generalForm.hostname}
                  onChange={(e) => setGeneralForm({ ...generalForm, hostname: e.target.value })}
                />
              </FormGroup>
              <FormGroup>
                <Label>Timezone</Label>
                <Select
                  value={generalForm.timezone}
                  onChange={(e) => setGeneralForm({ ...generalForm, timezone: e.target.value })}
                >
                  <option value="UTC">UTC</option>
                  <option value="America/New_York">America/New_York</option>
                  <option value="Europe/London">Europe/London</option>
                  <option value="Asia/Dhaka">Asia/Dhaka</option>
                </Select>
              </FormGroup>
              <Button onClick={saveGeneral} loading={saving}>
                <Save className="w-4 h-4 mr-2" />
                Save Changes
              </Button>
            </div>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>SSH Settings · Secure remote login</CardTitle>
            </CardHeader>
            <div className="space-y-4">
              <FormGroup>
                <Label>SSH Port · Port for remote login (default 22)</Label>
                <Input value={sshPort} onChange={(e) => setSshPort(e.target.value)} type="number" className="w-32" />
              </FormGroup>
              <FormGroup>
                <Label>Root Login · Allow super-admin direct login</Label>
                <Switch checked={sshRootLogin} onChange={(e) => setSshRootLogin(e.target.checked)} />
              </FormGroup>
              <Button onClick={saveSSH} loading={savingSSH}>
                <Save className="w-4 h-4 mr-2" />
                Save SSH Settings
              </Button>
            </div>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Nameserver Settings · Configure authoritative DNS</CardTitle>
              <CardDescription>Set the brand domain and hostnames used for customer nameservers</CardDescription>
            </CardHeader>
            <div className="space-y-4">
              <FormGroup>
                <Label>Server IP · Public address to point domains to</Label>
                <div className="flex items-center gap-2">
                  <Input value={serverIP} readOnly className="font-mono flex-1" />
                  <Button variant="outline" size="sm" onClick={refreshServerIP} loading={refreshingIP}>
                    <RefreshCw className="w-3.5 h-3.5 mr-1.5" />
                    Refresh
                  </Button>
                </div>
              </FormGroup>
              <FormGroup>
                <Label>Nameserver Brand Domain · Base domain for NS hostnames</Label>
                <Input
                  value={nsBrandDomain}
                  onChange={(e) => {
                    const value = e.target.value
                    setNsBrandDomain(value)
                    if (value && !ns1Hostname) setNs1Hostname('ns1.' + value)
                    if (value && !ns2Hostname) setNs2Hostname('ns2.' + value)
                  }}
                  placeholder="yourbrand.com"
                />
              </FormGroup>
              <FormGroup>
                <Label>NS1 Hostname</Label>
                <Input
                  value={ns1Hostname}
                  onChange={(e) => setNs1Hostname(e.target.value)}
                  placeholder="ns1.yourbrand.com"
                  className="font-mono"
                />
              </FormGroup>
              <FormGroup>
                <Label>NS2 Hostname</Label>
                <Input
                  value={ns2Hostname}
                  onChange={(e) => setNs2Hostname(e.target.value)}
                  placeholder="ns2.yourbrand.com"
                  className="font-mono"
                />
              </FormGroup>
              {ns1Hostname && ns2Hostname && (
                <div className="p-3 bg-accent/30 rounded text-sm text-text-secondary">
                  <p>Tell your customers to set their domain nameservers to:</p>
                  <code className="block mt-1 font-mono">{ns1Hostname}</code>
                  <code className="block font-mono">{ns2Hostname}</code>
                </div>
              )}
              <Button onClick={saveNameservers} loading={savingNS}>
                <Save className="w-4 h-4 mr-2" />
                Save Nameserver Settings
              </Button>
            </div>
          </Card>
        </div>
      )}

      {activeTab === 'security' && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Two-Factor Authentication</CardTitle>
              <CardDescription>Add an extra layer of security to your account</CardDescription>
            </CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">TOTP Authenticator · Phone app for login codes</p>
              </div>
              {user?.two_factor_enabled ? (
                <Badge variant="success">Enabled</Badge>
              ) : twoFAState === 'idle' ? (
                <Button variant="outline" onClick={start2FA} loading={twoFALoading}>
                  Enable 2FA
                </Button>
              ) : (
                <Button variant="outline" onClick={() => setTwoFAState('idle')}>
                  Cancel
                </Button>
              )}
            </div>

            {twoFAState === 'setup' && (
              <div className="mt-4 p-4 bg-accent/30 rounded space-y-4">
                <div className="flex items-start gap-3">
                  {twoFAURI && (
                    <div className="w-32 h-32 bg-white rounded flex items-center justify-center flex-shrink-0">
                      <img src={twoFAURI} alt="QR Code" className="w-28 h-28" />
                    </div>
                  )}
                  <div className="flex-1">
                    <p className="text-sm font-medium mb-1">Scan QR Code</p>
                    <p className="text-xs text-text-secondary mb-2">Or enter this key manually in your authenticator app:</p>
                    <div className="flex items-center gap-2 p-2 bg-background border border-border rounded">
                      <code className="text-xs font-mono flex-1 break-all">{twoFASecret}</code>
                      <Button variant="ghost" size="sm" onClick={copySecret}>
                        <Copy className="w-3.5 h-3.5" />
                      </Button>
                    </div>
                  </div>
                </div>
                <form onSubmit={verify2FA} className="space-y-3">
                  <div>
                    <Label>Enter the 6-digit code from your app</Label>
                    <Input
                      value={twoFACode}
                      onChange={(e) => setTwoFACode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                      placeholder="000000"
                      className="font-mono text-center tracking-widest"
                      maxLength={6}
                    />
                  </div>
                  {twoFAError && (
                    <p className="text-sm text-danger">{twoFAError}</p>
                  )}
                  <Button type="submit" loading={twoFALoading} disabled={twoFACode.length !== 6}>
                    Verify and Enable
                  </Button>
                </form>
              </div>
            )}
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Session Management</CardTitle>
              <CardDescription>Manage active sessions</CardDescription>
            </CardHeader>
            <div className="space-y-3">
              {sessions.length === 0 ? (
                <p className="text-sm text-text-secondary text-center py-4">No active sessions</p>
              ) : (
                sessions.map((session) => (
                  <div key={session.id} className="flex items-center justify-between p-3 bg-accent/30 rounded">
                    <div>
                      <div className="flex items-center gap-2">
                        <p className="font-medium text-sm">{session.ip}</p>
                        <Badge variant="success">Active</Badge>
                      </div>
                      <p className="text-xs text-text-secondary">{session.user_agent || 'Unknown browser'}</p>
                      <p className="text-xs text-text-secondary">Last active: {session.last_active ? new Date(session.last_active).toLocaleString() : 'just now'}</p>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => revokeSession(session.id)}
                      loading={revokingSession === session.id}
                      className="text-danger hover:text-danger"
                    >
                      Revoke
                    </Button>
                  </div>
                ))
              )}
            </div>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Login History</CardTitle>
              <CardDescription>Recent login attempts</CardDescription>
            </CardHeader>
            <div className="text-sm text-text-secondary text-center py-4">
              No recent login activity
            </div>
          </Card>
        </div>
      )}

      {activeTab === 'email' && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>SMTP Configuration · Outgoing email settings</CardTitle>
            </CardHeader>
            <div className="space-y-4">
              <FormGroup>
                <Label>SMTP Host · Mail server address</Label>
                <Input
                  value={smtpForm.smtp_host}
                  onChange={(e) => setSmtpForm({ ...smtpForm, smtp_host: e.target.value })}
                />
              </FormGroup>
              <FormGroup>
                <Label>SMTP Port · Email port (usually 587 or 465)</Label>
                <Input
                  value={smtpForm.smtp_port}
                  onChange={(e) => setSmtpForm({ ...smtpForm, smtp_port: e.target.value })}
                  type="number"
                  className="w-32"
                />
              </FormGroup>
              <FormGroup>
                <Label>SMTP Username · Login for mail server</Label>
                <Input
                  value={smtpForm.smtp_username}
                  onChange={(e) => setSmtpForm({ ...smtpForm, smtp_username: e.target.value })}
                />
              </FormGroup>
              <FormGroup>
                <Label>SMTP Password · Password for mail server</Label>
                <Input
                  type="password"
                  value={smtpForm.smtp_password}
                  onChange={(e) => setSmtpForm({ ...smtpForm, smtp_password: e.target.value })}
                  placeholder="Leave blank to keep current"
                />
              </FormGroup>
              <Button onClick={saveSMTP} loading={saving}>
                <Save className="w-4 h-4 mr-2" />
                Save SMTP Settings
              </Button>
            </div>
          </Card>
        </div>
      )}

      {activeTab === 'updates' && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Software Updates</CardTitle>
              <CardDescription>Keep your panel up to date</CardDescription>
            </CardHeader>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium">Current Version</p>
                  <p className="text-xs text-text-secondary mt-0.5">
                    v{updateInfo?.current_version || '0.1.0'}
                    {updateMessage && updateInfo?.latest_version === updateInfo?.current_version && ' — You are up to date'}
                    {updateMessage && updateInfo?.latest_version !== updateInfo?.current_version && ` — ${updateMessage}`}
                  </p>
                </div>
                <Button variant="outline" onClick={checkForUpdates} loading={checkingUpdate}>
                  <RefreshCw className="w-4 h-4 mr-2" />
                  Check for Updates
                </Button>
              </div>

              {updateInfo && updateInfo.latest_version !== updateInfo.current_version && (
                <div className="p-4 bg-accent/30 rounded space-y-3">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium">v{updateInfo.latest_version} available</p>
                      {updateInfo.changelog && (
                        <p className="text-xs text-text-secondary mt-1">{updateInfo.changelog}</p>
                      )}
                    </div>
                  </div>
                  {!updateDownloadPath ? (
                    <Button onClick={downloadUpdate} loading={downloadingUpdate} size="sm">
                      Download Update
                    </Button>
                  ) : (
                    <div className="flex items-center gap-2">
                      <Badge variant="success">Downloaded</Badge>
                      <Button onClick={applyUpdate} loading={applyingUpdate} size="sm">
                        Apply Update
                      </Button>
                    </div>
                  )}
                </div>
              )}

              {updateMessage && updateInfo?.latest_version === updateInfo?.current_version && (
                <div className="p-3 bg-success/10 border border-success/20 rounded flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-success" />
                  <span className="text-sm text-success">{updateMessage}</span>
                </div>
              )}
            </div>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Auto-Update · Get improvements automatically</CardTitle>
            </CardHeader>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm">Enable auto-updates</span>
                <Switch checked={autoUpdate} onChange={(e) => setAutoUpdate(e.target.checked)} />
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm">Update channel</span>
                <Select value={updateChannel} onChange={(e) => setUpdateChannel(e.target.value)} className="w-32">
                  <option value="stable">Stable</option>
                  <option value="beta">Beta</option>
                </Select>
              </div>
              <Button onClick={saveAutoUpdate} loading={saving}>
                <Save className="w-4 h-4 mr-2" />
                Save Auto-Update Settings
              </Button>
            </div>
          </Card>
        </div>
      )}

      {activeTab === 'backup' && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Backup Configuration</CardTitle>
              <CardDescription>Configure backup storage</CardDescription>
            </CardHeader>
            <div className="space-y-4">
              <FormGroup>
                <Label>Backup Storage</Label>
                <Select
                  value={backupForm.backup_storage}
                  onChange={(e) => setBackupForm({ ...backupForm, backup_storage: e.target.value })}
                >
                  <option value="local">Local Disk · On this server</option>
                  <option value="s3">Amazon S3 · Amazon cloud storage</option>
                  <option value="r2">Cloudflare R2 · Cloudflare cloud storage</option>
                  <option value="b2">Backblaze B2 · Backblaze cloud storage</option>
                </Select>
              </FormGroup>
              <FormGroup>
                <Label>Retention · Days to keep old backups</Label>
                <Input
                  value={backupForm.retention_days}
                  onChange={(e) => setBackupForm({ ...backupForm, retention_days: e.target.value })}
                  type="number"
                  className="w-32"
                />
              </FormGroup>
              <Button onClick={saveBackup} loading={saving}>
                <Save className="w-4 h-4 mr-2" />
                Save Backup Settings
              </Button>
            </div>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Export / Import</CardTitle>
              <CardDescription>Export or import panel configuration</CardDescription>
            </CardHeader>
            <div className="flex gap-3">
              <Button variant="outline" onClick={handleExportConfig}>
                <Download className="w-4 h-4 mr-2" />
                Export Config
              </Button>
              <Button variant="outline" onClick={handleImportConfig}>
                <Upload className="w-4 h-4 mr-2" />
                Import Config
              </Button>
              <input ref={fileInputRef} type="file" accept=".json" className="hidden" onChange={onFileSelected} />
            </div>
          </Card>
        </div>
      )}

      {activeTab === 'notifications' && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Notification Preferences</CardTitle>
              <CardDescription>Configure alert notifications</CardDescription>
            </CardHeader>
            <div className="space-y-4">
              {[
                { key: 'email_on_alert', label: 'Email on alert', desc: 'Send email when an alert is triggered' },
                { key: 'ssl_expiry_warnings', label: 'SSL expiry warnings', desc: 'Alert 14 days before SSL expires' },
                { key: 'backup_failures', label: 'Backup failures', desc: 'Alert when a backup fails' },
                { key: 'website_down_alerts', label: 'Website down alerts', desc: 'Alert when a website goes offline' },
              ].map((item) => (
                <div key={item.key} className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium">{item.label}</p>
                    <p className="text-xs text-text-secondary">{item.desc}</p>
                  </div>
                  <Switch
                    checked={alertsForm[item.key as keyof typeof alertsForm]}
                    onChange={(e) => setAlertsForm({ ...alertsForm, [item.key]: e.target.checked })}
                  />
                </div>
              ))}
              <Button onClick={saveAlerts} loading={saving}>Save Notification Settings</Button>
            </div>
          </Card>
        </div>
      )}

      <Modal
        open={showRecoveryModal}
        onClose={close2FAModal}
        title="Recovery Codes"
        description="Save these codes in a safe place. You can use them to access your account if you lose your authenticator device."
      >
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-2">
            {twoFARecoveryCodes.map((code, i) => (
              <div key={i} className="p-2 bg-accent/30 rounded text-sm font-mono text-center">
                {code}
              </div>
            ))}
          </div>
          <div className="flex justify-end">
            <Button onClick={close2FAModal}>Done</Button>
          </div>
        </div>
      </Modal>

      <Modal
        open={showImportConfirm}
        onClose={() => { setShowImportConfirm(false); setPendingImportData('') }}
        title="Import Configuration"
        description="This will overwrite your current panel settings. Are you sure you want to proceed?"
      >
        <div className="space-y-4">
          <div className="flex items-center gap-2 p-3 bg-warning/10 border border-warning/20 rounded">
            <AlertTriangle className="w-4 h-4 text-warning" />
            <span className="text-sm text-warning">This action cannot be undone.</span>
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => { setShowImportConfirm(false); setPendingImportData('') }}>
              Cancel
            </Button>
            <Button onClick={confirmImport} loading={importing}>
              Confirm Import
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
