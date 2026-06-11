import { useEffect, useRef, useState } from 'react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { Link } from 'react-router-dom'
import api from '../../lib/api'
import { WSClient } from '../../lib/ws'
import { ArrowLeft, Copy, Check } from 'lucide-react'
import '@xterm/xterm/css/xterm.css'

export default function TerminalPage() {
  const terminalRef = useRef<HTMLDivElement>(null)
  const termRef = useRef<Terminal | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const wsRef = useRef<WSClient | null>(null)
  const [sessionId, setSessionId] = useState<string | null>(null)
  const [connected, setConnected] = useState(false)
  const [copied, setCopied] = useState(false)
  const [sessionInfo, setSessionInfo] = useState({ started_at: new Date().toISOString() })

  useEffect(() => {
    if (!terminalRef.current) return

    const term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#0a0a0a',
        foreground: '#e5e5e5',
        cursor: '#e5e5e5',
      },
      scrollback: 10000,
    })
    termRef.current = term

    const fitAddon = new FitAddon()
    fitAddonRef.current = fitAddon
    term.loadAddon(fitAddon)
    term.open(terminalRef.current)
    fitAddon.fit()

    term.writeln('\x1b[1;32mJuvia Terminal\x1b[0m')
    term.writeln('Connecting to server...\r\n')

    createSession(term)

    const handleResize = () => {
      if (fitAddonRef.current) {
        fitAddonRef.current.fit()
        const { rows, cols } = term
        if (wsRef.current) {
          wsRef.current.send({ type: 'resize', data: { rows, cols } })
        }
      }
    }

    window.addEventListener('resize', handleResize)

    const resizeObserver = new ResizeObserver(handleResize)
    resizeObserver.observe(terminalRef.current)

    return () => {
      window.removeEventListener('resize', handleResize)
      resizeObserver.disconnect()
      if (wsRef.current) {
        wsRef.current.close()
      }
      if (sessionId) {
        api.delete(`/terminal/sessions/${sessionId}`).catch(() => {})
      }
      term.dispose()
    }
  }, [])

  const createSession = async (term: Terminal) => {
    try {
      const res = await api.post('/terminal/session', {})
      const data = res.data
      if (!data.success) {
        term.writeln('\x1b[31mFailed to create session\x1b[0m')
        return
      }

      const sid = data.data.session_id
      setSessionId(sid)
      setSessionInfo({ started_at: new Date().toISOString() })

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const host = window.location.host
      const ws = new WSClient(`${protocol}//${host}/ws/v1/terminal/${sid}`)
      wsRef.current = ws

      ws.on('output', (outputData: unknown) => {
        const output = (outputData as { data: string }).data
        term.write(output)
      })

      ws.on('connect', () => {
        setConnected(true)
        term.writeln('\x1b[32mConnected\x1b[0m — type commands below\r\n')
        const { rows, cols } = term
        ws.send({ type: 'resize', data: { rows, cols } })
      })

      ws.on('close', () => {
        setConnected(false)
        term.writeln('\r\n\x1b[33mSession closed\x1b[0m')
      })

      term.onData((data) => {
        ws.send({ type: 'input', data: { data } })
      })

      ws.connect()
    } catch (err) {
      term.writeln(`\x1b[31mError: ${err}\x1b[0m`)
    }
  }

  const copySessionLink = () => {
    if (sessionId) {
      navigator.clipboard.writeText(sessionId)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-3">
          <Link
            to="/terminal/recordings"
            className="p-2 hover:bg-accent rounded-lg transition-colors"
          >
            <ArrowLeft size={16} />
          </Link>
          <div>
            <h1 className="text-lg font-semibold">Terminal Session</h1>
            <p className="text-xs text-muted-foreground">
              Session ID: {sessionId || '—'}
              {sessionId && (
                <button
                  onClick={copySessionLink}
                  className="ml-2 inline-flex items-center gap-1 text-primary hover:underline"
                >
                  {copied ? <Check size={12} /> : <Copy size={12} />}
                  {copied ? 'Copied' : 'Copy'}
                </button>
              )}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <div
              className={`w-2 h-2 rounded-full ${connected ? 'bg-green-500' : 'bg-yellow-500'}`}
            />
            <span className="text-sm text-muted-foreground">
              {connected ? 'Connected' : 'Connecting...'}
            </span>
          </div>
          <Link
            to="/terminal/recordings"
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            View Recordings
          </Link>
        </div>
      </div>

      <div
        ref={terminalRef}
        className="flex-1 border rounded-lg overflow-hidden bg-black"
        style={{ minHeight: '400px' }}
      />

      <div className="mt-2 text-xs text-muted-foreground flex items-center gap-4">
        <span>
          Started: {new Date(sessionInfo.started_at).toLocaleTimeString()}
        </span>
        <span className="text-muted-foreground/50">
          Sessions are recorded for security purposes
        </span>
      </div>
    </div>
  )
}