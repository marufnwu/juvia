import { useEffect, useState } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import api from '../../lib/api'
import { Copy, Check, RefreshCw, GitBranch, Package } from 'lucide-react'

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
  git_config: string | null
}

interface Domain {
  id: number
  domain: string
  type: string
}

interface DNSRecord {
  id: number
  type: string
  name: string
  value: string
  priority: number | null
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

interface WebsiteDetail {
  website: Website
  domains: Domain[]
}

export default function WebsiteDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [data, setData] = useState<WebsiteDetail | null>(null)
  const [dnsRecords, setDnsRecords] = useState<DNSRecord[]>([])
  const [apps, setApps] = useState<App[]>([])
  const [gitConfig, setGitConfig] = useState<GitConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState('overview')
  const [sslLoading, setSslLoading] = useState(false)
  const [gitSetupData, setGitSetupData] = useState({ repo_url: '', branch: 'main', deploy_key: '', auto_deploy: true })

  useEffect(() => {
    if (!id) return
    api.get(`/websites/${id}`)
      .then((res) => setData(res.data.data))
      .catch((err) => console.error(err))
      .finally(() => setLoading(false))

    api.get(`/websites/${id}/dns/records`)
      .then((res) => setDnsRecords(res.data.data || []))
      .catch((err) => console.error(err))

    api.get(`/websites/${id}/apps`)
      .then((res) => setApps(res.data.data || []))
      .catch(() => {})

    api.get(`/websites/${id}/git`)
      .then((res) => {
        if (res.data.success && res.data.data.repo_url) {
          setGitConfig(res.data.data)
          setGitSetupData({
            repo_url: res.data.data.repo_url || '',
            branch: res.data.data.branch || 'main',
            deploy_key: res.data.data.deploy_key || '',
            auto_deploy: res.data.data.auto_deploy ?? true,
          })
        }
      })
      .catch(() => {})
  }, [id])

  const handleIssueSSL = async () => {
    if (!id) return
    setSslLoading(true)
    try {
      await api.post(`/websites/${id}/ssl`)
      window.location.reload()
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
      window.location.reload()
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to renew SSL')
    } finally {
      setSslLoading(false)
    }
  }

  if (loading) {
    return <div className="p-6 text-muted-foreground">Loading...</div>
  }

  if (!data?.website) {
    return <div className="p-6 text-red-600">Website not found</div>
  }

  const { website } = data

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">{website.domain}</h1>
          <p className="text-sm text-muted-foreground">
            {website.document_root}
          </p>
        </div>
        <div className="flex gap-2">
          <Link
            to="/websites"
            className="border border-border px-3 py-1.5 text-sm hover:bg-muted"
          >
            Back
          </Link>
        </div>
      </div>

      <div className="border-b border-border mb-6">
        <nav className="flex gap-4">
          {['overview', 'ssl', 'dns', 'files', 'apps', 'git'].map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`pb-2 text-sm font-medium capitalize ${
                activeTab === tab
                  ? 'border-b-2 border-primary text-primary'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              {tab}
            </button>
          ))}
        </nav>
      </div>

      {activeTab === 'overview' && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <InfoCard title="Website Details">
            <InfoRow label="Domain" value={website.domain} />
            <InfoRow label="Document Root" value={website.document_root} />
            <InfoRow label="Status" value={website.status} />
            <InfoRow label="Created" value={new Date(website.created_at).toLocaleDateString()} />
          </InfoCard>
          <InfoCard title="Runtime">
            <InfoRow label="PHP Version" value={`PHP ${website.php_version}`} />
            <InfoRow label="Web Server" value={website.web_server} />
          </InfoCard>
        </div>
      )}

      {activeTab === 'ssl' && (
        <div className="space-y-6">
          <div className="border border-border p-4">
            <h3 className="font-medium mb-4">SSL Certificate</h3>
            {website.ssl_enabled ? (
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <span className="text-green-600">Active</span>
                  {website.ssl_expiry && (
                    <span className="text-sm text-muted-foreground">
                      Expires {new Date(website.ssl_expiry).toLocaleDateString()}
                    </span>
                  )}
                </div>
                <button
                  onClick={handleRenewSSL}
                  disabled={sslLoading}
                  className="text-sm text-primary hover:underline disabled:opacity-50"
                >
                  {sslLoading ? 'Renewing...' : 'Renew Certificate'}
                </button>
              </div>
            ) : (
              <div className="space-y-2">
                <p className="text-sm text-muted-foreground">
                  No SSL certificate installed
                </p>
                <button
                  onClick={handleIssueSSL}
                  disabled={sslLoading}
                  className="bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90 disabled:opacity-50"
                >
                  {sslLoading ? 'Issuing...' : 'Issue SSL Certificate'}
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {activeTab === 'dns' && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="font-medium">DNS Records</h3>
            <Link
              to={`/websites/${id}/dns`}
              className="text-sm text-primary hover:underline"
            >
              Manage Records
            </Link>
          </div>
          {dnsRecords.length === 0 ? (
            <p className="text-sm text-muted-foreground">No DNS records</p>
          ) : (
            <div className="border border-border">
              <table className="w-full text-sm">
                <thead className="bg-muted/50 border-b border-border">
                  <tr>
                    <th className="text-left p-2 font-medium">Type</th>
                    <th className="text-left p-2 font-medium">Name</th>
                    <th className="text-left p-2 font-medium">Value</th>
                    <th className="text-left p-2 font-medium">Priority</th>
                  </tr>
                </thead>
                <tbody>
                  {dnsRecords.map((rec) => (
                    <tr key={rec.id} className="border-b border-border">
                      <td className="p-2">{rec.type}</td>
                      <td className="p-2">{rec.name}</td>
                      <td className="p-2">{rec.value}</td>
                      <td className="p-2">{rec.priority || '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {activeTab === 'files' && (
        <div className="space-y-4">
          <button
            onClick={() => navigate(`/files/${id}`)}
            className="bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90"
          >
            Open File Manager
          </button>
        </div>
      )}

      {activeTab === 'apps' && (
        <AppsTab
          websiteId={id!}
          apps={apps}
          onRefresh={() => {
            api.get(`/websites/${id}/apps`).then((res) => setApps(res.data.data || [])).catch(() => {})
          }}
        />
      )}

      {activeTab === 'git' && (
        <GitTab
          websiteId={id!}
          gitConfig={gitConfig}
          gitSetupData={gitSetupData}
          setGitSetupData={setGitSetupData}
          onRefresh={() => {
            api.get(`/websites/${id}/git`).then((res) => {
              if (res.data.success && res.data.data.repo_url) {
                setGitConfig(res.data.data)
              }
            }).catch(() => {})
          }}
        />
      )}
    </div>
  )
}

function InfoCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="border border-border p-4">
      <h3 className="font-medium mb-3">{title}</h3>
      <div className="space-y-2">{children}</div>
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between text-sm">
      <span className="text-muted-foreground">{label}</span>
      <span>{value}</span>
    </div>
  )
}

function AppsTab({ websiteId, apps, onRefresh }: { websiteId: string; apps: App[]; onRefresh: () => void }) {
  const [showModal, setShowModal] = useState(false)
  const [selectedApp, setSelectedApp] = useState('')
  const [installing, setInstalling] = useState(false)
  const [adminUsername, setAdminUsername] = useState('admin')
  const [adminPassword, setAdminPassword] = useState('')
  const [dbName, setDbName] = useState('')
  const navigate = useNavigate()

  const appTypes = [
    { id: 'wordpress', name: 'WordPress', description: 'Popular blog and website platform', version: 'Latest' },
    { id: 'laravel', name: 'Laravel', description: 'Modern PHP framework for web applications', version: 'Latest' },
  ]

  const handleInstall = async () => {
    setInstalling(true)
    try {
      await api.post(`/websites/${websiteId}/apps/install`, {
        app_type: selectedApp,
        admin_username: adminUsername,
        admin_password: adminPassword,
        db_name: dbName,
      })
      setShowModal(false)
      onRefresh()
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to install app')
    } finally {
      setInstalling(false)
    }
  }

  const handleUpdate = async (appId: number) => {
    try {
      await api.post(`/websites/${websiteId}/apps/${appId}/update`)
      onRefresh()
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to update app')
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">Installed Applications</h3>
        <button
          onClick={() => setShowModal(true)}
          className="bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90"
        >
          Install App
        </button>
      </div>

      {apps.length === 0 ? (
        <div className="border rounded-lg p-8 text-center">
          <Package className="mx-auto h-12 w-12 text-muted-foreground/50" />
          <h3 className="mt-4 font-medium">No apps installed</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Install a one-click app to get started quickly
          </p>
        </div>
      ) : (
        <div className="border rounded-lg">
          {apps.map((app) => (
            <div key={app.id} className="p-4 border-b last:border-b-0 flex items-center justify-between">
              <div>
                <p className="font-medium">{app.name}</p>
                <p className="text-sm text-muted-foreground">
                  {app.app_type} · v{app.version}
                  {app.update_available && (
                    <span className="ml-2 text-yellow-500">Update available</span>
                  )}
                </p>
              </div>
              <div className="flex gap-2">
                {app.app_type === 'wordpress' && (
                  <button
                    onClick={() => navigate(`/files/${websiteId}`)}
                    className="text-sm text-primary hover:underline"
                  >
                    Manage Files
                  </button>
                )}
                {app.update_available && (
                  <button
                    onClick={() => handleUpdate(app.id)}
                    className="text-sm text-primary hover:underline"
                  >
                    Update
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-background border rounded-lg w-full max-w-md p-6">
            <h2 className="text-lg font-semibold mb-4">Install Application</h2>
            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium">Application</label>
                <select
                  value={selectedApp}
                  onChange={(e) => setSelectedApp(e.target.value)}
                  className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
                >
                  <option value="">Select an app...</option>
                  {appTypes.map((app) => (
                    <option key={app.id} value={app.id}>{app.name}</option>
                  ))}
                </select>
                {selectedApp && (
                  <p className="text-xs text-muted-foreground mt-1">
                    {appTypes.find((a) => a.id === selectedApp)?.description}
                  </p>
                )}
              </div>
              <div>
                <label className="text-sm font-medium">Admin Username</label>
                <input
                  type="text"
                  value={adminUsername}
                  onChange={(e) => setAdminUsername(e.target.value)}
                  className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
                />
              </div>
              <div>
                <label className="text-sm font-medium">Admin Password</label>
                <input
                  type="password"
                  value={adminPassword}
                  onChange={(e) => setAdminPassword(e.target.value)}
                  placeholder="Leave blank to auto-generate"
                  className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
                />
              </div>
              <div>
                <label className="text-sm font-medium">Database Name</label>
                <input
                  type="text"
                  value={dbName}
                  onChange={(e) => setDbName(e.target.value)}
                  placeholder="Leave blank for auto-generated name"
                  className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
                />
              </div>
              <div className="flex justify-end gap-3 pt-4">
                <button
                  onClick={() => setShowModal(false)}
                  className="px-4 py-2 text-sm border rounded-lg hover:bg-muted"
                >
                  Cancel
                </button>
                <button
                  onClick={handleInstall}
                  disabled={!selectedApp || installing}
                  className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-lg hover:opacity-90 disabled:opacity-50"
                >
                  {installing ? 'Installing...' : 'Install'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function GitTab({
  websiteId,
  gitConfig,
  gitSetupData,
  setGitSetupData,
  onRefresh,
}: {
  websiteId: string
  gitConfig: GitConfig | null
  gitSetupData: { repo_url: string; branch: string; deploy_key: string; auto_deploy: boolean }
  setGitSetupData: (data: any) => void
  onRefresh: () => void
}) {
  const [saving, setSaving] = useState(false)
  const [pulling, setPulling] = useState(false)
  const [webhookCopied, setWebhookCopied] = useState(false)

  const handleSave = async () => {
    setSaving(true)
    try {
      await api.post(`/websites/${websiteId}/git`, gitSetupData)
      onRefresh()
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to save git config')
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

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">Git Deployment</h3>
        {gitConfig && (
          <button
            onClick={handlePull}
            disabled={pulling}
            className="inline-flex items-center gap-2 px-3 py-1.5 text-sm border rounded-lg hover:bg-muted disabled:opacity-50"
          >
            <RefreshCw size={14} className={pulling ? 'animate-spin' : ''} />
            {pulling ? 'Pulling...' : 'Pull Now'}
          </button>
        )}
      </div>

      <div className="border rounded-lg p-4 space-y-4">
        <div>
          <label className="text-sm font-medium">Repository URL</label>
          <input
            type="text"
            value={gitSetupData.repo_url}
            onChange={(e) => setGitSetupData({ ...gitSetupData, repo_url: e.target.value })}
            placeholder="https://github.com/user/repo.git"
            className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
          />
          <p className="text-xs text-muted-foreground mt-1">
            The Git repository URL to deploy from
          </p>
        </div>

        <div>
          <label className="text-sm font-medium">Branch</label>
          <input
            type="text"
            value={gitSetupData.branch}
            onChange={(e) => setGitSetupData({ ...gitSetupData, branch: e.target.value })}
            placeholder="main"
            className="w-full mt-1 px-3 py-2 border rounded-lg bg-background"
          />
        </div>

        <div>
          <label className="text-sm font-medium">Deploy Key (optional)</label>
          <textarea
            value={gitSetupData.deploy_key}
            onChange={(e) => setGitSetupData({ ...gitSetupData, deploy_key: e.target.value })}
            placeholder="Paste private key for private repos"
            className="w-full mt-1 px-3 py-2 border rounded-lg bg-background font-mono text-sm"
            rows={3}
          />
          <p className="text-xs text-muted-foreground mt-1">
            SSH private key for private repositories
          </p>
        </div>

        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="auto_deploy"
            checked={gitSetupData.auto_deploy}
            onChange={(e) => setGitSetupData({ ...gitSetupData, auto_deploy: e.target.checked })}
          />
          <label htmlFor="auto_deploy" className="text-sm font-medium">
            Auto-deploy on webhook trigger
          </label>
        </div>

        <button
          onClick={handleSave}
          disabled={saving || !gitSetupData.repo_url}
          className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 disabled:opacity-50"
        >
          {saving ? 'Saving...' : 'Save Configuration'}
        </button>
      </div>

      {gitConfig?.webhook_token && (
        <div className="border rounded-lg p-4">
          <h4 className="font-medium mb-2 flex items-center gap-2">
            <GitBranch size={16} />
            Webhook URL
          </h4>
          <p className="text-xs text-muted-foreground mb-2">
            Copy this URL and add it as a webhook in your Git repository settings to trigger automatic deployments
          </p>
          <div className="flex items-center gap-2">
            <input
              type="text"
              value={webhookURL}
              readOnly
              className="flex-1 px-3 py-2 border rounded-lg bg-muted text-sm font-mono"
            />
            <button
              onClick={() => {
                navigator.clipboard.writeText(webhookURL)
                setWebhookCopied(true)
                setTimeout(() => setWebhookCopied(false), 2000)
              }}
              className="px-3 py-2 border rounded-lg hover:bg-muted"
            >
              {webhookCopied ? <Check size={16} /> : <Copy size={16} />}
            </button>
          </div>
        </div>
      )}

      {gitConfig && (
        <div className="border rounded-lg p-4">
          <h4 className="font-medium mb-2">Current Configuration</h4>
          <div className="text-sm space-y-1 text-muted-foreground">
            <p>Repository: {gitConfig.repo_url}</p>
            <p>Branch: {gitConfig.branch}</p>
            <p>Auto-deploy: {gitConfig.auto_deploy ? 'Enabled' : 'Disabled'}</p>
          </div>
        </div>
      )}
    </div>
  )
}
