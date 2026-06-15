import { useEffect, useState } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import {
  ArrowLeft, Globe, Lock, Unlock, RefreshCw, ExternalLink, Copy,
  Check, AlertTriangle, Database, Folder, GitBranch, Package, Clock,
  FileText, ChevronDown, Loader2, Trash2, PauseCircle, Play, Pencil,
  Server
} from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge, StatusBadge } from '../../components/ui/Badge'
import { Tabs, TabPanel } from '../../components/ui/Tabs'
import { TableHeader, TableBody, TableRow, TableHead, TableCell } from '../../components/ui/Table'
import { ProgressBar } from '../../components/ui/Misc'
import { CopyButton } from '../../components/ui/Misc'
import { Modal, ConfirmModal } from '../../components/ui/Modal'
import { PageHeader } from '../../components/ui/Misc'
import api from '../../lib/api'
import { formatDate, formatDateTime } from '../../lib/utils'
import { cn } from '../../lib/utils'
import { Switch } from '../../components/ui/Input'
import { useApiError } from '../../hooks/useToast'

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

interface Domain {
  id: number
  domain: string
  type: string
  ssl_enabled: boolean
  ssl_expiry: string | null
  ssl_cert_type: string
  created_at: string
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
  const showError = useApiError()
  const [website, setWebsite] = useState<Website | null>(null)
  const [dnsRecords, setDnsRecords] = useState<DNSRecord[]>([])
  const [dnsZone, setDnsZone] = useState<{id: number; domain: string; serial: number; updated_at: string} | null>(null)
  const [domains, setDomains] = useState<Domain[]>([])
  const [apps, setApps] = useState<App[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState('overview')
  const [sslLoading, setSslLoading] = useState(false)
  const [showSuspendModal, setShowSuspendModal] = useState(false)
  const [suspending, setSuspending] = useState(false)
  const [showRestoreModal, setShowRestoreModal] = useState(false)
  const [restoring, setRestoring] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [editFormData, setEditFormData] = useState({ php_version: '', web_server: '' })
  const [updating, setUpdating] = useState(false)
  const [confirmRemoveSSL, setConfirmRemoveSSL] = useState(false)
  const [confirmDeleteDNS, setConfirmDeleteDNS] = useState<number | null>(null)
  const [serverIP, setServerIP] = useState<string>('')
  const [nameservers, setNameservers] = useState<string[]>([])

  const tabs = [
    { id: 'overview', label: 'Overview' },
    { id: 'domains', label: 'Domains', count: domains.length },
    { id: 'ssl', label: 'SSL · Security' },
    { id: 'dns', label: 'DNS · Domain records' },
    { id: 'apps', label: 'Apps', count: apps.length },
    { id: 'git', label: 'Git Deploy · Auto-update' },
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
      const [siteRes, dnsRes, domainsRes, appsRes, ipRes, nsRes] = await Promise.allSettled([
        api.get(`/websites/${id}`),
        api.get(`/websites/${id}/dns/records`),
        api.get(`/websites/${id}/domains`),
        api.get(`/websites/${id}/apps`),
        api.get('/server/ip'),
        api.get('/dns/nameservers'),
      ])

      if (siteRes.status === 'fulfilled') {
        setWebsite(siteRes.value.data.data.website || siteRes.value.data.data)
      }
      if (dnsRes.status === 'fulfilled') {
        setDnsRecords(dnsRes.value.data.data || [])
        setDnsZone(dnsRes.value.data.zone || null)
      }
      if (domainsRes.status === 'fulfilled') {
        setDomains(domainsRes.value.data.data || [])
      }
      if (appsRes.status === 'fulfilled') {
        setApps(appsRes.value.data.data || [])
      }
      if (ipRes.status === 'fulfilled') {
        setServerIP(ipRes.value.data.data?.ip || '')
      }
      if (nsRes.status === 'fulfilled') {
        const nsData = nsRes.value.data.data
        setNameservers([nsData?.ns1_hostname, nsData?.ns2_hostname].filter(Boolean))
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
      showError(err, 'Failed to issue SSL')
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
      showError(err, 'Failed to renew SSL')
    } finally {
      setSslLoading(false)
    }
  }

  const handleRemoveSSL = () => {
    setConfirmRemoveSSL(true)
  }

  const doRemoveSSL = async () => {
    if (!id) return
    setConfirmRemoveSSL(false)
    setSslLoading(true)
    try {
      await api.delete(`/websites/${id}/ssl`)
      loadData()
    } catch (err: any) {
      showError(err, 'Failed to remove SSL')
    } finally {
      setSslLoading(false)
    }
  }

  const handleUpdateWebsite = async () => {
    if (!id) return
    setUpdating(true)
    try {
      await api.put(`/websites/${id}`, {
        php_version: editFormData.php_version,
        web_server: editFormData.web_server,
      })
      setShowEditModal(false)
      loadData()
    } catch (err: any) {
      showError(err, 'Failed to update website')
    } finally {
      setUpdating(false)
    }
  }

  const openEditModal = () => {
    if (!website) return
    setEditFormData({
      php_version: website.php_version || '8.2',
      web_server: website.web_server || 'nginx',
    })
    setShowEditModal(true)
  }

  const handleSuspend = async () => {
    if (!id) return
    setSuspending(true)
    try {
      await api.post(`/websites/${id}/suspend`)
      loadData()
      setShowSuspendModal(false)
    } catch (err: any) {
      showError(err, 'Failed to suspend website')
    } finally {
      setSuspending(false)
    }
  }

  const handleRestore = async () => {
    if (!id) return
    setRestoring(true)
    try {
      await api.post(`/websites/${id}/restore`)
      loadData()
      setShowRestoreModal(false)
    } catch (err: any) {
      showError(err, 'Failed to restore website')
    } finally {
      setRestoring(false)
    }
  }

  const handleDelete = async () => {
    if (!id) return
    try {
      await api.delete(`/websites/${id}`)
      navigate('/websites')
    } catch (err: any) {
      showError(err, 'Failed to delete')
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
              <Button
                variant="warning"
                size="sm"
                onClick={() => setShowSuspendModal(true)}
              >
                <PauseCircle className="w-4 h-4 mr-1.5" />
                Suspend
              </Button>
            ) : (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowRestoreModal(true)}
              >
                <Play className="w-4 h-4 mr-1.5" />
                Restore
              </Button>
            )}
          </div>
        }
      />

      <Tabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />

      {activeTab === 'overview' && <OverviewTab website={website} onEdit={openEditModal} serverIP={serverIP} nameservers={nameservers} />}
      {activeTab === 'domains' && <DomainsTab websiteId={id!} domains={domains} website={website} onRefresh={loadData} />}
      {activeTab === 'ssl' && <SSLTabs website={website} onIssue={handleIssueSSL} onRenew={handleRenewSSL} onRemove={handleRemoveSSL} loading={sslLoading} />}
      {activeTab === 'dns' && <DNSTab records={dnsRecords} zone={dnsZone} websiteId={id!} onRefresh={loadData} />}
      {activeTab === 'apps' && <AppsTab websiteId={id!} apps={apps} onRefresh={loadData} />}
      {activeTab === 'git' && <GitTab websiteId={id!} />}
      {activeTab === 'logs' && <LogTab websiteId={id!} />}

      <Modal open={showEditModal} onClose={() => setShowEditModal(false)} title="Edit Website">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">PHP Version · The language your code runs on</label>
            <select
              value={editFormData.php_version}
              onChange={(e) => setEditFormData({ ...editFormData, php_version: e.target.value })}
              className="w-full h-10 px-3 bg-background border border-border rounded text-sm"
            >
              <option value="8.1">PHP 8.1</option>
              <option value="8.2">PHP 8.2</option>
              <option value="8.3">PHP 8.3</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Web Server · Software that serves pages</label>
            <select
              value={editFormData.web_server}
              onChange={(e) => setEditFormData({ ...editFormData, web_server: e.target.value })}
              className="w-full h-10 px-3 bg-background border border-border rounded text-sm"
            >
              <option value="nginx">Nginx</option>
              <option value="apache">Apache</option>
            </select>
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="outline" onClick={() => setShowEditModal(false)}>Cancel</Button>
            <Button onClick={handleUpdateWebsite} disabled={updating}>
              {updating ? 'Saving...' : 'Save Changes'}
            </Button>
          </div>
        </div>
      </Modal>

      <ConfirmModal
        open={confirmRemoveSSL}
        onClose={() => setConfirmRemoveSSL(false)}
        onConfirm={doRemoveSSL}
        title="Remove SSL Certificate"
        description="Remove the SSL certificate from this website? The site will revert to HTTP."
        confirmLabel="Remove"
        variant="danger"
        loading={sslLoading}
      />

      <ConfirmModal
        open={showSuspendModal}
        onClose={() => setShowSuspendModal(false)}
        onConfirm={handleSuspend}
        title="Suspend Website"
        description={`Suspend ${website?.domain}? This will take the website offline temporarily.`}
        confirmLabel="Suspend"
        variant="warning"
        loading={suspending}
      />

      <ConfirmModal
        open={showRestoreModal}
        onClose={() => setShowRestoreModal(false)}
        onConfirm={handleRestore}
        title="Restore Website"
        description={`Restore ${website?.domain} and bring it back online?`}
        confirmLabel="Restore"
        variant="primary"
        loading={restoring}
      />
    </div>
  )
}

function OverviewTab({ website, onEdit, serverIP, nameservers }: { website: Website; onEdit: () => void; serverIP: string; nameservers: string[] }) {
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
          <InfoRow label="Document Root · Where files are stored" value={website.document_root} />
          <InfoRow label="Status" value={<StatusBadge status={website.status} />} />
          <InfoRow label="Created" value={formatDate(website.created_at)} />
        </div>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Clock className="w-4 h-4" />
              Runtime
            </CardTitle>
            <Button
              variant="ghost"
              size="icon"
              onClick={onEdit}
              title="Edit"
            >
              <Pencil className="w-4 h-4" />
            </Button>
          </div>
        </CardHeader>
        <div className="space-y-3">
          <InfoRow label="PHP Version · The language your code runs on" value={`PHP ${website.php_version}`} />
          <InfoRow label="Web Server · Serves pages to visitors" value={website.web_server} />
          <InfoRow
            label="SSL · Security certificate"
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
            <InfoRow label="SSL Expiry · When cert needs renewal" value={formatDate(website.ssl_expiry)} />
          )}
        </div>
      </Card>

      {serverIP && (
        <Card className="lg:col-span-2 bg-gradient-to-r from-primary/5 to-transparent border-primary/20">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Server className="w-4 h-4 text-primary" />
              Point Your Domain · How to connect this domain to the server
            </CardTitle>
          </CardHeader>
          <div className="space-y-4 p-1">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-1">
                <p className="text-sm font-medium text-foreground">Option A: A Record</p>
                <p className="text-sm text-text-secondary">Create an A record for @ pointing to:</p>
                <div className="flex items-center gap-2">
                  <code className="px-2 py-1 bg-accent rounded text-sm font-mono">{serverIP}</code>
                  <Button variant="ghost" size="sm" onClick={() => navigator.clipboard.writeText(serverIP)}>
                    <Copy className="w-3.5 h-3.5" />
                  </Button>
                </div>
              </div>
              {nameservers.length > 0 && (
                <div className="space-y-1">
                  <p className="text-sm font-medium text-foreground">Option B: Nameservers</p>
                  <p className="text-sm text-text-secondary">Set your domain's nameservers to:</p>
                  <div className="flex flex-wrap gap-2">
                    {nameservers.map((ns) => (
                      <code key={ns} className="px-2 py-1 bg-accent rounded text-sm font-mono">{ns}</code>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        </Card>
      )}

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
          <CardTitle>SSL Certificate · Enables HTTPS on your site</CardTitle>
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

function DomainsTab({ websiteId, domains, website, onRefresh }: { websiteId: string; domains: Domain[]; website: Website; onRefresh: () => void }) {
  const showError = useApiError()
  const [showAddModal, setShowAddModal] = useState(false)
  const [newDomain, setNewDomain] = useState('')
  const [adding, setAdding] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState<Domain | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [sslLoading, setSslLoading] = useState<number | null>(null)
  const [showSSLModal, setShowSSLModal] = useState<Domain | null>(null)
  const [dnsProvider, setDnsProvider] = useState('cloudflare')
  const [verifying, setVerifying] = useState<number | null>(null)
  const [dnsStatus, setDnsStatus] = useState<Record<number, { all_ok: boolean; records: any[] }>>({})

  const handleAdd = async () => {
    if (!newDomain.trim()) return
    setAdding(true)
    try {
      await api.post(`/websites/${websiteId}/domains`, { domain: newDomain.trim() })
      setNewDomain('')
      setShowAddModal(false)
      onRefresh()
    } catch (err: any) {
      showError(err, 'Failed to add domain')
    } finally {
      setAdding(false)
    }
  }

  const handleDelete = async () => {
    if (!confirmDelete) return
    setDeleting(true)
    try {
      await api.delete(`/websites/${websiteId}/domains/${confirmDelete.id}`)
      setConfirmDelete(null)
      onRefresh()
    } catch (err: any) {
      showError(err, 'Failed to remove domain')
    } finally {
      setDeleting(false)
    }
  }

  const handleIssueSSL = async (domain: Domain) => {
    setSslLoading(domain.id)
    try {
      await api.post(`/websites/${websiteId}/domains/${domain.id}/ssl`, {
        dns_provider: domain.type === 'wildcard' ? dnsProvider : '',
      })
      onRefresh()
    } catch (err: any) {
      showError(err, 'Failed to issue SSL')
    } finally {
      setSslLoading(null)
      setShowSSLModal(null)
    }
  }

  const handleRenewSSL = async (domain: Domain) => {
    setSslLoading(domain.id)
    try {
      await api.post(`/websites/${websiteId}/domains/${domain.id}/ssl-renew`)
      onRefresh()
    } catch (err: any) {
      showError(err, 'Failed to renew SSL')
    } finally {
      setSslLoading(null)
    }
  }

  const handleVerifyDNS = async (domain: Domain) => {
    setVerifying(domain.id)
    try {
      const res = await api.post(`/websites/${websiteId}/domains/${domain.id}/dns-verify`)
      setDnsStatus(prev => ({ ...prev, [domain.id]: res.data.data }))
    } catch (err: any) {
      showError(err, 'Failed to verify DNS')
    } finally {
      setVerifying(null)
    }
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Domain Aliases · Other addresses that point to this site</CardTitle>
              <CardDescription className="mt-1">
                Add extra domains or subdomains that should serve the same content as {website.domain}
              </CardDescription>
            </div>
            <Button size="sm" onClick={() => setShowAddModal(true)}>
              Add Domain
            </Button>
          </div>
        </CardHeader>

        {domains.length === 0 ? (
          <div className="p-8 text-center">
            <Globe className="w-10 h-10 mx-auto text-muted-foreground mb-3" />
            <p className="text-sm text-muted-foreground mb-1">No extra domains yet</p>
            <p className="text-xs text-muted-foreground">
              Add subdomains like blog.{website.domain} or alias domains like www.other.com
            </p>
          </div>
        ) : (
          <div className="p-0">
            <table className="w-full">
              <thead>
                <tr className="border-b border-border">
                  <th className="px-4 py-3 text-left text-sm font-medium text-foreground">Domain</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-foreground">Type</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-foreground">DNS</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-foreground">SSL</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-foreground">Expires</th>
                  <th className="px-4 py-3 w-[140px]"></th>
                </tr>
              </thead>
              <tbody>
                {domains.map((d) => {
                  const status = dnsStatus[d.id]
                  return (
                  <tr key={d.id} className="border-b border-border last:border-0">
                    <td className="px-4 py-3 text-sm font-medium">{d.domain}</td>
                    <td className="px-4 py-3">
                      <Badge variant={d.type === 'wildcard' ? 'warning' : d.type === 'subdomain' ? 'info' : 'neutral'}>
                        {d.type}
                      </Badge>
                    </td>
                    <td className="px-4 py-3">
                      {status ? (
                        status.all_ok ? (
                          <Badge variant="success">Pointed</Badge>
                        ) : (
                          <Badge variant="danger">Not pointed</Badge>
                        )
                      ) : (
                        <Badge variant="neutral">Unknown</Badge>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      {d.ssl_enabled ? (
                        <Badge variant="success">Active</Badge>
                      ) : (
                        <Badge variant="neutral">None</Badge>
                      )}
                    </td>
                    <td className="px-4 py-3 text-sm text-muted-foreground">
                      {d.ssl_expiry ? formatDate(d.ssl_expiry) : '—'}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleVerifyDNS(d)}
                          disabled={verifying === d.id}
                          title="Verify DNS records"
                        >
                          {verifying === d.id ? (
                            <Loader2 className="w-4 h-4 animate-spin" />
                          ) : (
                            <Globe className="w-4 h-4" />
                          )}
                        </Button>
                        {!d.ssl_enabled ? (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setShowSSLModal(d)}
                            disabled={sslLoading === d.id}
                          >
                            <Lock className="w-4 h-4" />
                          </Button>
                        ) : (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRenewSSL(d)}
                            disabled={sslLoading === d.id}
                          >
                            <RefreshCw className="w-4 h-4" />
                          </Button>
                        )}
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setConfirmDelete(d)}
                        >
                          <Trash2 className="w-4 h-4 text-danger" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Quick Guide · How domains work</CardTitle>
        </CardHeader>
        <div className="p-4 space-y-3 text-sm text-muted-foreground">
          <div className="flex items-start gap-2">
            <span className="font-medium text-foreground">Subdomain</span>
            <span>blog.{website.domain} — part of the same domain</span>
          </div>
          <div className="flex items-start gap-2">
            <span className="font-medium text-foreground">Alias</span>
            <span>www.other.com — a completely different domain pointing here</span>
          </div>
          <div className="flex items-start gap-2">
            <span className="font-medium text-foreground">Wildcard</span>
            <span>*.{website.domain} — matches any subdomain automatically</span>
          </div>
          <div className="flex items-start gap-2">
            <span className="font-medium text-foreground">SSL</span>
            <span>Each domain can have its own certificate — wildcards require DNS verification</span>
          </div>
        </div>
      </Card>

      <Modal open={showAddModal} onClose={() => setShowAddModal(false)} title="Add Domain">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">
              Domain · The extra address to point to this site
            </label>
            <input
              type="text"
              value={newDomain}
              onChange={(e) => setNewDomain(e.target.value)}
              placeholder={`blog.${website.domain}`}
              className="w-full h-10 px-3 bg-background border border-border rounded text-sm"
              onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
              autoFocus
            />
          </div>
          <div className="text-xs text-muted-foreground">
            For subdomains: use something like <strong>blog.{website.domain}</strong><br />
            For wildcards: use <strong>*.{website.domain}</strong><br />
            For aliases: use the full domain like <strong>www.other.com</strong>
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="outline" onClick={() => setShowAddModal(false)}>Cancel</Button>
            <Button onClick={handleAdd} disabled={!newDomain.trim() || adding}>
              {adding ? 'Adding...' : 'Add Domain'}
            </Button>
          </div>
        </div>
      </Modal>

      <Modal
        open={!!showSSLModal}
        onClose={() => setShowSSLModal(null)}
        title="Issue SSL Certificate"
      >
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Issue a free Let's Encrypt certificate for <strong>{showSSLModal?.domain}</strong>
          </p>
          {showSSLModal?.type === 'wildcard' && (
            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">
                DNS Provider · Used to verify you own the domain
              </label>
              <select
                value={dnsProvider}
                onChange={(e) => setDnsProvider(e.target.value)}
                className="w-full h-10 px-3 bg-background border border-border rounded text-sm"
              >
                <option value="cloudflare">Cloudflare</option>
                <option value="route53">Amazon Route 53</option>
                <option value="digitalocean">DigitalOcean</option>
                <option value="google">Google Cloud DNS</option>
              </select>
              <p className="text-xs text-muted-foreground mt-1">
                Wildcards require DNS-01 verification — your DNS provider API credentials must be configured on the server
              </p>
            </div>
          )}
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="outline" onClick={() => setShowSSLModal(null)}>Cancel</Button>
            <Button
              onClick={() => showSSLModal && handleIssueSSL(showSSLModal)}
              disabled={sslLoading === showSSLModal?.id}
            >
              {sslLoading === showSSLModal?.id ? 'Issuing...' : 'Issue Certificate'}
            </Button>
          </div>
        </div>
      </Modal>

      <ConfirmModal
        open={!!confirmDelete}
        onClose={() => setConfirmDelete(null)}
        onConfirm={handleDelete}
        title="Remove Domain"
        description={`Are you sure you want to remove ${confirmDelete?.domain}? Nginx will be reloaded and SSL may be affected.`}
        confirmLabel="Remove"
        variant="danger"
        loading={deleting}
      />
    </div>
  )
}

function SSLTabs({ website, onIssue, onRenew, onRemove, loading }: { website: Website; onIssue: () => void; onRenew: () => void; onRemove: () => void; loading: boolean }) {
  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>SSL Certificate · Enables HTTPS on your site</CardTitle>
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
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={onRenew}
                  disabled={loading}
                >
                  <RefreshCw className={cn('w-4 h-4 mr-1.5', loading && 'animate-spin')} />
                  Renew
                </Button>
                <Button
                  variant="danger"
                  size="sm"
                  onClick={onRemove}
                  disabled={loading}
                >
                  <Unlock className="w-4 h-4 mr-1.5" />
                  Remove
                </Button>
              </div>
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
            <Button
              variant="primary"
              onClick={onIssue}
              disabled={loading}
              loading={loading}
            >
              <Lock className="w-4 h-4 mr-1.5" />
              {loading ? 'Issuing...' : 'Issue Certificate'}
            </Button>
          </div>
        )}
      </Card>
    </div>
  )
}

function DNSTab({ records, zone, websiteId, onRefresh }: { records: DNSRecord[]; zone: {id: number; domain: string; serial: number; updated_at: string} | null; websiteId: string; onRefresh: () => void }) {
  const showError = useApiError()
  const [showAddModal, setShowAddModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [editingRecord, setEditingRecord] = useState<DNSRecord | null>(null)
  const [formData, setFormData] = useState({ type: 'A', name: '', value: '', priority: '', ttl: '3600' })
  const [saving, setSaving] = useState(false)
  const [confirmDeleteDNS, setConfirmDeleteDNS] = useState<DNSRecord | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [validationError, setValidationError] = useState('')
  const [syncError, setSyncError] = useState<string | null>(null)

  const validateDNSRecord = (type: string, name: string, value: string): string => {
    if (!value) return 'Value is required'
    if (name && name !== '@' && /\s/.test(name)) return 'Record name cannot contain spaces'
    if (/\n|\r/.test(value)) return 'Value cannot contain newlines'
    switch (type) {
      case 'A': {
        const parts = value.split('.')
        if (parts.length !== 4 || !parts.every(p => { const n = Number(p); return n >= 0 && n <= 255 && p === n.toString() })) {
          return 'Invalid IPv4 address (e.g. 1.2.3.4)'
        }
        break
      }
      case 'AAAA': {
        const parts = value.split(':')
        if (parts.length < 2 || parts.length > 8) return 'Invalid IPv6 address'
        const valid = parts.every(p => p === '' || /^[0-9a-fA-F]{0,4}$/.test(p))
        if (!valid) return 'Invalid IPv6 address (e.g. 2001:db8::1)'
        break
      }
      case 'CNAME':
      case 'NS':
        if (/^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$/.test(value)) return `${type} cannot point to an IP address`
        if (!/^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*\.?$/.test(value)) return `Invalid hostname for ${type}`
        break
      case 'MX':
        if (!/^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*\.?$/.test(value)) return 'Invalid hostname for MX'
        break
      case 'TXT':
        if (value.length > 255) return 'TXT record value too long (max 255 characters)'
        break
      case 'SRV': {
        const parts = value.split(/\s+/)
        if (parts.length !== 4) return 'SRV must have format: priority weight port target'
        break
      }
    }
    return ''
  }

  const handleAdd = async () => {
    const err = validateDNSRecord(formData.type, formData.name, formData.value)
    if (err) { setValidationError(err); return }
    setValidationError('')
    setSaving(true)
    try {
      const payload: any = {
        name: formData.name,
        value: formData.value,
        ttl: parseInt(formData.ttl),
      }
      if (formData.priority) {
        payload.priority = parseInt(formData.priority)
      }
      await api.post(`/websites/${websiteId}/dns/records`, payload)
      setShowAddModal(false)
      setFormData({ type: 'A', name: '', value: '', priority: '', ttl: '3600' })
      onRefresh()
    } catch (err: any) {
      const msg = err?.response?.data?.error?.message || err.message || 'Failed to add record'
      setSyncError(msg)
    } finally {
      setSaving(false)
    }
  }

  const handleEdit = async () => {
    if (!editingRecord) return
    const err = validateDNSRecord(editingRecord.type, formData.name, formData.value)
    if (err) { setValidationError(err); return }
    setValidationError('')
    setSaving(true)
    try {
      const payload: any = {
        name: formData.name,
        value: formData.value,
        ttl: parseInt(formData.ttl),
      }
      if (formData.priority) {
        payload.priority = parseInt(formData.priority)
      }
      await api.put(`/dns/records/${editingRecord.id}`, payload)
      setShowEditModal(false)
      setEditingRecord(null)
      onRefresh()
    } catch (err: any) {
      const msg = err?.response?.data?.error?.message || err.message || 'Failed to update record'
      setSyncError(msg)
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = (record: DNSRecord) => {
    setConfirmDeleteDNS(record)
  }

  const doDeleteDNS = async () => {
    if (!confirmDeleteDNS) return
    const record = confirmDeleteDNS
    setConfirmDeleteDNS(null)
    setDeleting(true)
    try {
      await api.delete(`/dns/records/${record.id}`)
      onRefresh()
    } catch (err: any) {
      const msg = err?.response?.data?.error?.message || err.message || 'Failed to delete record'
      setSyncError(msg)
    } finally {
      setDeleting(false)
    }
  }

  const openEditModal = (record: DNSRecord) => {
    setEditingRecord(record)
    setFormData({
      type: record.type,
      name: record.name,
      value: record.value,
      priority: record.priority?.toString() || '',
      ttl: record.ttl?.toString() || '3600',
    })
    setShowEditModal(true)
  }

  return (
    <>
      <Card padding="none">
        <div className="p-4 border-b border-border flex items-center justify-between">
          <div>
            <h3 className="font-medium">DNS Records · Entries pointing your domain</h3>
            <p className="text-xs text-text-secondary mt-0.5">Manage DNS records for {websiteId}</p>
          </div>
          <Button
            variant="primary"
            size="sm"
            onClick={() => { setSyncError(null); setShowAddModal(true) }}
          >
            <Plus className="w-4 h-4 mr-1.5" />
            Add Record
          </Button>
        </div>
        {zone && (
          <div className="px-4 py-2 bg-accent/50 border-b border-border flex items-center gap-4 text-xs text-text-secondary">
            <span>Zone: <span className="font-mono font-medium text-foreground">{zone.domain}</span></span>
            <span>Serial: <span className="font-mono font-medium text-foreground">{zone.serial}</span></span>
            <span>Last updated: <span className="font-medium text-foreground">{formatDateTime(zone.updated_at)}</span></span>
          </div>
        )}
        {syncError && (
          <div className="mx-4 mt-4 p-3 bg-danger/10 border border-danger/20 rounded flex items-start gap-3">
            <AlertTriangle className="w-4 h-4 text-danger mt-0.5 shrink-0" />
            <div className="flex-1 min-w-0">
              <p className="text-sm text-danger font-medium">Sync failed</p>
              <p className="text-xs text-text-secondary mt-0.5">{syncError}</p>
            </div>
            <Button variant="ghost" size="sm" onClick={() => setSyncError(null)}>
              Dismiss
            </Button>
          </div>
        )}
        {!zone ? (
          <div className="p-8 text-center">
            <Globe className="w-10 h-10 mx-auto text-muted-foreground mb-3" />
            <p className="text-sm font-medium text-foreground mb-1">No DNS zone for this website</p>
            <p className="text-xs text-text-secondary mb-3">A DNS zone is needed before you can add records. Create a website first to set up DNS automatically.</p>
          </div>
        ) : records.length === 0 ? (
          <div className="p-8 text-center">
            <Globe className="w-10 h-10 mx-auto text-muted-foreground mb-3" />
            <p className="text-sm font-medium text-foreground mb-1">No DNS records yet</p>
            <p className="text-xs text-text-secondary mb-3">Add an A record to point your domain to the server IP address.</p>
            <Button variant="outline" size="sm" onClick={() => { setSyncError(null); setShowAddModal(true) }}>
              <Plus className="w-4 h-4 mr-1.5" />
              Add your first record
            </Button>
          </div>
        ) : (
          <table className="w-full">
            <TableHeader>
              <TableRow>
                <TableHead>Type</TableHead>
                <TableHead>Name</TableHead>
                <TableHead>Value</TableHead>
                <TableHead>TTL · Cache time (seconds)</TableHead>
                <TableHead>Priority · Server preference (lower = higher)</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {records.map((rec) => (
                <TableRow key={rec.id}>
                  <TableCell><Badge variant="neutral">{rec.type}</Badge></TableCell>
                  <TableCell className="font-mono">{rec.name}</TableCell>
                  <TableCell className="font-mono text-text-secondary">{rec.value}</TableCell>
                  <TableCell className="text-text-secondary">{rec.ttl}</TableCell>
                  <TableCell className="text-text-secondary">{rec.priority || '—'}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => openEditModal(rec)}
                        title="Edit"
                      >
                        <Pencil className="w-3.5 h-3.5" />
                      </Button>
                      <Button
                        variant="danger"
                        size="icon"
                        onClick={() => handleDelete(rec)}
                        title="Delete"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </table>
        )}
      </Card>

      <Modal open={showAddModal} onClose={() => setShowAddModal(false)} title="Add DNS Record">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Type</label>
            <select
              value={formData.type}
              onChange={(e) => setFormData({ ...formData, type: e.target.value })}
              className="w-full h-10 px-3 bg-background border border-border rounded text-sm"
            >
              <option value="A">A · Points to IP address</option>
              <option value="AAAA">AAAA · Points to IPv6</option>
              <option value="CNAME">CNAME · Points to another domain</option>
              <option value="MX">MX · Email delivery server</option>
              <option value="TXT">TXT · Text for verification</option>
              <option value="NS">NS · Nameserver for domain</option>
              <option value="CAA">CAA · Certificate authority restriction</option>
              <option value="SRV">SRV · Service host and port</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Name</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="@ or subdomain"
              className="w-full h-10 px-3 bg-background border border-border rounded text-sm"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Value</label>
            <input
              type="text"
              value={formData.value}
              onChange={(e) => setFormData({ ...formData, value: e.target.value })}
              placeholder="IP address or hostname"
              className="w-full h-10 px-3 bg-background border border-border rounded text-sm"
            />
          </div>
          {(formData.type === 'MX' || formData.type === 'SRV') && (
            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">Priority</label>
              <input
                type="number"
                value={formData.priority}
                onChange={(e) => setFormData({ ...formData, priority: e.target.value })}
                placeholder="10"
                className="w-full h-10 px-3 bg-background border border-border rounded text-sm"
              />
            </div>
          )}
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">TTL · Cache time (seconds)</label>
            <select
              value={formData.ttl}
              onChange={(e) => setFormData({ ...formData, ttl: e.target.value })}
              className="w-full h-10 px-3 bg-background border border-border rounded text-sm"
            >
              <option value="300">5 minutes</option>
              <option value="900">15 minutes</option>
              <option value="1800">30 minutes</option>
              <option value="3600">1 hour</option>
              <option value="7200">2 hours</option>
              <option value="14400">4 hours</option>
              <option value="86400">1 day</option>
            </select>
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="outline" onClick={() => { setShowAddModal(false); setValidationError('') }}>Cancel</Button>
            <Button onClick={handleAdd} disabled={saving || !formData.name || !formData.value}>
              {saving ? 'Adding...' : 'Add Record'}
            </Button>
          </div>
          {validationError && <p className="text-sm text-red-500 mt-2">{validationError}</p>}
        </div>
      </Modal>

      <Modal open={showEditModal} onClose={() => setShowEditModal(false)} title="Edit DNS Record">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Type</label>
            <input
              type="text"
              value={editingRecord?.type || ''}
              disabled
              className="w-full h-10 px-3 bg-accent/50 border border-border rounded text-sm"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Name</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="w-full h-10 px-3 bg-background border border-border rounded text-sm"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Value</label>
            <input
              type="text"
              value={formData.value}
              onChange={(e) => setFormData({ ...formData, value: e.target.value })}
              className="w-full h-10 px-3 bg-background border border-border rounded text-sm"
            />
          </div>
          {(formData.type === 'MX' || formData.type === 'SRV') && (
            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">Priority</label>
              <input
                type="number"
                value={formData.priority}
                onChange={(e) => setFormData({ ...formData, priority: e.target.value })}
                className="w-full h-10 px-3 bg-background border border-border rounded text-sm"
              />
            </div>
          )}
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">TTL · Cache time (seconds)</label>
            <select
              value={formData.ttl}
              onChange={(e) => setFormData({ ...formData, ttl: e.target.value })}
              className="w-full h-10 px-3 bg-background border border-border rounded text-sm"
            >
              <option value="300">5 minutes</option>
              <option value="900">15 minutes</option>
              <option value="1800">30 minutes</option>
              <option value="3600">1 hour</option>
              <option value="7200">2 hours</option>
              <option value="14400">4 hours</option>
              <option value="86400">1 day</option>
            </select>
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="outline" onClick={() => { setShowEditModal(false); setValidationError('') }}>Cancel</Button>
            <Button onClick={handleEdit} disabled={saving || !formData.value}>
              {saving ? 'Saving...' : 'Save Changes'}
            </Button>
          </div>
          {validationError && <p className="text-sm text-red-500 mt-2">{validationError}</p>}
        </div>
      </Modal>

      <ConfirmModal
        open={!!confirmDeleteDNS}
        onClose={() => setConfirmDeleteDNS(null)}
        onConfirm={doDeleteDNS}
        title="Delete DNS Record"
        description={`Permanently delete this ${confirmDeleteDNS?.type} record? This cannot be undone.`}
        confirmLabel="Delete"
        variant="danger"
        loading={deleting}
      />
    </>
  )
}

function AppsTab({ websiteId, apps, onRefresh }: { websiteId: string; apps: App[]; onRefresh: () => void }) {
  const showError = useApiError()
  const [showInstallModal, setShowInstallModal] = useState(false)
  const [appType, setAppType] = useState('wordpress')
  const [installing, setInstalling] = useState(false)
  const [updating, setUpdating] = useState<number | null>(null)

  const handleInstall = async () => {
    setInstalling(true)
    try {
      await api.post(`/websites/${websiteId}/apps`, { app_type: appType })
      setShowInstallModal(false)
      onRefresh()
    } catch (err: any) {
      showError(err, 'Failed to install app')
    } finally {
      setInstalling(false)
    }
  }

  const handleUpdate = async (app: App) => {
    setUpdating(app.id)
    try {
      await api.post(`/websites/${websiteId}/apps/${app.id}/update`)
      onRefresh()
    } catch (err: any) {
      showError(err, 'Failed to update app')
    } finally {
      setUpdating(null)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="font-medium">Installed Applications</h3>
          <p className="text-xs text-text-secondary mt-0.5">One-click apps installed on this website</p>
        </div>
        <Button
          variant="primary"
          onClick={() => setShowInstallModal(true)}
        >
          <Package className="w-4 h-4 mr-1.5" />
          Install App
        </Button>
      </div>

      {apps.length === 0 ? (
        <Card className="text-center py-8">
          <Package className="w-12 h-12 mx-auto text-text-secondary/50 mb-3" />
          <p className="text-sm text-text-secondary mb-3">No apps installed</p>
          <Button
            variant="ghost"
            onClick={() => setShowInstallModal(true)}
          >
            Install your first app
          </Button>
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
              {app.update_available ? (
                <Button
                  variant="warning"
                  size="sm"
                  onClick={() => handleUpdate(app)}
                  disabled={updating === app.id}
                  loading={updating === app.id}
                >
                  <RefreshCw className={cn('w-3.5 h-3.5 mr-1', updating === app.id && 'animate-spin')} />
                  {updating === app.id ? 'Updating...' : 'Update'}
                </Button>
              ) : (
                <Badge variant="success">Current</Badge>
              )}
            </Card>
          ))}
        </div>
      )}

      <Modal open={showInstallModal} onClose={() => setShowInstallModal(false)} title="Install Application" size="sm">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Application Type</label>
            <select
              value={appType}
              onChange={(e) => setAppType(e.target.value)}
              className="w-full h-10 px-3 bg-background border border-border rounded text-sm"
            >
              <option value="wordpress">WordPress</option>
              <option value="woocommerce">WooCommerce</option>
              <option value="ghost">Ghost</option>
              <option value="nextcloud">Nextcloud</option>
              <option value="drupal">Drupal</option>
              <option value="joomla">Joomla</option>
            </select>
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="outline" onClick={() => setShowInstallModal(false)}>Cancel</Button>
            <Button onClick={handleInstall} disabled={installing}>
              {installing ? 'Installing...' : 'Install'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}

function GitTab({ websiteId }: { websiteId: string }) {
  const showError = useApiError()
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
      showError(err, 'Failed to save')
    } finally {
      setSaving(false)
    }
  }

  const handlePull = async () => {
    setPulling(true)
    try {
      await api.post(`/websites/${websiteId}/git/pull`)
    } catch (err: any) {
      showError(err, 'Failed to pull')
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
            Git Deployment · Auto-update from repository
          </CardTitle>
        </CardHeader>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1.5">Repository URL · Your code's web address</label>
            <input
              type="text"
              value={formData.repo_url}
              onChange={(e) => setFormData({ ...formData, repo_url: e.target.value })}
              placeholder="https://github.com/user/repo.git"
              className="w-full h-10 px-3 bg-background border border-border rounded text-sm font-mono focus:outline-none focus:ring-2 focus:ring-primary/50"
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1.5">Branch · Version to deploy</label>
            <input
              type="text"
              value={formData.branch}
              onChange={(e) => setFormData({ ...formData, branch: e.target.value })}
              placeholder="main"
              className="w-full h-10 px-3 bg-background border border-border rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
            />
          </div>
          <Switch
            checked={formData.auto_deploy}
            onChange={(e) => setFormData({ ...formData, auto_deploy: e.target.checked })}
            label="Auto-deploy on push · Update site when code is saved"
          />
          <div className="flex items-center gap-2">
            <Button
              variant="primary"
              onClick={handleSave}
              disabled={saving || !formData.repo_url}
              loading={saving}
            >
              {saving ? 'Saving...' : 'Save Configuration'}
            </Button>
            {gitConfig && (
              <Button
                variant="outline"
                size="sm"
                onClick={handlePull}
                disabled={pulling}
                loading={pulling}
              >
                <RefreshCw className={cn('w-4 h-4 mr-1.5', pulling && 'animate-spin')} />
                {pulling ? 'Pulling...' : 'Pull Now'}
              </Button>
            )}
          </div>
        </div>
      </Card>

      {webhookURL && (
        <Card>
          <CardHeader>
            <CardTitle>Webhook URL · Receives notifications from repository</CardTitle>
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
