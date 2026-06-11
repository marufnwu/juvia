import { useEffect, useState } from 'react'
import { Archive, Plus, Download } from 'lucide-react'
import api from '../../lib/api'

interface Backup {
  id: number
  website_id: number
  type: string
  status: string
  storage: string
  path: string
  size_bytes: number
  checksum: string
  verified: boolean
  verified_at: string
  created_at: string
}

interface BackupSchedule {
  id: number
  website_id: number
  schedule: string
  retention_days: number
  storage: string
  enabled: boolean
  created_at: string
}

export default function Backups() {
  const [backups, setBackups] = useState<Backup[]>([])
  const [schedules, setSchedules] = useState<BackupSchedule[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<'backups' | 'schedules'>('backups')

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const [backupRes, scheduleRes] = await Promise.all([
        api.get('/backups'),
        api.get('/backup-schedules'),
      ])
      setBackups(backupRes.data.data || [])
      setSchedules(scheduleRes.data.data || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const createBackup = async () => {
    try {
      await api.post('/backups', { type: 'full', storage: 'local' })
      loadData()
    } catch (err) {
      console.error(err)
    }
  }

  const deleteBackup = async (id: number) => {
    if (!confirm('Delete this backup?')) return
    try {
      await api.delete(`/backups/${id}`)
      loadData()
    } catch (err) {
      console.error(err)
    }
  }

  const formatSize = (bytes: number): string => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
  }

  const getBackupTypeLabel = (type: string): string => {
    const labels: Record<string, string> = {
      full: 'Full · All files and database',
      files: 'Files Only',
      database: 'Database Only',
    }
    return labels[type] || type
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Backups</h1>
          <p className="text-sm text-muted-foreground">
            Backup · A snapshot of your website you can restore if something goes wrong
          </p>
        </div>
        <button
          onClick={createBackup}
          className="flex items-center gap-2 bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90"
        >
          <Plus size={16} />
          Create Backup
        </button>
      </div>

      <div className="flex border-b border-border mb-6">
        <button
          onClick={() => setActiveTab('backups')}
          className={`px-4 py-2 text-sm font-medium ${activeTab === 'backups' ? 'border-b-2 border-primary' : 'text-muted-foreground'}`}
        >
          Backups
        </button>
        <button
          onClick={() => setActiveTab('schedules')}
          className={`px-4 py-2 text-sm font-medium ${activeTab === 'schedules' ? 'border-b-2 border-primary' : 'text-muted-foreground'}`}
        >
          Schedules
        </button>
      </div>

      {loading ? (
        <div className="text-muted-foreground">Loading...</div>
      ) : activeTab === 'backups' ? (
        backups.length === 0 ? (
          <div className="text-center py-12 border border-border">
            <Archive size={48} className="mx-auto text-muted-foreground mb-4" />
            <p className="text-muted-foreground">No backups yet</p>
          </div>
        ) : (
          <div className="border border-border">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 border-b border-border">
                <tr>
                  <th className="text-left p-3 font-medium">Type</th>
                  <th className="text-left p-3 font-medium">Status</th>
                  <th className="text-left p-3 font-medium">Size</th>
                  <th className="text-left p-3 font-medium">Verified</th>
                  <th className="text-left p-3 font-medium">Created</th>
                  <th className="text-left p-3 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {backups.map((backup) => (
                  <tr key={backup.id} className="border-b border-border hover:bg-muted/30">
                    <td className="p-3">{getBackupTypeLabel(backup.type)}</td>
                    <td className="p-3">
                      <span className={`text-xs px-2 py-1 rounded ${
                        backup.status === 'completed' ? 'bg-green-100 text-green-700' :
                        backup.status === 'failed' ? 'bg-red-100 text-red-700' :
                        'bg-yellow-100 text-yellow-700'
                      }`}>
                        {backup.status}
                      </span>
                    </td>
                    <td className="p-3 text-muted-foreground">{formatSize(backup.size_bytes)}</td>
                    <td className="p-3">
                      {backup.verified ? (
                        <span className="text-green-600 text-xs">Verified</span>
                      ) : (
                        <span className="text-muted-foreground text-xs">—</span>
                      )}
                    </td>
                    <td className="p-3 text-muted-foreground">
                      {new Date(backup.created_at).toLocaleString()}
                    </td>
                    <td className="p-3">
                      <div className="flex gap-2">
                        <button className="text-primary hover:underline text-xs flex items-center gap-1">
                          <Download size={12} /> Download
                        </button>
                        <button className="text-primary hover:underline text-xs">Restore</button>
                        <button onClick={() => deleteBackup(backup.id)} className="text-red-600 hover:underline text-xs">
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      ) : (
        <div className="border border-border">
          {schedules.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <p>No backup schedules configured</p>
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead className="bg-muted/50 border-b border-border">
                <tr>
                  <th className="text-left p-3 font-medium">Schedule</th>
                  <th className="text-left p-3 font-medium">Retention</th>
                  <th className="text-left p-3 font-medium">Storage</th>
                  <th className="text-left p-3 font-medium">Status</th>
                </tr>
              </thead>
              <tbody>
                {schedules.map((s) => (
                  <tr key={s.id} className="border-b border-border">
                    <td className="p-3 font-mono text-xs">{s.schedule}</td>
                    <td className="p-3 text-muted-foreground">{s.retention_days} days</td>
                    <td className="p-3 text-muted-foreground">{s.storage}</td>
                    <td className="p-3">
                      <span className={`text-xs px-2 py-1 rounded ${s.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'}`}>
                        {s.enabled ? 'Active' : 'Disabled'}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}
    </div>
  )
}