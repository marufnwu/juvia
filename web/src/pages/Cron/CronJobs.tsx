import { useEffect, useState } from 'react'
import { Clock, Plus, Trash2, Play, Pause, RefreshCw, CheckCircle, XCircle, Terminal, Pencil, FileText } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { Table } from '../../components/ui/Table'
import { PageHeader } from '../../components/ui/Misc'
import { Modal, ConfirmModal } from '../../components/ui/Modal'
import { Input, Label, Select, FormGroup, Switch } from '../../components/ui/Input'
import api from '../../lib/api'
import { formatDate, timeAgo } from '../../lib/utils'
import { cn } from '../../lib/utils'
import { useApiError } from '../../hooks/useToast'

interface CronJob {
  id: number
  name: string
  command: string
  schedule: string
  type: 'shell' | 'url' | 'php'
  status: 'active' | 'paused'
  last_run: string | null
  last_status: 'success' | 'failed' | null
  next_run: string
}

export default function CronJobs() {
  const showError = useApiError()
  const [jobs, setJobs] = useState<CronJob[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [formData, setFormData] = useState({
    name: '', command: '', schedule: '0 * * * *', type: 'shell', frequency: 'hourly', run_as: 'www-data',
  })
  const [saving, setSaving] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [editingJob, setEditingJob] = useState<CronJob | null>(null)
  const [editFormData, setEditFormData] = useState({ name: '', command: '', schedule: '', type: 'shell' as 'shell' | 'url' | 'php', frequency: 'hourly' })
  const [showLogsModal, setShowLogsModal] = useState(false)
  const [logsJob, setLogsJob] = useState<CronJob | null>(null)
  const [logs, setLogs] = useState<string[]>([])
  const [logsLoading, setLogsLoading] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState<number | null>(null)
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    loadJobs()
  }, [])

  const loadJobs = async () => {
    try {
      const res = await api.get('/cron')
      setJobs(res.data.data || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = async () => {
    if (!formData.name || !formData.command) return
    setSaving(true)
    try {
      await api.post('/cron', {
        name: formData.name,
        command: formData.command,
        schedule: formData.schedule,
        type: formData.type,
        run_as: formData.run_as,
      })
      setShowModal(false)
      setFormData({ name: '', command: '', schedule: '0 * * * *', type: 'shell', frequency: 'hourly', run_as: 'www-data' })
      loadJobs()
    } catch (err: any) {
      showError(err, 'Failed to create job')
    } finally {
      setSaving(false)
    }
  }

  const handleToggle = async (id: number, enabled: boolean) => {
    const action = enabled ? 'disable' : 'enable'
    try {
      await api.post(`/cron/${id}/${action}`)
      loadJobs()
    } catch (err: any) {
      showError(err, 'Failed to update job')
    }
  }

  const handleDelete = (id: number) => {
    setConfirmDelete(id)
  }

  const doDelete = async () => {
    if (!confirmDelete) return
    const id = confirmDelete
    setConfirmDelete(null)
    setDeleting(true)
    try {
      await api.delete(`/cron/${id}`)
      setJobs(jobs.filter((j) => j.id !== id))
    } catch (err: any) {
      showError(err, 'Failed to delete')
    } finally {
      setDeleting(false)
    }
  }

  const openEditModal = (job: CronJob) => {
    setEditingJob(job)
    setEditFormData({
      name: job.name,
      command: job.command,
      schedule: job.schedule,
      type: job.type,
      frequency: 'hourly',
    })
    setShowEditModal(true)
  }

  const handleEdit = async () => {
    if (!editingJob || !editFormData.name || !editFormData.command) return
    setSaving(true)
    try {
      await api.put(`/cron/${editingJob.id}`, {
        name: editFormData.name,
        command: editFormData.command,
        schedule: editFormData.schedule,
        type: editFormData.type,
      })
      setShowEditModal(false)
      setEditingJob(null)
      loadJobs()
    } catch (err: any) {
      showError(err, 'Failed to update job')
    } finally {
      setSaving(false)
    }
  }

  const viewLogs = async (job: CronJob) => {
    setLogsJob(job)
    setShowLogsModal(true)
    setLogsLoading(true)
    setLogs([])
    try {
      const res = await api.get(`/cron/${job.id}/logs`)
      setLogs(res.data.data || [])
    } catch (err: any) {
      showError(err, 'Failed to load logs')
    } finally {
      setLogsLoading(false)
    }
  }

  const columns = [
    {
      key: 'name',
      header: 'Job',
      render: (job: CronJob) => (
        <div className="flex items-center gap-3">
          <div className={cn('w-8 h-8 rounded flex items-center justify-center',
            job.status === 'active' ? 'bg-success/10' : 'bg-border')}>
            <Terminal className="w-4 h-4 text-text-secondary" />
          </div>
          <div>
            <p className="font-medium text-sm">{job.name}</p>
            <p className="text-xs font-mono text-text-secondary max-w-xs truncate">{job.command}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'schedule',
      header: 'Schedule',
      render: (job: CronJob) => (
        <div>
          <span className="text-sm font-mono">{job.schedule}</span>
          <p className="text-xs text-text-secondary">{job.type}</p>
        </div>
      ),
    },
    {
      key: 'last_run',
      header: 'Last Run',
      render: (job: CronJob) => (
        <div className="flex items-center gap-2">
          {job.last_status === 'success' && <CheckCircle className="w-3.5 h-3.5 text-success" />}
          {job.last_status === 'failed' && <XCircle className="w-3.5 h-3.5 text-danger" />}
          <span className="text-xs text-text-secondary">
            {job.last_run ? timeAgo(job.last_run) : 'Never'}
          </span>
        </div>
      ),
    },
    {
      key: 'next_run',
      header: 'Next Run',
      render: (job: CronJob) => (
        <span className="text-xs text-text-secondary">
          {job.next_run ? formatDate(job.next_run) : '—'}
        </span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (job: CronJob) => (
        <Switch
          checked={job.status === 'active'}
          onChange={() => handleToggle(job.id, job.status === 'active')}
        />
      ),
    },
    {
      key: 'actions',
      header: '',
      width: '120px',
      render: (job: CronJob) => (
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => viewLogs(job)}
            title="View Logs"
          >
            <FileText className="w-3.5 h-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => openEditModal(job)}
            title="Edit"
          >
            <Pencil className="w-3.5 h-3.5" />
          </Button>
          <Button
            variant="danger"
            size="icon"
            onClick={() => handleDelete(job.id)}
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
        title="Cron Jobs"
        description="Schedule automated tasks to run on your server"
        breadcrumbs={[{ label: 'Cron Jobs' }]}
        actions={
          <Button onClick={() => setShowModal(true)}>
            <Plus className="w-4 h-4 mr-2" />
            Add Job
          </Button>
        }
      />

      <Card padding="none">
        <div className="p-4 flex items-center justify-between border-b border-border">
          <div className="flex items-center gap-3">
            <Clock className="w-5 h-5 text-text-secondary" />
            <div>
              <p className="font-medium">{jobs.filter(j => j.status === 'active').length} active jobs</p>
              <p className="text-xs text-text-secondary">Automated task scheduler</p>
            </div>
          </div>
        </div>
        <Table
          columns={columns}
          data={jobs}
          keyField="id"
          loading={loading}
          emptyMessage={
            <div className="p-8 text-center">
              <Clock className="w-12 h-12 mx-auto text-text-secondary/50 mb-3" />
              <p className="text-sm text-text-secondary mb-3">No cron jobs configured</p>
              <Button variant="primary" size="sm" onClick={() => setShowModal(true)}>Create your first cron job</Button>
            </div>
          }
        />
      </Card>

      <Modal
        open={showModal}
        onClose={() => setShowModal(false)}
        title="Add Cron Job"
        size="md"
      >
        <div className="space-y-4">
          <FormGroup>
            <Label>Job Name</Label>
            <Input
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="Daily backup"
            />
          </FormGroup>
          <FormGroup>
            <Label>Command</Label>
            <Input
              value={formData.command}
              onChange={(e) => setFormData({ ...formData, command: e.target.value })}
              placeholder="/usr/local/bin/backup.sh"
            />
          </FormGroup>
          <FormGroup>
            <Label>Frequency</Label>
            <Select
              value={formData.frequency}
              onChange={(e) => {
                const freq = e.target.value
                setFormData({ ...formData, frequency: freq })
                const presets: Record<string, string> = {
                  hourly: '0 * * * *',
                  daily: '0 0 * * *',
                  weekly: '0 0 * * 0',
                  monthly: '0 0 1 * *',
                }
                if (presets[freq]) setFormData({ ...formData, schedule: presets[freq] })
              }}
            >
              <option value="hourly">Every hour</option>
              <option value="daily">Daily at midnight</option>
              <option value="weekly">Weekly on Sunday</option>
              <option value="monthly">Monthly on 1st</option>
            </Select>
          </FormGroup>
          <FormGroup>
                <Label>Cron Expression · When to run (5-part code)</Label>
            <Input
              value={formData.schedule}
              onChange={(e) => setFormData({ ...formData, schedule: e.target.value })}
              placeholder="0 3 * * *"
            />
          </FormGroup>
          <FormGroup>
            <Label>Run As User · Linux account for the task</Label>
            <Input
              value={formData.run_as}
              onChange={(e) => setFormData({ ...formData, run_as: e.target.value })}
              placeholder="www-data"
            />
            <p className="text-xs text-text-secondary mt-1">The Linux user to run the cron job as</p>
          </FormGroup>
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="outline" onClick={() => setShowModal(false)}>Cancel</Button>
            <Button onClick={handleCreate} loading={saving}>Create Job</Button>
          </div>
        </div>
      </Modal>

      <Modal open={showEditModal} onClose={() => setShowEditModal(false)} title="Edit Cron Job" size="md">
        <div className="space-y-4">
          {editingJob && (
            <>
              <FormGroup>
                <Label>Job Name</Label>
                <Input
                  value={editFormData.name}
                  onChange={(e) => setEditFormData({ ...editFormData, name: e.target.value })}
                  placeholder="Daily backup"
                />
              </FormGroup>
              <FormGroup>
                <Label>Command</Label>
                <Input
                  value={editFormData.command}
                  onChange={(e) => setEditFormData({ ...editFormData, command: e.target.value })}
                  placeholder="/usr/local/bin/backup.sh"
                />
              </FormGroup>
              <FormGroup>
                <Label>Frequency</Label>
                <Select
                  value={editFormData.frequency}
                  onChange={(e) => {
                    const freq = e.target.value
                    setEditFormData({ ...editFormData, frequency: freq })
                    const presets: Record<string, string> = {
                      hourly: '0 * * * *',
                      daily: '0 0 * * *',
                      weekly: '0 0 * * 0',
                      monthly: '0 0 1 * *',
                    }
                    if (presets[freq]) setEditFormData({ ...editFormData, schedule: presets[freq] })
                  }}
                >
                  <option value="hourly">Every hour</option>
                  <option value="daily">Daily at midnight</option>
                  <option value="weekly">Weekly on Sunday</option>
                  <option value="monthly">Monthly on 1st</option>
                </Select>
              </FormGroup>
              <FormGroup>
            <Label>Cron Expression · When to run (5-part code)</Label>
                <Input
                  value={editFormData.schedule}
                  onChange={(e) => setEditFormData({ ...editFormData, schedule: e.target.value })}
                  placeholder="0 3 * * *"
                />
              </FormGroup>
              <div className="flex justify-end gap-3 pt-2">
                <Button variant="outline" onClick={() => setShowEditModal(false)}>Cancel</Button>
                <Button onClick={handleEdit} loading={saving}>Save Changes</Button>
              </div>
            </>
          )}
        </div>
      </Modal>

      <Modal open={showLogsModal} onClose={() => setShowLogsModal(false)} title={logsJob ? `Logs: ${logsJob.name}` : 'Cron Logs'} size="lg">
        <div className="space-y-3">
          {logsLoading ? (
            <div className="flex items-center justify-center p-8">
              <RefreshCw className="w-5 h-5 animate-spin text-text-secondary" />
            </div>
          ) : logs.length === 0 ? (
            <div className="p-8 text-center text-text-secondary text-sm">No logs available</div>
          ) : (
            <div className="bg-accent/50 rounded p-3 font-mono text-xs max-h-80 overflow-auto space-y-1">
              {logs.map((line, i) => (
                <div key={i} className="text-text-secondary">{line}</div>
              ))}
            </div>
          )}
          <div className="flex justify-end pt-2">
            <Button variant="outline" onClick={() => setShowLogsModal(false)}>Close</Button>
          </div>
        </div>
      </Modal>

      <ConfirmModal
        open={!!confirmDelete}
        onClose={() => setConfirmDelete(null)}
        onConfirm={doDelete}
        title="Delete Cron Job"
        description="Delete this scheduled task? It will no longer run."
        confirmLabel="Delete"
        variant="danger"
        loading={deleting}
      />
    </div>
  )
}
