import { useState, useEffect } from 'react'
import api from '../../lib/api'

interface Recording {
  id: string
  filename: string
  user_id: number
  started_at: string
  duration_seconds: number
  size_bytes: number
}

export default function Recordings() {
  const [recordings, setRecordings] = useState<Recording[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedRecording, setSelectedRecording] = useState<Recording | null>(null)
  const [playbackData, setPlaybackData] = useState<string>('')

  useEffect(() => {
    fetchRecordings()
  }, [])

  const fetchRecordings = async () => {
    try {
      const res = await api.get('/terminal/recordings')
      const data = res.data
      if (data.success) {
        setRecordings(data.data)
      }
    } catch (err) {
      console.error('Failed to fetch recordings:', err)
    } finally {
      setLoading(false)
    }
  }

  const loadRecording = async (recording: Recording) => {
    try {
      const res = await api.get(`/terminal/recordings/${encodeURIComponent(recording.filename)}`)
      const data = res.data
      if (data.success) {
        setPlaybackData(data.data.content)
        setSelectedRecording(recording)
      }
    } catch (err) {
      console.error('Failed to load recording:', err)
    }
  }

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString()
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-muted-foreground">Loading recordings...</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Terminal Recordings</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Session recordings · Stored as asciinema format
          </p>
        </div>
      </div>

      {recordings.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          No terminal recordings yet. Sessions are recorded when you use the terminal.
        </div>
      ) : (
        <div className="grid gap-4">
          <div className="border rounded-lg">
            <table className="w-full">
              <thead>
                <tr className="border-b bg-muted/50">
                  <th className="text-left p-3 text-sm font-medium">Date</th>
                  <th className="text-left p-3 text-sm font-medium">Filename</th>
                  <th className="text-left p-3 text-sm font-medium">Duration</th>
                  <th className="text-left p-3 text-sm font-medium">Size</th>
                  <th className="text-right p-3 text-sm font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {recordings.map((recording) => (
                  <tr key={recording.id} className="border-b hover:bg-muted/30">
                    <td className="p-3 text-sm">{formatDate(recording.started_at)}</td>
                    <td className="p-3 text-sm font-mono text-xs">{recording.filename}</td>
                    <td className="p-3 text-sm">{formatDuration(recording.duration_seconds)}</td>
                    <td className="p-3 text-sm">{formatSize(recording.size_bytes)}</td>
                    <td className="p-3 text-right">
                      <button
                        onClick={() => loadRecording(recording)}
                        className="text-sm text-primary hover:underline"
                      >
                        View
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {selectedRecording && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-background border rounded-lg w-3/4 max-w-4xl max-h-[80vh] flex flex-col">
            <div className="flex items-center justify-between p-4 border-b">
              <div>
                <h2 className="font-semibold">{selectedRecording.filename}</h2>
                <p className="text-sm text-muted-foreground">
                  {formatDuration(selectedRecording.duration_seconds)} · {formatSize(selectedRecording.size_bytes)}
                </p>
              </div>
              <button
                onClick={() => setSelectedRecording(null)}
                className="p-2 hover:bg-accent rounded"
              >
                ✕
              </button>
            </div>
            <div className="flex-1 overflow-auto p-4 bg-black text-green-400 font-mono text-sm">
              <pre className="whitespace-pre-wrap">{playbackData}</pre>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}