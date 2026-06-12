import { useEffect, useState } from 'react'
import { Terminal as TerminalIcon, Plus, AlertTriangle } from 'lucide-react'
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
  const [showRootWarning, setShowRootWarning] = useState(false)

  useEffect(() => {
    loadSessions()
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

  const createSession = async (type: 'root' | 'site', siteId?: string) => {
    if (type === 'root') {
      setShowRootWarning(true)
      return
    }
    try {
      const res = await api.post('/terminal/session', { type, website_id: siteId })
      if (res.data.success) {
        setActiveSession(res.data.data.id)
        loadSessions()
      }
    } catch (err) {
      console.error(err)
    }
  }

  const closeSession = async (id: string) => {
    try {
      await api.delete(`/terminal/sessions/${id}`)
      if (activeSession === id) setActiveSession(null)
      loadSessions()
    } catch (err) {
      console.error(err)
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Terminal"
        description="Interactive server terminal access"
        breadcrumbs={[{ label: 'Terminal' }]}
        actions={
          <div className="flex items-center gap-2">
            <Button variant="outline" onClick={() => createSession('site')}>
              <Plus className="w-4 h-4 mr-2" />
              New Site Terminal
            </Button>
            <Button variant="outline" onClick={() => createSession('root')}>
              <AlertTriangle className="w-4 h-4 mr-2" />
              New Root Terminal
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
                        <Badge variant="danger">ROOT</Badge>
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
                    <Button size="sm" variant="outline">Connect</Button>
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