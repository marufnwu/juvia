import { useEffect, useState } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import {
  ArrowLeft, Globe, Lock, Unlock, RefreshCw, ExternalLink, Copy,
  Check, AlertTriangle, Database, Folder, GitBranch, Package, Clock,
  FileText, ChevronDown, Loader2, Trash2, PauseCircle, Play
} from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge, StatusBadge } from '../../components/ui/Badge'
import { Tabs, TabPanel } from '../../components/ui/Tabs'
import { ProgressBar } from '../../components/ui/Misc'
import { CopyButton } from '../../components/ui/Misc'
import { Modal } from '../../components/ui/Modal'
import { PageHeader } from '../../components/ui/Misc'
import api from '../../lib/api'
import { formatDate, formatDateTime } from '../../lib/utils'
import { cn } from '../../lib/utils'

interface Website {
  id: number
  domain: string
  document_root: string
  php_version: string
  web_server: string
  ssl_enabled: boolean
  ssl_expiry: string | null
  status: string
  created_at: string
  git_config: any
}

interface DNSRecord {
  id: number
  type: string
  name: string
  value: string
  priority: number | null
  ttl: number
}

interface App {
  id: number
  app_type: string
  name: string
  version: string
  update_available: boolean
}

interface GitConfig {
  repo_url: string
  branch: string
  auto_deploy: boolean
  webhook_token: string
}

export default function WebsiteDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [website, setWebsite] = useState<Website | null>(null)
  const [dnsRecords, setDnsRecords] = useState<DNSRecord[]>([])
  const [apps, setApps] = useState<App[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState('overview')
  const [sslLoading, setSslLoading] = useState(false)
  const [showDeleteModal, setShowDeleteModal] = useState(false)

  const tabs = [
    { id: 'overview', label: 'Overview' },
    { id: 'ssl', label: 'SSL' },
    { id: 'dns', label: 'DNS' },
    { id: 'apps', label: 'Apps', count: apps.length },
    { id: 'git', label: 'Git Deploy' },
    { id: 'logs', label: 'Logs' },
  ]

  useEffect(() => {
    if (!id) return
    loadData()
  }, [id])

  const loadData = async () => {
    if (!id) return
    setLoading(true)
    try {
      const [siteRes, dnsRes, appsRes] = await Promise.allSettled([
        api.get(`/websites/${id}`),
        api.get(`/websites/${id}/dns/records`),
        api.get(`/websites/${id}/apps`),
      ])

      if (siteRes.status === 'fulfilled') {
        setWebsite(siteRes.value.data.data.website || siteRes.value.data.data)
      }
      if (dnsRes.status === 'fulfilled') {
        setDnsRecords(dnsRes.value.data.data || [])
      }
      if (appsRes.status === 'fulfilled') {
        setApps(appsRes.value.data.data || [])
      }
    } catch (err) {
      console.error('Failed to load website:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleIssueSSL = async () => {
    if (!id) return
    setSslLoading(true)
    try {
      await api.post(`/websites/${id}/ssl`)
      loadData()
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to issue SSL')
    } finally {
      setSslLoading(false)
    }
  }

  const handleRenewSSL = async () => {
    if (!id) return
    setSslLoading(true)
    try {
      await api.post(`/websites/${id}/ssl-renew`)
      loadData()
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to renew SSL')
    } finally {
      setSslLoading(false)
    }
  }

  const handleSuspend = async () => {
    if (!id || !website) return
    try {
      await api.post(`/websites/${id}/suspend`)
      loadData()
      setShowDeleteModal(false)
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to suspend')
    }
  }

  const handleDelete = async () => {
    if (!id) return
    try {
      await api.delete(`/websites/${id}`)
      navigate('/websites')
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to delete')
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-6 h-6 animate-spin text-primary" />
      </div>
    )
  }

  if (!website) {
    return (
      <div className="text-center py-12">
        <AlertTriangle className="w-12 h-12 mx-auto text-danger mb-4" />
        <h2 className="text-lg font-semibold mb-2">Website not found</h2>
        <Link to="/websites" className="text-primary hover:underline">Back to websites</Link>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={website.domain}
        description={website.document_root}
        breadcrumbs={[
          { label: 'Websites', href: '/websites' },
          { label: website.domain },
        ]}
        actions={
          <div className="flex items-center gap-2">
            <StatusBadge status={website.status} />
            <Link
              to={`/websites/${id}/logs`}
              className="inline-flex items-center gap-2 px-3 py-2 border border-border text-sm rounded hover:bg-accent transition-colors"
            >
              <FileText className="w-4 h-4" />
              Logs
            </Link>
            {website.status === 'active' ? (
              <button
                onClick={() => setShowDeleteModal(true)}
                className="inline-flex items-center gap-2 px-3 py-2 border border-border text-warning text-sm rounded hover:bg-warning/10 transition-colors"
              >
                <PauseCircle className="w-4 h-4" />
                Suspend
              </button>
            ) : (
              <button
                onClick={handleSuspend}
                className="inline-flex items-center gap-2 px-3 py-2 border border-border text-success text-sm rounded hover:bg-success/10 transition-colors"
              >
                <Play className="w-4 h-4" />
                Restore
              </button>
            )}
          </div>
        }
      />

      <Tabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />

      {activeTab === 'overview' && <OverviewTab website={website} />}
      {activeTab === 'ssl' && <SSLTabs website={website} onIssue={handleIssueSSL} onRenew={handleRenewSSL} loading={sslLoading} />}
      {activeTab === 'dns' && <DNSTab records={dnsRecords} websiteId={id!} />}
      {activeTab === 'apps' && <AppsTab websiteId={id!} apps={apps} onRefresh={loadData} />}
      {activeTab === 'git' && <GitTab websiteId={id!} />}
      {activeTab === 'logs' && <LogTab websiteId={id!} />}
    </div>
  )
}

function OverviewTab({ website }: { website: Website }) {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Globe className="w-4 h-4" />
            Website Details
          </CardTitle>
        </CardHeader>
        <div className="space-y-3">
          <InfoRow label="Domain" value={website.domain} />
          <InfoRow label="Document Root" value={website.document_root} />
          <InfoRow label="Status" value={<StatusBadge status={website.status} />} />
          <InfoRow label="Created" value={formatDate(website.created_at)} />
        </div>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Clock className="w-4 h-4" />
            Runtime
          </CardTitle>
        </CardHeader>
        <div className="space-y-3">
          <InfoRow label="PHP Version" value={`PHP ${website.php_version}`} />
          <InfoRow label="Web Server" value={website.web_server} />
          <InfoRow
            label="SSL"
            value={
              website.ssl_enabled ? (
                <span className="flex items-center gap-1 text-success">
                  <Lock className="w-3.5 h-3.5" /> Active
                </span>
              ) : (
                <span className="flex items-center gap-1 text-text-secondary">
                  <Unlock className="w-3.5 h-3.5" /> Not configured
                </span>
              )
            }
          />
          {website.ssl_expiry && (
            <InfoRow label="SSL Expiry" value={formatDate(website.ssl_expiry)} />
          )}
        </div>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Quick Actions</CardTitle>
        </CardHeader>
        <div className="grid grid-cols-2 gap-2">
          <Link
            to={`/files/${website.id}`}
            className="flex items-center gap-2 p-3 bg-accent/50 rounded hover:bg-accent transition-colors"
          >
            <Folder className="w-4 h-4 text-primary" />
            <span className="text-sm">File Manager</span>
          </Link>
          <Link
            to={`/databases?website=${website.id}`}
            className="flex items-center gap-2 p-3 bg-accent/50 rounded hover:bg-accent transition-colors"
          >
            <Database className="w-4 h-4 text-primary" />
            <span className="text-sm">Databases</span>
          </Link>
        </div>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>SSL Certificate</CardTitle>
        </CardHeader>
        {website.ssl_enabled ? (
          <div className="space-y-3">
            <div className="flex items-center gap-2 text-success">
              <Check className="w-4 h-4" />
              <span className="text-sm font-medium">Certificate Active</span>
            </div>
            {website.ssl_expiry && (
              <p className="text-xs text-text-secondary">
                Expires {formatDate(website.ssl_expiry)}
              </p>
            )}
          </div>
        ) : (
          <div className="space-y-3">
            <p className="text-sm text-text-secondary">No SSL certificate installed</p>
            <Link
              to={`/websites/${website.id}`}
              className="inline-flex items-center gap-2 px-3 py-2 bg-primary text-white text-sm rounded hover:bg-primary/90 transition-colors"
            >
              <Lock className="w-4 h-4" />
              Issue Certificate
            </Link>
          </div>
        )}
      </Card>
    </div>
  )
}

function SSLTabs({ website, onIssue, onRenew, loading }: { website: Website; onIssue: () => void; onRenew: () => void; loading: boolean }) {
  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>SSL Certificate</CardTitle>
        </CardHeader>
        {website.ssl_enabled ? (
          <div className="space-y-4">
            <div className="flex items-center justify-between p-4 bg-success/10 border border-success/20 rounded">
              <div className="flex items-center gap-3">
                <Lock className="w-5 h-5 text-success" />
                <div>
                  <p className="text-sm font-medium text-success">Active</p>
                  {website.ssl_expiry && (
                    <p className="text-xs text-text-secondary">
                      Expires {formatDate(website.ssl_expiry)}
                    </p>
                  )}
                </div>
              </div>
              <button
                onClick={onRenew}
                disabled={loading}
                className="inline-flex items-center gap-2 px-4 py-2 border border-border text-sm rounded hover:bg-accent transition-colors disabled:opacity-50"
              >
                <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
                Renew
              </button>
            </div>
          </div>
        ) : (
          <div className="space-y-4">
            <div className="flex items-center gap-3 p-4 bg-warning/10 border border-warning/20 rounded">
              <Unlock className="w-5 h-5 text-warning" />
              <div className="flex-1">
                <p className="text-sm font-medium">No SSL Certificate</p>
                <p className="text-xs text-text-secondary">
                  Issue a free Let's Encrypt certificate for {website.domain}
                </p>
              </div>
            </div>
            <button
              onClick={onIssue}
              disabled={loading}
              className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-white text-sm font-medium rounded hover:bg-primary/90 disabled:opacity-50 transition-colors"
            >
              <Lock className="w-4 h-4" />
              {loading ? 'Issuing...' : 'Issue Certificate'}
            </button>
          </div>
        )}
      </Card>
    </div>
  )
}

function DNSTab({ records, websiteId }: { records: DNSRecord[]; websiteId: string }) {
  return (
    <Card padding="none">
      <div className="p-4 border-b border-border flex items-center justify-between">
        <div>
          <h3 className="font-medium">DNS Records</h3>
          <p className="text-xs text-text-secondary mt-0.5">Manage DNS records for {websiteId}</p>
        </div>
        <button className="inline-flex items-center gap-2 px-3 py-2 bg-primary text-white text-sm rounded hover:bg-primary/90 transition-colors">
          <Plus className="w-4 h-4" />
          Add Record
        </button>
      </div>
      {records.length === 0 ? (
        <div className="p-8 text-center text-text-secondary text-sm">
          No DNS records found
        </div>
      ) : (
        <table className="w-full">
          <thead className="bg-accent/50">
            <tr>
              <th className="text-left p-3 text-xs font-medium text-text-secondary uppercase">Type</th>
              <th className="text-left p-3 text-xs font-medium text-text-secondary uppercase">Name</th>
              <th className="text-left p-3 text-xs font-medium text-text-secondary uppercase">Value</th>
              <th className="text-left p-3 text-xs font-medium text-text-secondary uppercase">TTL</th>
              <th className="text-left p-3 text-xs font-medium text-text-secondary uppercase">Priority</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {records.map((rec) => (
              <tr key={rec.id} className="hover:bg-accent/30 transition-colors">
                <td className="p-3"><Badge variant="neutral">{rec.type}</Badge></td>
                <td className="p-3 text-sm font-mono">{rec.name}</td>
                <td className="p-3 text-sm font-mono text-text-secondary">{rec.value}</td>
                <td className="p-3 text-sm text-text-secondary">{rec.ttl}</td>
                <td className="p-3 text-sm text-text-secondary">{rec.priority || '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </Card>
  )
}

function AppsTab({ websiteId, apps, onRefresh }: { websiteId: string; apps: App[]; onRefresh: () => void }) {
  const [showInstallModal, setShowInstallModal] = useState(false)

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="font-medium">Installed Applications</h3>
          <p className="text-xs text-text-secondary mt-0.5">One-click apps installed on this website</p>
        </div>
        <button
          onClick={() => setShowInstallModal(true)}
          className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-white text-sm font-medium rounded hover:bg-primary/90 transition-colors"
        >
          <Package className="w-4 h-4" />
          Install App
        </button>
      </div>

      {apps.length === 0 ? (
        <Card className="text-center py-8">
          <Package className="w-12 h-12 mx-auto text-text-secondary/50 mb-3" />
          <p className="text-sm text-text-secondary mb-3">No apps installed</p>
          <button
            onClick={() => setShowInstallModal(true)}
            className="text-sm text-primary hover:underline"
          >
            Install your first app
          </button>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {apps.map((app) => (
            <Card key={app.id} className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded bg-accent/50 flex items-center justify-center">
                  <Package className="w-5 h-5 text-text-secondary" />
                </div>
                <div>
                  <p className="font-medium">{app.name}</p>
                  <p className="text-xs text-text-secondary">{app.app_type} v{app.version}</p>
                </div>
              </div>
              {app.update_available && (
                <Badge variant="warning">Update available</Badge>
              )}
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}

function GitTab({ websiteId }: { websiteId: string }) {
  const [gitConfig, setGitConfig] = useState<GitConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [formData, setFormData] = useState({ repo_url: '', branch: 'main', auto_deploy: true })
  const [saving, setSaving] = useState(false)
  const [pulling, setPulling] = useState(false)

  useEffect(() => {
    loadGitConfig()
  }, [websiteId])

  const loadGitConfig = async () => {
    try {
      const res = await api.get(`/websites/${websiteId}/git`)
      if (res.data.success && res.data.data?.repo_url) {
        setGitConfig(res.data.data)
        setFormData({
          repo_url: res.data.data.repo_url || '',
          branch: res.data.data.branch || 'main',
          auto_deploy: res.data.data.auto_deploy ?? true,
        })
      }
    } catch (err) {
      console.error('Failed to load git config:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      await api.post(`/websites/${websiteId}/git`, formData)
      loadGitConfig()
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to save')
    } finally {
      setSaving(false)
    }
  }

  const handlePull = async () => {
    setPulling(true)
    try {
      await api.post(`/websites/${websiteId}/git/pull`)
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to pull')
    } finally {
      setPulling(false)
    }
  }

  const webhookURL = gitConfig?.webhook_token
    ? `${window.location.protocol}//${window.location.host}/api/v1/webhooks/git/${gitConfig.webhook_token}`
    : ''

  if (loading) {
    return <div className="flex items-center justify-center h-32"><Loader2 className="w-5 h-5 animate-spin" /></div>
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <GitBranch className="w-4 h-4" />
            Git Deployment
          </CardTitle>
        </CardHeader>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1.5">Repository URL</label>
            <input
              type="text"
              value={formData.repo_url}
              onChange={(e) => setFormData({ ...formData, repo_url: e.target.value })}
              placeholder="https://github.com/user/repo.git"
              className="w-full h-10 px-3 bg-background border border-border rounded text-sm font-mono focus:outline-none focus:ring-2 focus:ring-primary/50"
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1.5">Branch</label>
            <input
              type="text"
              value={formData.branch}
              onChange={(e) => setFormData({ ...formData, branch: e.target.value })}
              placeholder="main"
              className="w-full h-10 px-3 bg-background border border-border rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
            />
          </div>
          <div className="flex items-center gap-3">
            <input
              type="checkbox"
              id="auto_deploy"
              checked={formData.auto_deploy}
              onChange={(e) => setFormData({ ...formData, auto_deploy: e.target.checked })}
              className="w-4 h-4 rounded border-border"
            />
            <label htmlFor="auto_deploy" className="text-sm">Auto-deploy on webhook trigger</label>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={handleSave}
              disabled={saving || !formData.repo_url}
              className="px-4 py-2 bg-primary text-white text-sm font-medium rounded hover:bg-primary/90 disabled:opacity-50 transition-colors"
            >
              {saving ? 'Saving...' : 'Save Configuration'}
            </button>
            {gitConfig && (
              <button
                onClick={handlePull}
                disabled={pulling}
                className="inline-flex items-center gap-2 px-4 py-2 border border-border text-sm rounded hover:bg-accent transition-colors disabled:opacity-50"
              >
                <RefreshCw className={cn('w-4 h-4', pulling && 'animate-spin')} />
                {pulling ? 'Pulling...' : 'Pull Now'}
              </button>
            )}
          </div>
        </div>
      </Card>

      {webhookURL && (
        <Card>
          <CardHeader>
            <CardTitle>Webhook URL</CardTitle>
            <CardDescription>Add this URL to your Git repository to trigger auto-deploy</CardDescription>
          </CardHeader>
          <div className="flex items-center gap-2">
            <input
              type="text"
              value={webhookURL}
              readOnly
              className="flex-1 h-10 px-3 bg-accent/50 border border-border rounded text-sm font-mono"
            />
            <CopyButton text={webhookURL} />
          </div>
        </Card>
      )}

      {gitConfig && (
        <Card>
          <CardHeader>
            <CardTitle>Current Configuration</CardTitle>
          </CardHeader>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-text-secondary">Repository</span>
              <span className="font-mono">{gitConfig.repo_url}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-text-secondary">Branch</span>
              <span>{gitConfig.branch}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-text-secondary">Auto-deploy</span>
              <span>{gitConfig.auto_deploy ? 'Enabled' : 'Disabled'}</span>
            </div>
          </div>
        </Card>
      )}
    </div>
  )
}

function LogTab({ websiteId }: { websiteId: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Website Logs</CardTitle>
      </CardHeader>
      <div className="flex gap-4 mb-4">
        <Link
          to={`/websites/${websiteId}/logs`}
          className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-white text-sm rounded hover:bg-primary/90 transition-colors"
        >
          <FileText className="w-4 h-4" />
          View Access & Error Logs
        </Link>
      </div>
      <p className="text-sm text-text-secondary">
        Access and error logs help you debug issues with your website.
      </p>
    </Card>
  )
}

function InfoRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between py-2 border-b border-border last:border-0">
      <span className="text-sm text-text-secondary">{label}</span>
      <span className="text-sm font-medium">{value}</span>
    </div>
  )
}

function Plus(props: any) {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <line x1="12" y1="5" x2="12" y2="19" />
      <line x1="5" y1="12" x2="19" y2="12" />
    </svg>
  )
}