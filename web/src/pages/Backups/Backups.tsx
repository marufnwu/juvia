import { useEffect, useState } from 'react'
import { Archive, Plus, Download, Trash2, RefreshCw, CheckCircle, AlertTriangle, Clock, Server } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { Table } from '../../components/ui/Table'
import { PageHeader } from '../../components/ui/Misc'
import { Modal, ConfirmModal } from '../../components/ui/Modal'
import { Input, Label, Select, FormGroup, Switch } from '../../components/ui/Input'
import api from '../../lib/api'
import { formatDate, formatBytes } from '../../lib/utils'
import { cn } from '../../lib/utils'
import { useApiError } from '../../hooks/useToast'

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
  const showError = useApiError()
  const [backups, setBackups] = useState<Backup[]>([])
  const [schedules, setSchedules] = useState<BackupSchedule[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<'backups' | 'schedules'>('backups')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [selectedBackup, setSelectedBackup] = useState<Backup | null>(null)
  const [showRestoreModal, setShowRestoreModal] = useState(false)
  const [showAddScheduleModal, setShowAddScheduleModal] = useState(false)
  const [scheduleForm, setScheduleForm] = useState({ name: '', frequency: 'daily', retention: 7, storage: 'local' })
  const [creating, setCreating] = useState(false)
  const [toggling, setToggling] = useState<number | null>(null)
  const [backupTypeForm, setBackupTypeForm] = useState({ type: 'full', storage: 'local' })
  const [confirmDeleteBackup, setConfirmDeleteBackup] = useState<number | null>(null)
  const [confirmDeleteSchedule, setConfirmDeleteSchedule] = useState<number | null>(null)
  const [deletingBackup, setDeletingBackup] = useState(false)
  const [deletingSchedule, setDeletingSchedule] = useState(false)

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
      await api.post('/backups', { type: backupTypeForm.type, storage: backupTypeForm.storage })
      setShowCreateModal(false)
      setBackupTypeForm({ type: 'full', storage: 'local' })
      loadData()
    } catch (err: any) {
      showError(err, 'Failed to create backup')
    }
  }

  const handleRestore = async (id: number) => {
    try {
      await api.post(`/backups/${id}/restore`)
      setShowRestoreModal(false)
    } catch (err: any) {
      showError(err, 'Failed to restore')
    }
  }

  const handleDelete = async (id: number) => {
    setConfirmDeleteBackup(id)
  }

  const doDeleteBackup = async () => {
    if (!confirmDeleteBackup) return
    const id = confirmDeleteBackup
    setConfirmDeleteBackup(null)
    setDeletingBackup(true)
    try {
      await api.delete(`/backups/${id}`)
      setBackups(backups.filter((b) => b.id !== id))
    } catch (err: any) {
      showError(err, 'Failed to delete')
    } finally {
      setDeletingBackup(false)
    }
  }

  const handleToggleSchedule = async (id: number, enabled: boolean) => {
    setToggling(id)
    try {
      await api.put(`/backup-schedules/${id}`, { enabled: !enabled })
      loadData()
    } catch (err: any) {
      showError(err, 'Failed to update schedule')
    } finally {
      setToggling(null)
    }
  }

  const handleDeleteSchedule = async (id: number) => {
    setConfirmDeleteSchedule(id)
  }

  const doDeleteSchedule = async () => {
    if (!confirmDeleteSchedule) return
    const id = confirmDeleteSchedule
    setConfirmDeleteSchedule(null)
    setDeletingSchedule(true)
    try {
      await api.delete(`/backup-schedules/${id}`)
      setSchedules(schedules.filter((s) => s.id !== id))
    } catch (err: any) {
      showError(err, 'Failed to delete schedule')
    } finally {
      setDeletingSchedule(false)
    }
  }

  const handleAddSchedule = async () => {
    if (!scheduleForm.name) return
    setCreating(true)
    try {
      const presets: Record<string, string> = {
        hourly: '0 * * * *',
        daily: '0 0 * * *',
        weekly: '0 0 * * 0',
        monthly: '0 0 1 * *',
      }
      await api.post('/backup-schedules', {
        name: scheduleForm.name,
        schedule: presets[scheduleForm.frequency] || scheduleForm.frequency,
        retention_days: scheduleForm.retention,
        storage: scheduleForm.storage || 'local',
      })
      setShowAddScheduleModal(false)
      setScheduleForm({ name: '', frequency: 'daily', retention: 7, storage: 'local' })
      loadData()
    } catch (err: any) {
      showError(err, 'Failed to create schedule')
    } finally {
      setCreating(false)
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
            <p className="text-xs text-text-secondary">{formatBytes(backup.size)} · {backup.type}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (backup: Backup) => (
        <Badge variant={backup.status === 'completed' ? 'success' : backup.status === 'failed' ? 'danger' : 'neutral'}>
          {backup.status}
        </Badge>
      ),
    },
    {
      key: 'created',
      header: 'Created',
      render: (backup: Backup) => (
        <span className="text-xs text-text-secondary">{formatDate(backup.created_at)}</span>
      ),
    },
    {
      key: 'verified',
      header: 'Verified',
      render: (backup: Backup) => (
        backup.verified ? <CheckCircle className="w-4 h-4 text-success" /> : <span className="text-xs text-text-secondary">—</span>
      ),
    },
    {
      key: 'actions',
      header: '',
      width: '140px',
      render: (backup: Backup) => (
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => { setSelectedBackup(backup); setShowRestoreModal(true) }}
            title="Restore"
          >
            <RefreshCw className="w-3.5 h-3.5" />
          </Button>
          <Button
            variant="danger"
            size="icon"
            onClick={() => handleDelete(backup.id)}
            title="Delete"
          >
            <Trash2 className="w-3.5 h-3.5" />
          </Button>
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <PageHeader
        title="Backups"
        description="Backup and restore your server data"
        breadcrumbs={[{ label: 'Backups' }]}
        actions={
          <Button onClick={() => setShowCreateModal(true)}>
            <Plus className="w-4 h-4 mr-2" />
            Create Backup
          </Button>
        }
      />

      <div className="flex gap-4 border-b border-border">
        <Button
          variant="ghost"
          onClick={() => setActiveTab('backups')}
          className={cn('px-4 py-2.5 text-sm font-medium border-b-2 transition-colors',
            activeTab === 'backups' ? 'border-primary text-primary' : 'border-transparent text-text-secondary hover:text-foreground')}
        >
          Backups
        </Button>
        <Button
          variant="ghost"
          onClick={() => setActiveTab('schedules')}
          className={cn('px-4 py-2.5 text-sm font-medium border-b-2 transition-colors',
            activeTab === 'schedules' ? 'border-primary text-primary' : 'border-transparent text-text-secondary hover:text-foreground')}
        >
          Schedules
        </Button>
      </div>

      {activeTab === 'backups' && (
        <Card padding="none">
          <Table
            columns={backupColumns}
            data={backups}
            keyField="id"
            loading={loading}
            emptyMessage={
              <div className="p-8 text-center">
                <Archive className="w-12 h-12 mx-auto text-text-secondary/50 mb-3" />
                <p className="text-sm text-text-secondary mb-3">No backups yet</p>
                <Button variant="primary" size="sm" onClick={() => setShowCreateModal(true)}>Create your first backup</Button>
              </div>
            }
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
            <Button size="sm" onClick={() => setShowAddScheduleModal(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Add Schedule
            </Button>
          </div>
          <div className="divide-y divide-border">
            {schedules.length === 0 ? (
              <div className="p-8 text-center">
                <Clock className="w-12 h-12 mx-auto text-text-secondary/50 mb-3" />
                <p className="text-sm text-text-secondary mb-3">No schedules configured</p>
                <Button variant="primary" size="sm" onClick={() => setShowAddScheduleModal(true)}>Create a backup schedule</Button>
              </div>
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
                    <Switch
                      checked={schedule.enabled}
                      onChange={() => handleToggleSchedule(schedule.id, schedule.enabled)}
                      disabled={toggling === schedule.id}
                    />
                    <Button
                      variant="danger"
                      size="icon"
                      onClick={() => handleDeleteSchedule(schedule.id)}
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </Button>
                  </div>
                </div>
              ))
            )}
          </div>
        </Card>
      )}

      <Modal
        open={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        title="Create Backup"
        size="sm"
      >
        <div className="space-y-4">
          <FormGroup>
            <Label>Backup Type</Label>
            <Select value={backupTypeForm.type} onChange={(e) => setBackupTypeForm({ ...backupTypeForm, type: e.target.value })}>
              <option value="full">Full Backup · Everything including files and database</option>
              <option value="files">Files Only · Website files only</option>
              <option value="database">Database Only · Database contents only</option>
            </Select>
          </FormGroup>
          <FormGroup>
            <Label>Storage Location</Label>
            <Select value={backupTypeForm.storage} onChange={(e) => setBackupTypeForm({ ...backupTypeForm, storage: e.target.value })}>
              <option value="local">Local · Stored on server disk</option>
            </Select>
          </FormGroup>
          <div className="flex justify-end gap-3">
            <Button variant="outline" onClick={() => setShowCreateModal(false)}>Cancel</Button>
            <Button onClick={handleCreate}>Create Backup</Button>
          </div>
        </div>
      </Modal>

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

      <Modal open={showAddScheduleModal} onClose={() => setShowAddScheduleModal(false)} title="Add Backup Schedule" size="sm">
        <div className="space-y-4">
          <FormGroup>
            <Label>Schedule Name</Label>
            <Input
              value={scheduleForm.name}
              onChange={(e) => setScheduleForm({ ...scheduleForm, name: e.target.value })}
              placeholder="Daily backup"
            />
          </FormGroup>
          <FormGroup>
            <Label>Frequency</Label>
            <Select
              value={scheduleForm.frequency}
              onChange={(e) => setScheduleForm({ ...scheduleForm, frequency: e.target.value })}
            >
              <option value="hourly">Every hour</option>
              <option value="daily">Daily</option>
              <option value="weekly">Weekly</option>
              <option value="monthly">Monthly</option>
            </Select>
          </FormGroup>
          <FormGroup>
            <Label>Retention · Backup copies to keep</Label>
            <Input
              type="number"
              value={scheduleForm.retention}
              onChange={(e) => setScheduleForm({ ...scheduleForm, retention: parseInt(e.target.value) || 7 })}
              className="w-32"
            />
          </FormGroup>
          <FormGroup>
            <Label>Storage Location</Label>
            <Select
              value={scheduleForm.storage}
              onChange={(e) => setScheduleForm({ ...scheduleForm, storage: e.target.value })}
            >
              <option value="local">Local · Stored on server disk</option>
            </Select>
          </FormGroup>
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="outline" onClick={() => setShowAddScheduleModal(false)}>Cancel</Button>
            <Button onClick={handleAddSchedule} loading={creating} disabled={!scheduleForm.name}>Create Schedule</Button>
          </div>
        </div>
      </Modal>

      <ConfirmModal
        open={!!confirmDeleteBackup}
        onClose={() => setConfirmDeleteBackup(null)}
        onConfirm={doDeleteBackup}
        title="Delete Backup"
        description="Permanently delete this backup? This cannot be undone."
        confirmLabel="Delete"
        variant="danger"
        loading={deletingBackup}
      />

      <ConfirmModal
        open={!!confirmDeleteSchedule}
        onClose={() => setConfirmDeleteSchedule(null)}
        onConfirm={doDeleteSchedule}
        title="Delete Backup Schedule"
        description="Delete this scheduled backup? Future backups will no longer run."
        confirmLabel="Delete"
        variant="danger"
        loading={deletingSchedule}
      />
    </div>
  )
}
