import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  Plus, Globe, Search, Filter, MoreHorizontal, Trash2, PauseCircle,
  CheckCircle, XCircle, ExternalLink, RefreshCw, ChevronDown, Copy, Loader2
} from 'lucide-react'
import { Card } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge, StatusBadge } from '../../components/ui/Badge'
import { Table } from '../../components/ui/Table'
import { SearchInput } from '../../components/ui/SearchInput'
import { Select } from '../../components/ui/Input'
import { PageHeader } from '../../components/ui/Misc'
import { CopyButton } from '../../components/ui/Misc'
import { Modal, ConfirmModal } from '../../components/ui/Modal'
import api from '../../lib/api'
import { formatDate } from '../../lib/utils'
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
  traffic?: number
}

export default function WebsiteList() {
  const navigate = useNavigate()
  const showError = useApiError()
  const [websites, setWebsites] = useState<Website[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [selectedWebsite, setSelectedWebsite] = useState<Website | null>(null)
  const [showMenu, setShowMenu] = useState<number | null>(null)
  const [actionLoading, setActionLoading] = useState<number | null>(null)
  const [confirmDelete, setConfirmDelete] = useState<number | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [dnsCheck, setDnsCheck] = useState<Record<string, { ok: boolean; a_record: string }>>({})
  const [checkingDns, setCheckingDns] = useState(false)

  useEffect(() => {
    loadWebsites()
  }, [])

  const loadWebsites = async () => {
    try {
      const res = await api.get('/websites')
      setWebsites(res.data.data || [])
    } catch (err: any) {
      setError(err.response?.data?.error?.user_message || 'Failed to load websites')
    } finally {
      setLoading(false)
    }
  }

  const handleSuspend = async (id: number) => {
    setActionLoading(id)
    try {
      await api.post(`/websites/${id}/suspend`)
      loadWebsites()
    } catch (err: any) {
      showError(err, 'Failed to suspend')
    } finally {
      setActionLoading(null)
      setShowMenu(null)
    }
  }

  const handleRestore = async (id: number) => {
    setActionLoading(id)
    try {
      await api.post(`/websites/${id}/restore`)
      loadWebsites()
    } catch (err: any) {
      showError(err, 'Failed to restore')
    } finally {
      setActionLoading(null)
      setShowMenu(null)
    }
  }

  const handleDelete = async (id: number) => {
    setConfirmDelete(id)
  }

  const doDelete = async () => {
    if (!confirmDelete) return
    const id = confirmDelete
    setConfirmDelete(null)
    setDeleting(true)
    try {
      await api.delete(`/websites/${id}`)
      setWebsites(websites.filter((w) => w.id !== id))
    } catch (err: any) {
      showError(err, 'Failed to delete')
    } finally {
      setDeleting(false)
      setShowMenu(null)
    }
  }

  const handleCheckDNS = async () => {
    setCheckingDns(true)
    try {
      const res = await api.get('/dns/check-all')
      const map: Record<string, { ok: boolean; a_record: string }> = {}
      for (const item of res.data.data || []) {
        map[item.domain] = { ok: item.ok, a_record: item.a_record || '' }
      }
      setDnsCheck(map)
    } catch (err: any) {
      showError(err, 'Failed to check DNS')
    } finally {
      setCheckingDns(false)
    }
  }

  const filteredWebsites = websites.filter((site) => {
    const matchesSearch = site.domain.toLowerCase().includes(search.toLowerCase())
    const matchesStatus = !statusFilter || site.status === statusFilter
    return matchesSearch && matchesStatus
  })

  const columns = [
    {
      key: 'domain',
      header: 'Domain',
      sortable: true,
      render: (site: Website) => (
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded bg-accent/50 flex items-center justify-center">
            <Globe className="w-4 h-4 text-text-secondary" />
          </div>
          <div>
            <p className="font-medium text-foreground">{site.domain}</p>
            <p className="text-xs text-text-secondary font-mono">{site.document_root}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'runtime',
      header: 'Runtime',
      render: (site: Website) => (
        <div>
          <p className="text-sm">PHP {site.php_version}</p>
          <p className="text-xs text-text-secondary capitalize">{site.web_server}</p>
        </div>
      ),
    },
    {
      key: 'ssl',
      header: 'SSL',
      render: (site: Website) => (
        site.ssl_enabled ? (
          <div className="flex items-center gap-1.5">
            <CheckCircle className="w-3.5 h-3.5 text-success" />
            <span className="text-xs text-success">Active</span>
            {site.ssl_expiry && (
              <span className="text-xs text-text-secondary">
                ({new Date(site.ssl_expiry) < new Date(Date.now() + 14 * 86400000) ? '⚠️ ' : ''}
                {formatDate(site.ssl_expiry)})
              </span>
            )}
          </div>
        ) : (
          <span className="text-xs text-text-secondary">Not configured</span>
        )
      ),
    },
    {
      key: 'dns',
      header: 'DNS',
      render: (site: Website) => {
        const dns = dnsCheck[site.domain]
        if (!dns && Object.keys(dnsCheck).length === 0) {
          return <span className="text-xs text-text-secondary">—</span>
        }
        if (!dns) {
          return <span className="text-xs text-text-secondary">Not in DNS</span>
        }
        return dns.ok ? (
          <div className="flex items-center gap-1.5">
            <CheckCircle className="w-3.5 h-3.5 text-success" />
            <span className="text-xs text-success">Pointed</span>
          </div>
        ) : (
          <div className="flex items-center gap-1.5">
            <XCircle className="w-3.5 h-3.5 text-danger" />
            <span className="text-xs text-danger">Not pointed</span>
            {dns.a_record && (
              <span className="text-xs text-text-secondary">→ {dns.a_record}</span>
            )}
          </div>
        )
      },
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (site: Website) => <StatusBadge status={site.status} />,
    },
    {
      key: 'created',
      header: 'Created',
      sortable: true,
      render: (site: Website) => (
        <span className="text-xs text-text-secondary">{formatDate(site.created_at)}</span>
      ),
    },
    {
      key: 'actions',
      header: '',
      width: '48px',
      render: (site: Website) => (
        <div className="relative">
          <Button
            variant="ghost"
            size="sm"
            onClick={(e) => {
              e.stopPropagation()
              setShowMenu(showMenu === site.id ? null : site.id)
              setSelectedWebsite(site)
            }}
          >
            <MoreHorizontal className="w-4 h-4" />
          </Button>
          {showMenu === site.id && (
            <>
              <div
                className="fixed inset-0 z-10"
                onClick={() => setShowMenu(null)}
              />
              <div className="absolute right-0 top-full mt-1 w-48 bg-surface border border-border rounded-card shadow-xl z-20 py-1">
                <Link
                  to={`/websites/${site.id}`}
                  className="flex items-center gap-2 px-3 py-2 text-sm hover:bg-accent/50 transition-colors"
                  onClick={() => setShowMenu(null)}
                >
                  <ExternalLink className="w-3.5 h-3.5" />
                  View Details
                </Link>
                {site.status === 'active' && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleSuspend(site.id)}
                    disabled={actionLoading === site.id}
                    loading={actionLoading === site.id}
                    className="w-full justify-start text-warning"
                  >
                    <PauseCircle className="w-3.5 h-3.5 mr-2" />
                    Suspend
                  </Button>
                )}
                {site.status === 'suspended' && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleRestore(site.id)}
                    disabled={actionLoading === site.id}
                    loading={actionLoading === site.id}
                    className="w-full justify-start text-success"
                  >
                    <RefreshCw className="w-3.5 h-3.5 mr-2" />
                    Restore
                  </Button>
                )}
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    navigator.clipboard.writeText(site.domain)
                    setShowMenu(null)
                  }}
                  className="w-full justify-start"
                >
                  <Copy className="w-3.5 h-3.5 mr-2" />
                  Copy Domain
                </Button>
                <hr className="my-1 border-border" />
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleDelete(site.id)}
                  className="w-full justify-start text-danger hover:text-danger"
                >
                  <Trash2 className="w-3.5 h-3.5 mr-2" />
                  Move to Trash
                </Button>
              </div>
            </>
          )}
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <PageHeader
        title="Websites"
        description="Manage your websites and applications"
        breadcrumbs={[{ label: 'Websites' }]}
        actions={
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={handleCheckDNS} disabled={checkingDns}>
              {checkingDns ? (
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <Globe className="w-4 h-4 mr-2" />
              )}
              Check DNS
            </Button>
            <Link
              to="/websites/create"
              className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-white text-sm font-medium rounded hover:bg-primary/90 transition-colors"
            >
              <Plus className="w-4 h-4" />
              Create Website
            </Link>
          </div>
        }
      />

      <Card padding="none">
        <div className="p-4 flex items-center gap-4 border-b border-border">
          <div className="w-64">
            <SearchInput
              value={search}
              onChange={setSearch}
              placeholder="Search domains..."
            />
          </div>
          <Select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="w-48"
          >
            <option value="">All Status</option>
            <option value="active">Active</option>
            <option value="suspended">Suspended</option>
          </Select>
          <div className="flex-1" />
          <Link
            to="/websites/trash"
            className="flex items-center gap-2 px-3 py-2 text-sm text-text-secondary hover:text-foreground hover:bg-accent rounded transition-colors"
          >
            <Trash2 className="w-4 h-4" />
            Trash
          </Link>
        </div>

        <Table
          columns={columns}
          data={filteredWebsites}
          keyField="id"
          onRowClick={(site) => navigate(`/websites/${site.id}`)}
          loading={loading}
          emptyMessage={
            <div className="p-8 text-center">
              <Globe className="w-12 h-12 mx-auto text-text-secondary/50 mb-3" />
              <p className="text-sm text-text-secondary mb-3">No websites yet</p>
              <Button variant="primary" size="sm" onClick={() => navigate('/websites/create')}>Create your first website</Button>
            </div>
          }
        />
      </Card>

      <ConfirmModal
        open={!!confirmDelete}
        onClose={() => setConfirmDelete(null)}
        onConfirm={doDelete}
        title="Move Website to Trash"
        description={confirmDelete ? `Move ${websites.find((w) => w.id === confirmDelete)?.domain || 'this website'} to trash? You can restore it later.` : ''}
        confirmLabel="Move to Trash"
        variant="danger"
        loading={deleting}
      />
    </div>
  )
}