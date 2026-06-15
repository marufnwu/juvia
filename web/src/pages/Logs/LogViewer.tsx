import { useEffect, useState, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { FileText, Download, Search, Pause, Play, RefreshCw, Filter } from 'lucide-react'
import { Card } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { PageHeader } from '../../components/ui/Misc'
import { Input, Select } from '../../components/ui/Input'
import api from '../../lib/api'
import { cn } from '../../lib/utils'

interface LogEntry {
  timestamp: string
  level: 'info' | 'warn' | 'error'
  message: string
  source?: string
}

interface LogViewerProps {
  websiteId?: number
  logType?: 'access' | 'error' | 'system'
}

export default function LogViewer({ websiteId, logType = 'system' }: LogViewerProps) {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState('')
  const [levelFilter, setLevelFilter] = useState<string>('')
  const [autoRefresh, setAutoRefresh] = useState(false)
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    loadLogs()
    if (autoRefresh) {
      const interval = setInterval(loadLogs, 3000)
      return () => clearInterval(interval)
    }
  }, [autoRefresh, websiteId, logType])

  const loadLogs = async () => {
    try {
      let res
      if (websiteId) {
        res = await api.get(`/websites/${websiteId}/logs/${logType}`)
      } else {
        res = await api.get('/logs/system')
      }
      if (res.data.success) {
        setLogs(res.data.data || [])
      }
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const filteredLogs = logs.filter((log) => {
    const matchesSearch = !filter || log.message.toLowerCase().includes(filter.toLowerCase())
    const matchesLevel = !levelFilter || log.level === levelFilter
    return matchesSearch && matchesLevel
  })

  const levelColors: Record<string, string> = {
    info: 'text-primary',
    warn: 'text-warning',
    error: 'text-danger',
  }

  const levelBgColors: Record<string, string> = {
    info: 'bg-primary/10',
    warn: 'bg-warning/10',
    error: 'bg-danger/10',
  }

  const getTitle = () => {
    if (websiteId) {
      return logType === 'access' ? 'Access Logs · Every visit to your website' : 'Error Logs · Problems your site encountered'
    }
    return 'System Logs · Server internal activity'
  }

  const getDescription = () => {
    if (websiteId) {
      return logType === 'access' ? 'Website access log monitoring' : 'Website error log monitoring'
    }
    return 'Real-time server log monitoring'
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={getTitle()}
        description={getDescription()}
        breadcrumbs={[
          { label: 'Logs' },
          websiteId ? { label: `Website #${websiteId}` } : null,
        ].filter(Boolean) as { label: string; href?: string }[]}
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant={autoRefresh ? 'primary' : 'outline'}
              size="sm"
              onClick={() => setAutoRefresh(!autoRefresh)}
            >
              <RefreshCw className={cn('w-4 h-4 mr-1', autoRefresh && 'animate-spin')} />
              Auto-refresh
            </Button>
            <Button variant="outline" size="sm" onClick={loadLogs}>
              <RefreshCw className="w-4 h-4 mr-1" />
              Refresh
            </Button>
          </div>
        }
      />

      <Card padding="none">
        <div className="p-4 flex items-center gap-4 border-b border-border">
          <div className="w-64">
            <Input
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              placeholder="Search logs..."
            />
          </div>
          <Select
            value={levelFilter}
            onChange={(e) => setLevelFilter(e.target.value)}
            className="w-48"
          >
            <option value="">All Levels</option>
            <option value="info">Info</option>
            <option value="warn">Warning</option>
            <option value="error">Error</option>
          </Select>
          <div className="flex-1" />
          <div className="flex items-center gap-4 text-xs text-text-secondary">
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-primary" /> Info
            </span>
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-warning" /> Warning
            </span>
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-danger" /> Error
            </span>
          </div>
        </div>

        <div ref={scrollRef} className="h-[calc(100vh-20rem)] overflow-auto font-mono text-xs">
          {loading ? (
            <div className="p-8 text-center text-text-secondary">Loading logs...</div>
          ) : filteredLogs.length === 0 ? (
            <div className="p-8 text-center text-text-secondary">
              <FileText className="w-10 h-10 mx-auto text-text-secondary/50 mb-3" />
              <p>No logs found</p>
            </div>
          ) : (
            <div className="divide-y divide-border">
              {filteredLogs.map((log, i) => (
                <div
                  key={i}
                  className={cn('flex items-start gap-3 p-3 hover:bg-accent/30 transition-colors', levelBgColors[log.level] || '')}
                >
                  <span className="text-text-secondary shrink-0">{log.timestamp}</span>
                  <span className={cn('w-16 font-medium shrink-0 uppercase', levelColors[log.level] || '')}>
                    {log.level}
                  </span>
                  {log.source && (
                    <span className="text-text-secondary shrink-0">[{log.source}]</span>
                  )}
                  <span className="text-foreground flex-1">{log.message}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      </Card>
    </div>
  )
}
