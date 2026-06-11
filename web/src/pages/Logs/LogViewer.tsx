import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { RefreshCw, Download } from 'lucide-react'
import api from '../../lib/api'

interface LogLine {
  index: number
  content: string
}

export default function LogViewer() {
  const { id } = useParams<{ id: string }>()
  const [accessLogs, setAccessLogs] = useState<LogLine[]>([])
  const [errorLogs, setErrorLogs] = useState<LogLine[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<'access' | 'error'>('access')
  const [filter, setFilter] = useState('')
  const [lines, setLines] = useState(100)

  const loadLogs = async () => {
    if (!id) return
    setLoading(true)
    try {
      const [accessRes, errorRes] = await Promise.all([
        api.get(`/websites/${id}/logs/access?lines=${lines}`),
        api.get(`/websites/${id}/logs/error?lines=${lines}`),
      ])
      setAccessLogs(formatLogs(accessRes.data.data.lines || []))
      setErrorLogs(formatLogs(errorRes.data.data.lines || []))
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadLogs()
    const interval = setInterval(loadLogs, 5000)
    return () => clearInterval(interval)
  }, [id, lines])

  const formatLogs = (rawLogs: string[]): LogLine[] => {
    return rawLogs.map((content, index) => ({ index, content }))
  }

  const filteredLogs = activeTab === 'access'
    ? accessLogs.filter(l => l.content.toLowerCase().includes(filter.toLowerCase()))
    : errorLogs.filter(l => l.content.toLowerCase().includes(filter.toLowerCase()))

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-lg font-semibold">Website Logs</h1>
        <div className="flex items-center gap-2">
          <select
            value={lines}
            onChange={(e) => setLines(Number(e.target.value))}
            className="border border-border px-2 py-1 text-sm"
          >
            <option value={50}>50 lines</option>
            <option value={100}>100 lines</option>
            <option value={500}>500 lines</option>
            <option value={1000}>1000 lines</option>
          </select>
          <button
            onClick={loadLogs}
            className="flex items-center gap-1 px-3 py-1 text-sm border border-border hover:bg-muted"
          >
            <RefreshCw size={14} />
            Refresh
          </button>
          <button className="flex items-center gap-1 px-3 py-1 text-sm border border-border hover:bg-muted">
            <Download size={14} />
            Download
          </button>
        </div>
      </div>

      <div className="flex border-b border-border mb-4">
        <button
          onClick={() => setActiveTab('access')}
          className={`px-4 py-2 text-sm font-medium ${activeTab === 'access' ? 'border-b-2 border-primary' : 'text-muted-foreground'}`}
        >
          Access Log
        </button>
        <button
          onClick={() => setActiveTab('error')}
          className={`px-4 py-2 text-sm font-medium ${activeTab === 'error' ? 'border-b-2 border-primary' : 'text-muted-foreground'}`}
        >
          Error Log
        </button>
      </div>

      <p className="text-sm text-muted-foreground mb-4">
        {activeTab === 'access'
          ? 'Access Log · Records every visit to your website'
          : 'Error Log · Records errors and warnings from your website'
        }
      </p>

      <input
        type="text"
        placeholder="Filter logs..."
        value={filter}
        onChange={(e) => setFilter(e.target.value)}
        className="w-full border border-border px-3 py-2 text-sm mb-4"
      />

      {loading ? (
        <div className="text-muted-foreground">Loading logs...</div>
      ) : filteredLogs.length === 0 ? (
        <div className="text-muted-foreground text-center py-8">No log entries</div>
      ) : (
        <div className="border border-border bg-black text-white font-mono text-xs overflow-auto max-h-[calc(100vh-16rem)]">
          <table className="w-full">
            <tbody>
              {filteredLogs.map((log) => (
                <tr key={log.index} className="border-b border-gray-800">
                  <td className="px-3 py-1 text-gray-500 w-12">{log.index + 1}</td>
                  <td className="px-3 py-1 whitespace-pre-wrap break-all">{log.content}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}