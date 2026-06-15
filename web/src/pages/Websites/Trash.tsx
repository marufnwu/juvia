import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Trash2, RotateCcw, AlertTriangle, Clock, Globe } from 'lucide-react'
import { Button } from '../../components/ui/Button'
import { Card } from '../../components/ui/Card'
import { Badge, StatusBadge } from '../../components/ui/Badge'
import { Table } from '../../components/ui/Table'
import { PageHeader } from '../../components/ui/Misc'
import { Modal, ConfirmModal } from '../../components/ui/Modal'
import api from '../../lib/api'
import { formatDate, timeAgo } from '../../lib/utils'
import { useApiError } from '../../hooks/useToast'

interface TrashedWebsite {
  id: number
  domain: string
  deleted_at: string
  days_remaining: number
  status: string
}

export default function Trash() {
  const showError = useApiError()
  const [websites, setWebsites] = useState<TrashedWebsite[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedSite, setSelectedSite] = useState<TrashedWebsite | null>(null)
  const [showRestoreModal, setShowRestoreModal] = useState(false)
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [actionLoading, setActionLoading] = useState(false)
  const [confirmPurgeAll, setConfirmPurgeAll] = useState(false)
  const [purging, setPurging] = useState(false)

  useEffect(() => {
    loadTrash()
  }, [])

  const loadTrash = async () => {
    try {
      const res = await api.get('/websites/trash')
      setWebsites(res.data.data || [])
    } catch (err) {
      console.error('Failed to load trash:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleRestore = async (id: number) => {
    setActionLoading(true)
    try {
      await api.post(`/websites/${id}/restore`)
      loadTrash()
      setShowRestoreModal(false)
    } catch (err: any) {
      showError(err, 'Failed to restore')
    } finally {
      setActionLoading(false)
    }
  }

  const handlePermanentDelete = async (id: number) => {
    try {
      await api.delete(`/websites/${id}/permanent`)
      setWebsites(websites.filter((w) => w.id !== id))
      setShowDeleteModal(false)
    } catch (err: any) {
      showError(err, 'Failed to delete')
    }
  }

  const handlePurgeAll = () => {
    setConfirmPurgeAll(true)
  }

  const doPurgeAll = async () => {
    setConfirmPurgeAll(false)
    setPurging(true)
    try {
      await api.delete('/websites/trash/purge')
      setWebsites([])
    } catch (err: any) {
      showError(err, 'Failed to purge')
    } finally {
      setPurging(false)
    }
  }

  const columns = [
    {
      key: 'domain',
      header: 'Domain',
      render: (site: TrashedWebsite) => (
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded bg-accent/50 flex items-center justify-center">
            <Globe className="w-4 h-4 text-text-secondary" />
          </div>
          <div>
            <p className="font-medium">{site.domain}</p>
            <p className="text-xs text-text-secondary">Deleted {timeAgo(site.deleted_at)}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (site: TrashedWebsite) => <StatusBadge status={site.status} />,
    },
    {
      key: 'days_remaining',
      header: 'Days Until Permanent Delete',
      render: (site: TrashedWebsite) => (
        <div className="flex items-center gap-2">
          <Clock className="w-3.5 h-3.5 text-text-secondary" />
          <span className="text-sm">{site.days_remaining} days</span>
        </div>
      ),
    },
    {
      key: 'actions',
      header: '',
      width: '160px',
      render: (site: TrashedWebsite) => (
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={(e) => {
              e.stopPropagation()
              setSelectedSite(site)
              setShowRestoreModal(true)
            }}
          >
            <RotateCcw className="w-3.5 h-3.5 mr-1" />
            Restore
          </Button>
          <Button
            variant="danger"
            size="sm"
            onClick={(e) => {
              e.stopPropagation()
              setSelectedSite(site)
              setShowDeleteModal(true)
            }}
          >
            Delete
          </Button>
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <PageHeader
        title="Trash"
        description="Deleted websites are kept for 30 days before permanent deletion"
        breadcrumbs={[
          { label: 'Websites', href: '/websites' },
          { label: 'Trash' },
        ]}
        actions={
          websites.length > 0 && (
            <Button variant="danger" onClick={handlePurgeAll}>
              <Trash2 className="w-4 h-4 mr-2" />
              Purge All
            </Button>
          )
        }
      />

      <Card padding="none">
        <Table
          columns={columns}
          data={websites}
          keyField="id"
          loading={loading}
          emptyMessage="Trash is empty"
        />
      </Card>

      <ConfirmModal
        open={showRestoreModal}
        onClose={() => setShowRestoreModal(false)}
        onConfirm={() => selectedSite && handleRestore(selectedSite.id)}
        title="Restore Website"
        description={`Restore ${selectedSite?.domain} from trash?`}
        confirmLabel="Restore"
        variant="primary"
        loading={actionLoading}
      />

      <ConfirmModal
        open={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onConfirm={() => selectedSite && handlePermanentDelete(selectedSite.id)}
        title="Permanently Delete"
        description={`This will permanently delete ${selectedSite?.domain}. This action cannot be undone.`}
        confirmLabel="Delete Forever"
        variant="danger"
        loading={actionLoading}
      />

      <ConfirmModal
        open={confirmPurgeAll}
        onClose={() => setConfirmPurgeAll(false)}
        onConfirm={doPurgeAll}
        title="Empty Trash"
        description="Permanently delete all trashed websites? This cannot be undone."
        confirmLabel="Empty Trash"
        variant="danger"
        loading={purging}
      />
    </div>
  )
}