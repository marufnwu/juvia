import { useEffect, useState, useRef } from 'react'
import { Terminal as TerminalIcon, Plus, AlertTriangle, X } from 'lucide-react'
import { Terminal as XTerminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { useApiError } from '../../hooks/useToast'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { PageHeader } from '../../components/ui/Misc'
import api from '../../lib/api'
import { cn } from '../../lib/utils'

interface TerminalSession {
  id: string
  name: string
  type: 'root' | 'site'
  site_name?: string
  created_at: string
  status: 'active' | 'closed'
}

export default function Terminal() {
  const [sessions, setSessions] = useState<TerminalSession[]>([])
  const [loading, setLoading] = useState(true)
  const [activeSession, setActiveSession] = useState<string | null>(null)
  const [activeTerminal, setActiveTerminal] = useState<XTerminal | null>(null)
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null)
  const [showRootWarning, setShowRootWarning] = useState(false)
  const termRef = useRef<HTMLDivElement>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const handleError = useApiError()

  useEffect(() => {
    loadSessions()
    return () => {
      wsRef.current?.close()
      activeTerminal?.dispose()
    }
  }, [])

  const loadSessions = async () => {
    try {
      const res = await api.get('/terminal/sessions')
      setSessions(res.data.data || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const connectToSession = (sessionId: string) => {
    wsRef.current?.close()
    activeTerminal?.dispose()

    const term = new XTerminal({
      theme: {
        background: '#0d1117',
        foreground: '#e6edf3',
        cursor: '#e6edf3',
        cursorAccent: '#0d1117',
        selectionBackground: '#264f78',
        black: '#0d1117',
        red: '#ff7b72',
        green: '#3fb950',
        yellow: '#d29922',
        blue: '#58a6ff',
        magenta: '#bc8cff',
        cyan: '#39c5cf',
        white: '#e6edf3',
        brightBlack: '#484f58',
        brightRed: '#ffa198',
        brightGreen: '#56d364',
        brightYellow: '#e3b341',
        brightBlue: '#79c0ff',
        brightMagenta: '#d2a8ff',
        brightCyan: '#56d4dd',
        brightWhite: '#f0f6fc',
      },
      fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace',
      fontSize: 14,
      cursorBlink: true,
    })

    const fitAddon = new FitAddon()
    term.loadAddon(fitAddon)

    if (termRef.current) {
      termRef.current.innerHTML = ''
      term.open(termRef.current)
      fitAddon.fit()
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${protocol}//${window.location.host}/ws/v1/terminal/${sessionId}`)
    wsRef.current = ws

    ws.onopen = () => {
      ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }))
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'output') {
          term.write(msg.data)
        }
      } catch {
        term.write(event.data)
      }
    }

    ws.onclose = () => {
      term.write('\r\n\x1b[33m[Session closed]\x1b[0m\r\n')
    }

    ws.onerror = () => {
      term.write('\r\n\x1b[31m[Connection error]\x1b[0m\r\n')
    }

    term.onData((data: string) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'input', data }))
      }
    })

    const resizeHandler = () => {
      fitAddon.fit()
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }))
      }
    }
    window.addEventListener('resize', resizeHandler)

    const origDispose = term.dispose.bind(term)
    term.dispose = () => {
      window.removeEventListener('resize', resizeHandler)
      origDispose()
    }

    setActiveTerminal(term)
    setActiveSessionId(sessionId)
  }

  const createSession = async (type: 'root' | 'site', siteId?: string) => {
    if (type === 'root') {
      setShowRootWarning(true)
      return
    }
    try {
      const res = await api.post('/terminal/session', { type, website_id: siteId })
      if (res.data.success) {
        const sessionId = res.data.data.id
        setActiveSession(sessionId)
        loadSessions()
        connectToSession(sessionId)
      }
    } catch (err) {
      handleError(err, 'Failed to create terminal session')
    }
  }

  const closeSession = async (id: string) => {
    try {
      await api.delete(`/terminal/sessions/${id}`)
      if (activeSession === id) {
        wsRef.current?.close()
        activeTerminal?.dispose()
        setActiveSession(null)
        setActiveTerminal(null)
        setActiveSessionId(null)
      }
      loadSessions()
    } catch (err) {
      handleError(err, 'Failed to close session')
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Terminal · Command-line for server management"
        description="Interactive server terminal access"
        breadcrumbs={[{ label: 'Terminal' }]}
        actions={
          <div className="flex items-center gap-2">
            <Button variant="outline" onClick={() => createSession('site')}>
              <Plus className="w-4 h-4 mr-2" />
              Site Terminal · Terminal for one website
            </Button>
            <Button variant="outline" onClick={() => createSession('root')}>
              <AlertTriangle className="w-4 h-4 mr-2" />
              Root Terminal · Full admin access (use carefully)
            </Button>
          </div>
        }
      />

      {showRootWarning && (
        <Card className="border-warning/50 bg-warning/5">
          <div className="flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-warning flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <h3 className="font-semibold text-warning mb-1">Root Terminal Warning</h3>
              <p className="text-sm text-text-secondary mb-3">
                You are about to open a root terminal. All actions are logged and cannot be undone.
                Root access should only be used for system administration tasks.
              </p>
              <div className="flex gap-2">
                <Button onClick={() => { setShowRootWarning(false); createSession('root') }}>
                  Continue as Root
                </Button>
                <Button variant="outline" onClick={() => setShowRootWarning(false)}>
                  Cancel
                </Button>
              </div>
            </div>
          </div>
        </Card>
      )}

      <Card padding="none">
        <div className="p-4 flex items-center justify-between border-b border-border">
          <div className="flex items-center gap-3">
            <TerminalIcon className="w-5 h-5 text-text-secondary" />
            <div>
              <p className="font-medium">{sessions.filter(s => s.status === 'active').length} active sessions</p>
              <p className="text-xs text-text-secondary">All sessions are recorded</p>
            </div>
          </div>
        </div>
        <div className="divide-y divide-border">
          {sessions.length === 0 ? (
            <div className="p-8 text-center">
              <TerminalIcon className="w-12 h-12 mx-auto text-text-secondary/50 mb-3" />
              <p className="text-text-secondary text-sm mb-3">No active terminal sessions</p>
              <p className="text-xs text-text-secondary">
                Sessions are isolated per website user or root
              </p>
            </div>
          ) : (
            sessions.map((session) => (
              <div
                key={session.id}
                className="flex items-center justify-between p-4 hover:bg-accent/30 transition-colors"
              >
                <div className="flex items-center gap-3">
                  <div className={cn('w-8 h-8 rounded flex items-center justify-center',
                    session.type === 'root' ? 'bg-danger/10' : 'bg-primary/10')}>
                    <TerminalIcon className={cn('w-4 h-4',
                      session.type === 'root' ? 'text-danger' : 'text-primary')} />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <p className="font-medium text-sm">{session.name}</p>
                      {session.type === 'root' && (
                        <Badge variant="danger">ROOT · Super-admin account</Badge>
                      )}
                    </div>
                    <p className="text-xs text-text-secondary mt-0.5">
                      {session.site_name || session.type} · {session.created_at}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant={session.status === 'active' ? 'success' : 'neutral'}>
                    {session.status}
                  </Badge>
                  {session.status === 'active' ? (
                    <Button size="sm" variant="outline" onClick={() => { setActiveSession(session.id); connectToSession(session.id) }}>Connect</Button>
                  ) : (
                    <Button size="sm" variant="outline" onClick={() => closeSession(session.id)}>
                      Delete
                    </Button>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      </Card>

      {activeSession && (
        <Card className="bg-[#0d1117] border-border">
          <div className="flex items-center justify-between p-3 border-b border-border/50">
            <div className="flex items-center gap-2">
              <TerminalIcon className="w-4 h-4 text-text-secondary" />
              <span className="text-sm font-medium text-white">Terminal Session</span>
              {activeSessionId && (
                <Badge variant="success" className="text-xs">Connected</Badge>
              )}
            </div>
            <Button variant="ghost" size="icon" onClick={() => {
              wsRef.current?.close()
              activeTerminal?.dispose()
              setActiveSession(null)
              setActiveTerminal(null)
              setActiveSessionId(null)
            }}>
              <X className="w-4 h-4" />
            </Button>
          </div>
          <div ref={termRef} className="w-full h-[500px] bg-[#0d1117] p-1" />
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardDescription>
            Terminal sessions are recorded in asciinema format for security and audit purposes.
            Sessions automatically terminate after 15 minutes of inactivity.
          </CardDescription>
        </CardHeader>
      </Card>
    </div>
  )
}