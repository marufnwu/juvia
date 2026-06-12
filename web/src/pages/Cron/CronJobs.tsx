import { useEffect, useState } from 'react'
import { Clock, Plus, Trash2, Play, Pause, RefreshCw, CheckCircle, XCircle, Terminal } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { Table } from '../../components/ui/Table'
import { PageHeader } from '../../components/ui/Misc'
import { Modal } from '../../components/ui/Modal'
import { Input, Label, Select, FormGroup } from '../../components/ui/Input'
import api from '../../lib/api'
import { formatDate, timeAgo } from '../../lib/utils'
import { cn } from '../../lib/utils'

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
  const [jobs, setJobs] = useState<CronJob[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [formData, setFormData] = useState({
    name: '', command: '', schedule: '0 * * * *', type: 'shell', frequency: 'hourly',
  })
  const [saving, setSaving] = useState(false)

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
      })
      setShowModal(false)
      setFormData({ name: '', command: '', schedule: '0 * * * *', type: 'shell', frequency: 'hourly' })
      loadJobs()
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to create job')
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
      alert(err.response?.data?.error?.user_message || 'Failed to update job')
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this cron job?')) return
    try {
      await api.delete(`/cron/${id}`)
      setJobs(jobs.filter((j) => j.id !== id))
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to delete')
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
        <button
          onClick={() => handleToggle(job.id, job.status === 'active')}
          className={cn('relative w-10 h-5 rounded-full transition-colors',
            job.status === 'active' ? 'bg-success' : 'bg-border')}
        >
          <span className={cn('absolute top-0.5 w-4 h-4 bg-white rounded-full transition-transform',
            job.status === 'active' ? 'left-5.5' : 'left-0.5')} />
        </button>
      ),
    },
    {
      key: 'actions',
      header: '',
      width: '80px',
      render: (job: CronJob) => (
        <button
          onClick={() => handleDelete(job.id)}
          className="p-1.5 text-text-secondary hover:text-danger hover:bg-danger/10 rounded transition-colors"
        >
          <Trash2 className="w-3.5 h-3.5" />
        </button>
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
          emptyMessage="No cron jobs configured"
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
            <Label>Cron Expression (advanced)</Label>
            <Input
              value={formData.schedule}
              onChange={(e) => setFormData({ ...formData, schedule: e.target.value })}
              placeholder="0 3 * * *"
            />
          </FormGroup>
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="outline" onClick={() => setShowModal(false)}>Cancel</Button>
            <Button onClick={handleCreate} loading={saving}>Create Job</Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}