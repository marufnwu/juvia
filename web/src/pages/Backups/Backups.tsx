import { useEffect, useState } from 'react'
import { Archive, Plus, Download, Trash2, RefreshCw, CheckCircle, AlertTriangle, Clock, Server } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { Table } from '../../components/ui/Table'
import { PageHeader } from '../../components/ui/Misc'
import { Modal } from '../../components/ui/Modal'
import { Input, Label, Select, FormGroup } from '../../components/ui/Input'
import api from '../../lib/api'
import { formatDate, formatBytes } from '../../lib/utils'
import { cn } from '../../lib/utils'

interface Backup {
  id: number
  name: string
  type: 'full' | 'database' | 'files'
  size: number
  status: 'completed' | 'failed' | 'running'
  created_at: string
  verified: boolean
}

interface BackupSchedule {
  id: number
  name: string
  frequency: string
  retention: number
  enabled: boolean
  last_run: string
  next_run: string
}

export default function Backups() {
  const [backups, setBackups] = useState<Backup[]>([])
  const [schedules, setSchedules] = useState<BackupSchedule[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<'backups' | 'schedules'>('backups')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [selectedBackup, setSelectedBackup] = useState<Backup | null>(null)
  const [showRestoreModal, setShowRestoreModal] = useState(false)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    try {
      const [backupRes, scheduleRes] = await Promise.allSettled([
        api.get('/backups'),
        api.get('/backup-schedules'),
      ])
      if (backupRes.status === 'fulfilled') setBackups(backupRes.value.data.data || [])
      if (scheduleRes.status === 'fulfilled') setSchedules(scheduleRes.value.data.data || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = async () => {
    try {
      await api.post('/backups', { type: 'full' })
      loadData()
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to create backup')
    }
  }

  const handleRestore = async (id: number) => {
    try {
      await api.post(`/backups/${id}/restore`)
      setShowRestoreModal(false)
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to restore')
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this backup?')) return
    try {
      await api.delete(`/backups/${id}`)
      setBackups(backups.filter((b) => b.id !== id))
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to delete')
    }
  }

  const backupColumns = [
    {
      key: 'name',
      header: 'Backup',
      render: (backup: Backup) => (
        <div className="flex items-center gap-3">
          <div className={cn('w-8 h-8 rounded flex items-center justify-center',
            backup.status === 'completed' ? 'bg-success/10' :
            backup.status === 'failed' ? 'bg-danger/10' : 'bg-primary/10')}>
            <Archive className="w-4 h-4 text-text-secondary" />
          </div>
          <div>
            <p className="font-medium text-sm">{backup.name}</p>
            <p className="text-xs text-text-secondary">{formatDate(backup.created_at)}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'type',
      header: 'Type',
      render: (backup: Backup) => (
        <Badge variant={backup.type === 'full' ? 'success' : backup.type === 'database' ? 'info' : 'neutral'}>
          {backup.type}
        </Badge>
      ),
    },
    {
      key: 'size',
      header: 'Size',
      render: (backup: Backup) => (
        <span className="text-sm text-text-secondary">{formatBytes(backup.size)}</span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (backup: Backup) => (
        <div className="flex items-center gap-2">
          {backup.status === 'completed' && <CheckCircle className="w-4 h-4 text-success" />}
          {backup.status === 'failed' && <AlertTriangle className="w-4 h-4 text-danger" />}
          {backup.status === 'running' && <RefreshCw className="w-4 h-4 text-primary animate-spin" />}
          <span className="text-sm capitalize">{backup.status}</span>
        </div>
      ),
    },
    {
      key: 'verified',
      header: 'Verified',
      render: (backup: Backup) => (
        backup.verified ? (
          <span className="flex items-center gap-1 text-xs text-success">
            <CheckCircle className="w-3.5 h-3.5" /> Verified
          </span>
        ) : (
          <span className="text-xs text-text-secondary">Not verified</span>
        )
      ),
    },
    {
      key: 'actions',
      header: '',
      width: '160px',
      render: (backup: Backup) => (
        <div className="flex items-center gap-2">
          <button
            onClick={() => { setSelectedBackup(backup); setShowRestoreModal(true) }}
            className="px-3 py-1.5 text-xs font-medium border border-border rounded hover:bg-accent transition-colors"
          >
            <RefreshCw className="w-3.5 h-3.5 inline mr-1" />
            Restore
          </button>
          <button
            onClick={() => handleDelete(backup.id)}
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
        title="Backups"
        description="Create and restore server backups"
        breadcrumbs={[{ label: 'Backups' }]}
        actions={
          <Button onClick={handleCreate}>
            <Plus className="w-4 h-4 mr-2" />
            Create Backup
          </Button>
        }
      />

      <div className="flex gap-4 border-b border-border">
        <button
          onClick={() => setActiveTab('backups')}
          className={cn('px-4 py-2.5 text-sm font-medium border-b-2 transition-colors',
            activeTab === 'backups' ? 'border-primary text-primary' : 'border-transparent text-text-secondary hover:text-foreground')}
        >
          Backups
        </button>
        <button
          onClick={() => setActiveTab('schedules')}
          className={cn('px-4 py-2.5 text-sm font-medium border-b-2 transition-colors',
            activeTab === 'schedules' ? 'border-primary text-primary' : 'border-transparent text-text-secondary hover:text-foreground')}
        >
          Schedules
        </button>
      </div>

      {activeTab === 'backups' && (
        <Card padding="none">
          <Table
            columns={backupColumns}
            data={backups}
            keyField="id"
            loading={loading}
            emptyMessage="No backups yet"
          />
        </Card>
      )}

      {activeTab === 'schedules' && (
        <Card padding="none">
          <div className="p-4 flex items-center justify-between border-b border-border">
            <div className="flex items-center gap-3">
              <Clock className="w-5 h-5 text-text-secondary" />
              <div>
                <p className="font-medium">{schedules.filter(s => s.enabled).length} active schedules</p>
                <p className="text-xs text-text-secondary">Automated backups</p>
              </div>
            </div>
            <Button size="sm">
              <Plus className="w-4 h-4 mr-2" />
              Add Schedule
            </Button>
          </div>
          <div className="divide-y divide-border">
            {schedules.length === 0 ? (
              <div className="p-8 text-center text-text-secondary text-sm">No schedules configured</div>
            ) : (
              schedules.map((schedule) => (
                <div key={schedule.id} className="flex items-center justify-between p-4 hover:bg-accent/30 transition-colors">
                  <div>
                    <p className="font-medium">{schedule.name}</p>
                    <p className="text-xs text-text-secondary mt-0.5">
                      {schedule.frequency} · Keep {schedule.retention} copies
                    </p>
                    <p className="text-xs text-text-secondary">
                      Next: {schedule.next_run ? formatDate(schedule.next_run) : 'Not scheduled'}
                    </p>
                  </div>
                  <div className="flex items-center gap-3">
                    <button className={cn('relative w-10 h-5 rounded-full transition-colors',
                      schedule.enabled ? 'bg-success' : 'bg-border')}>
                      <span className={cn('absolute top-0.5 w-4 h-4 bg-white rounded-full transition-transform',
                        schedule.enabled ? 'left-5.5' : 'left-0.5')} />
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </Card>
      )}

      <Modal
        open={showRestoreModal}
        onClose={() => setShowRestoreModal(false)}
        title="Restore Backup"
        size="sm"
      >
        <div className="space-y-4">
          <p className="text-sm text-text-secondary">
            Restore backup <strong>{selectedBackup?.name}</strong>? This will overwrite current data.
          </p>
          <div className="flex justify-end gap-3">
            <Button variant="outline" onClick={() => setShowRestoreModal(false)}>Cancel</Button>
            <Button onClick={() => selectedBackup && handleRestore(selectedBackup.id)}>
              <RefreshCw className="w-4 h-4 mr-2" />
              Restore
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}