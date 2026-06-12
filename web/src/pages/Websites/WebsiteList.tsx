import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  Plus, Globe, Search, Filter, MoreHorizontal, Trash2, PauseCircle,
  CheckCircle, XCircle, ExternalLink, RefreshCw, ChevronDown, Copy
} from 'lucide-react'
import { Card } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge, StatusBadge } from '../../components/ui/Badge'
import { Table } from '../../components/ui/Table'
import { SearchInput } from '../../components/ui/SearchInput'
import { PageHeader } from '../../components/ui/Misc'
import { CopyButton } from '../../components/ui/Misc'
import { Modal, ConfirmModal } from '../../components/ui/Modal'
import api from '../../lib/api'
import { formatDate } from '../../lib/utils'

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
  const [websites, setWebsites] = useState<Website[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [selectedWebsite, setSelectedWebsite] = useState<Website | null>(null)
  const [showMenu, setShowMenu] = useState<number | null>(null)
  const [actionLoading, setActionLoading] = useState<number | null>(null)

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
      alert(err.response?.data?.error?.user_message || 'Failed to suspend')
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
      alert(err.response?.data?.error?.user_message || 'Failed to restore')
    } finally {
      setActionLoading(null)
      setShowMenu(null)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Move this website to trash?')) return
    try {
      await api.delete(`/websites/${id}`)
      setWebsites(websites.filter((w) => w.id !== id))
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to delete')
    }
    setShowMenu(null)
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
          <button
            onClick={(e) => {
              e.stopPropagation()
              setShowMenu(showMenu === site.id ? null : site.id)
              setSelectedWebsite(site)
            }}
            className="p-1.5 text-text-secondary hover:text-foreground hover:bg-accent rounded transition-colors"
          >
            <MoreHorizontal className="w-4 h-4" />
          </button>
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
                  <button
                    onClick={() => handleSuspend(site.id)}
                    disabled={actionLoading === site.id}
                    className="w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-accent/50 transition-colors text-warning"
                  >
                    <PauseCircle className="w-3.5 h-3.5" />
                    {actionLoading === site.id ? 'Suspending...' : 'Suspend'}
                  </button>
                )}
                {site.status === 'suspended' && (
                  <button
                    onClick={() => handleRestore(site.id)}
                    disabled={actionLoading === site.id}
                    className="w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-accent/50 transition-colors text-success"
                  >
                    <RefreshCw className="w-3.5 h-3.5" />
                    {actionLoading === site.id ? 'Restoring...' : 'Restore'}
                  </button>
                )}
                <button
                  onClick={() => {
                    navigator.clipboard.writeText(site.domain)
                    setShowMenu(null)
                  }}
                  className="w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-accent/50 transition-colors"
                >
                  <Copy className="w-3.5 h-3.5" />
                  Copy Domain
                </button>
                <hr className="my-1 border-border" />
                <button
                  onClick={() => handleDelete(site.id)}
                  className="w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-accent/50 transition-colors text-danger"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                  Move to Trash
                </button>
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
          <Link
            to="/websites/create"
            className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-white text-sm font-medium rounded hover:bg-primary/90 transition-colors"
          >
            <Plus className="w-4 h-4" />
            Create Website
          </Link>
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
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="h-9 px-3 bg-surface border border-border rounded text-sm text-foreground"
          >
            <option value="">All Status</option>
            <option value="active">Active</option>
            <option value="suspended">Suspended</option>
          </select>
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
          emptyMessage="No websites found. Create your first website to get started."
        />
      </Card>
    </div>
  )
}