import { useEffect, useState } from 'react'
import { Clock, Plus, X } from 'lucide-react'
import api from '../../lib/api'

interface CronJob {
  id: number
  schedule: string
  command: string
  run_as: string
  type: string
  enabled: boolean
  last_run: string
  last_status: string
  last_output: string
  created_at: string
}

interface CronJobLog {
  id: number
  cron_job_id: number
  run_at: string
  duration_ms: number
  status: string
  output: string
}

export default function CronJobs() {
  const [jobs, setJobs] = useState<CronJob[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [showLogs, setShowLogs] = useState<CronJob | null>(null)
  const [logs, setLogs] = useState<CronJobLog[]>([])
  const [formData, setFormData] = useState({
    schedule: '0 3 * * *',
    command: '',
    run_as: 'root',
    type: 'shell',
  })

  useEffect(() => {
    loadJobs()
  }, [])

  const loadJobs = async () => {
    setLoading(true)
    try {
      const res = await api.get('/cron')
      setJobs(res.data.data || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const createJob = async () => {
    try {
      await api.post('/cron', formData)
      setShowModal(false)
      setFormData({ schedule: '0 3 * * *', command: '', run_as: 'root', type: 'shell' })
      loadJobs()
    } catch (err) {
      console.error(err)
    }
  }

  const deleteJob = async (id: number) => {
    if (!confirm('Delete this cron job?')) return
    try {
      await api.delete(`/cron/${id}`)
      loadJobs()
    } catch (err) {
      console.error(err)
    }
  }

  const toggleJob = async (id: number, enabled: boolean) => {
    try {
      await api.post(`/cron/${id}/${enabled ? 'enable' : 'disable'}`)
      loadJobs()
    } catch (err) {
      console.error(err)
    }
  }

  const viewLogs = async (job: CronJob) => {
    setShowLogs(job)
    try {
      const res = await api.get(`/cron/${job.id}/logs`)
      setLogs(res.data.data || [])
    } catch (err) {
      console.error(err)
    }
  }

  const describeSchedule = (schedule: string): string => {
    const parts = schedule.split(' ')
    if (parts.length !== 5) return schedule

    const [min, hour, dom, mon, dow] = parts

    if (dom === '*' && mon === '*' && dow === '*') {
      return `Every day at ${hour}:${min.padStart(2, '0')}`
    }
    if (dow === '0' && dom === '*') {
      return `Every Sunday at ${hour}:${min.padStart(2, '0')}`
    }
    if (dom === '1' && mon === '*') {
      return `First day of every month at ${hour}:${min.padStart(2, '0')}`
    }

    return schedule
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Cron Jobs</h1>
          <p className="text-sm text-muted-foreground">
            Cron Job · A task that runs automatically on a schedule
          </p>
        </div>
        <button
          onClick={() => setShowModal(true)}
          className="flex items-center gap-2 bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90"
        >
          <Plus size={16} />
          Create Job
        </button>
      </div>

      {loading ? (
        <div className="text-muted-foreground">Loading...</div>
      ) : jobs.length === 0 ? (
        <div className="text-center py-12 border border-border">
          <Clock size={48} className="mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">No cron jobs configured</p>
        </div>
      ) : (
        <div className="border border-border">
          <table className="w-full text-sm">
            <thead className="bg-muted/50 border-b border-border">
              <tr>
                <th className="text-left p-3 font-medium">Schedule</th>
                <th className="text-left p-3 font-medium">Command</th>
                <th className="text-left p-3 font-medium">Run As</th>
                <th className="text-left p-3 font-medium">Last Run</th>
                <th className="text-left p-3 font-medium">Status</th>
                <th className="text-left p-3 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {jobs.map((job) => (
                <tr key={job.id} className="border-b border-border hover:bg-muted/30">
                  <td className="p-3">
                    <div className="font-mono text-xs">{job.schedule}</div>
                    <div className="text-xs text-muted-foreground">{describeSchedule(job.schedule)}</div>
                  </td>
                  <td className="p-3 font-mono text-xs max-w-xs truncate">{job.command}</td>
                  <td className="p-3 text-muted-foreground">{job.run_as}</td>
                  <td className="p-3 text-muted-foreground text-xs">
                    {job.last_run ? new Date(job.last_run).toLocaleString() : 'Never'}
                  </td>
                  <td className="p-3">
                    {job.last_status ? (
                      <span className={`text-xs px-2 py-1 rounded ${job.last_status === 'success' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
                        {job.last_status}
                      </span>
                    ) : (
                      <span className="text-muted-foreground">—</span>
                    )}
                  </td>
                  <td className="p-3">
                    <div className="flex gap-2">
                      <button onClick={() => toggleJob(job.id, !job.enabled)} className="text-primary hover:underline text-xs">
                        {job.enabled ? 'Disable' : 'Enable'}
                      </button>
                      <button onClick={() => viewLogs(job)} className="text-primary hover:underline text-xs">Logs</button>
                      <button onClick={() => deleteJob(job.id)} className="text-red-600 hover:underline text-xs">Delete</button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card border border-border p-6 w-[480px]">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-medium">Create Cron Job</h3>
              <button onClick={() => setShowModal(false)} className="text-muted-foreground hover:text-foreground">
                <X size={20} />
              </button>
            </div>

            <div className="space-y-3">
              <div>
                <label className="block text-sm mb-1">Schedule (Cron Expression)</label>
                <input
                  type="text"
                  value={formData.schedule}
                  onChange={(e) => setFormData({ ...formData, schedule: e.target.value })}
                  className="w-full border border-border px-3 py-2 text-sm font-mono"
                  placeholder="0 3 * * *"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Format: minute hour day month weekday. Use standard cron syntax.
                </p>
              </div>
              <div>
                <label className="block text-sm mb-1">Command</label>
                <textarea
                  value={formData.command}
                  onChange={(e) => setFormData({ ...formData, command: e.target.value })}
                  className="w-full border border-border px-3 py-2 text-sm font-mono h-20"
                  placeholder="/usr/local/bin/backup.sh"
                />
              </div>
              <div>
                <label className="block text-sm mb-1">Run As</label>
                <select
                  value={formData.run_as}
                  onChange={(e) => setFormData({ ...formData, run_as: e.target.value })}
                  className="w-full border border-border px-3 py-2 text-sm"
                >
                  <option value="root">root</option>
                  <option value="juvia">juvia</option>
                </select>
              </div>

              <button
                onClick={createJob}
                className="w-full bg-primary text-primary-foreground py-2 text-sm font-medium hover:opacity-90"
              >
                Create Cron Job
              </button>
            </div>
          </div>
        </div>
      )}

      {showLogs && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card border border-border p-6 w-[640px] max-h-[80vh] overflow-auto">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-medium">Execution History</h3>
              <button onClick={() => setShowLogs(null)} className="text-muted-foreground hover:text-foreground">
                <X size={20} />
              </button>
            </div>

            {logs.length === 0 ? (
              <p className="text-muted-foreground text-center py-4">No execution history</p>
            ) : (
              <div className="space-y-3">
                {logs.map((log) => (
                  <div key={log.id} className="border border-border p-3">
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-xs text-muted-foreground">
                        {new Date(log.run_at).toLocaleString()}
                      </span>
                      <span className={`text-xs px-2 py-1 rounded ${log.status === 'success' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
                        {log.status}
                      </span>
                    </div>
                    <pre className="text-xs font-mono bg-black text-white p-2 overflow-auto max-h-32">
                      {log.output || '(no output)'}
                    </pre>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}