import { useEffect, useState, useRef } from 'react'
import { Play, Trash2, Clock, RefreshCw, Download, X } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { Modal, ConfirmModal } from '../../components/ui/Modal'
import { PageHeader } from '../../components/ui/Misc'
import { cn } from '../../lib/utils'
import { formatDate, formatBytes, timeAgo } from '../../lib/utils'
import { useApiError } from '../../hooks/useToast'
import api from '../../lib/api'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

interface Recording {
  id: string
  name: string
  duration: number
  created_at: string
  size: number
}

export default function Recordings() {
  const showError = useApiError()
  const [recordings, setRecordings] = useState<Recording[]>([])
  const [loading, setLoading] = useState(true)
  const [deleting, setDeleting] = useState<string | null>(null)
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null)
  const [selectedRecording, setSelectedRecording] = useState<Recording | null>(null)
  const [showViewer, setShowViewer] = useState(false)
  const [viewerLoading, setViewerLoading] = useState(false)
  const terminalRef = useRef<HTMLDivElement>(null)
  const termRef = useRef<Terminal | null>(null)

  useEffect(() => {
    loadRecordings()
  }, [])

  const loadRecordings = async () => {
    try {
      const res = await api.get('/terminal/recordings')
      setRecordings(res.data.data || [])
    } catch (err: any) {
      showError(err, 'Failed to load recordings')
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = (id: string) => {
    setConfirmDelete(id)
  }

  const doDelete = async () => {
    if (!confirmDelete) return
    const id = confirmDelete
    setConfirmDelete(null)
    setDeleting(id)
    try {
      await api.delete(`/terminal/recordings/${id}`)
      setRecordings(recordings.filter((r) => r.id !== id))
    } catch (err: any) {
      showError(err, 'Failed to delete recording')
    } finally {
      setDeleting(null)
    }
  }

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}m ${secs}s`
  }

  const handlePlay = async (rec: Recording) => {
    setSelectedRecording(rec)
    setShowViewer(true)
    setViewerLoading(true)
    try {
      const res = await api.get(`/terminal/recordings/${rec.id}`)
      const content = res.data.data?.content || ''

      if (termRef.current) {
        termRef.current.dispose()
      }

      const term = new Terminal({
        convertEol: true,
        disableStdin: true,
        fontSize: 13,
        fontFamily: 'Menlo, Monaco, "Courier New", monospace',
        theme: {
          background: '#0d1117',
          foreground: '#c9d1d9',
        },
        rows: 30,
      })
      const fitAddon = new FitAddon()
      term.loadAddon(fitAddon)
      term.open(termRef.current as unknown as HTMLElement)
      fitAddon.fit()
      termRef.current = term

      const lines = content.split('\n')
      for (const line of lines) {
        term.writeln(line)
      }
    } catch (err: any) {
      showError(err, 'Failed to load recording')
    } finally {
      setViewerLoading(false)
    }
  }

  const closeViewer = () => {
    setShowViewer(false)
    setSelectedRecording(null)
    if (termRef.current) {
      termRef.current.dispose()
      termRef.current = null
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Session Recordings · Saved terminal playback"
        description="Recorded terminal sessions for review and audit"
        breadcrumbs={[
          { label: 'Terminal', href: '/terminal' },
          { label: 'Recordings' },
        ]}
        actions={
          <Button variant="outline" onClick={loadRecordings}>
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
        }
      />

      <Card padding="none">
        {loading ? (
          <div className="flex items-center justify-center h-48">
            <RefreshCw className="w-6 h-6 animate-spin text-text-secondary" />
          </div>
        ) : recordings.length === 0 ? (
          <div className="p-8 text-center">
            <Play className="w-12 h-12 mx-auto text-text-secondary/50 mb-3" />
            <p className="text-text-secondary text-sm mb-3">No recordings yet</p>
            <p className="text-xs text-text-secondary">
              Terminal sessions are automatically recorded when enabled
            </p>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {recordings.map((rec) => (
              <div key={rec.id} className="flex items-center gap-4 p-4 hover:bg-accent/30 transition-colors">
                <div className="w-10 h-10 rounded bg-accent/50 flex items-center justify-center">
                  <Play className="w-5 h-5 text-text-secondary" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="font-medium text-sm truncate">{rec.name}</p>
                  <div className="flex items-center gap-3 mt-1 text-xs text-text-secondary">
                    <span className="flex items-center gap-1">
                      <Clock className="w-3 h-3" />
                      {formatDuration(rec.duration)}
                    </span>
                    <span>{formatBytes(rec.size)}</span>
                    <span>{timeAgo(rec.created_at)}</span>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handlePlay(rec)}
                    title="Play"
                  >
                    <Play className="w-4 h-4" />
                  </Button>
                  <Button
                    variant="danger"
                    size="sm"
                    onClick={() => handleDelete(rec.id)}
                    disabled={deleting === rec.id}
                    loading={deleting === rec.id}
                    title="Delete"
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>

      <Modal open={showViewer} onClose={closeViewer} title={selectedRecording?.name || 'Recording'} size="lg">
        <div className="space-y-3">
          {viewerLoading ? (
            <div className="flex items-center justify-center h-64">
              <RefreshCw className="w-6 h-6 animate-spin text-text-secondary" />
            </div>
          ) : (
            <div ref={terminalRef} className="rounded border border-border overflow-hidden" style={{ background: '#0d1117' }} />
          )}
          <div className="flex justify-end">
            <Button variant="outline" onClick={closeViewer}>Close</Button>
          </div>
        </div>
      </Modal>

      <ConfirmModal
        open={!!confirmDelete}
        onClose={() => setConfirmDelete(null)}
        onConfirm={doDelete}
        title="Delete Recording"
        description="Permanently delete this terminal recording? This cannot be undone."
        confirmLabel="Delete"
        variant="danger"
        loading={!!deleting}
      />
    </div>
  )
}
