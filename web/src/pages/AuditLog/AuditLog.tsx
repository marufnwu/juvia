import { useEffect, useState } from 'react'
import { SearchInput } from '../../components/ui/SearchInput'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Badge } from '../../components/ui/Badge'
import { PageHeader } from '../../components/ui/Misc'
import { formatDate, timeAgo } from '../../lib/utils'
import { useApiError } from '../../hooks/useToast'
import api from '../../lib/api'
import { cn } from '../../lib/utils'
import { Shield, User, Globe, Mail, Database, Clock, Settings, Activity } from 'lucide-react'

interface AuditEntry {
  id: string
  timestamp: string
  user: string
  action: string
  resource: string
  details: string
  ip: string
}

const actionIcons: Record<string, React.ElementType> = {
  'website.create': Globe,
  'website.update': Globe,
  'website.delete': Globe,
  'website.suspend': Globe,
  'mailbox.create': Mail,
  'mailbox.update': Mail,
  'mailbox.delete': Mail,
  'database.create': Database,
  'database.delete': Database,
  'user.login': User,
  'user.logout': User,
  'user.create': User,
  'settings.update': Settings,
}

const actionColors: Record<string, string> = {
  create: 'text-success',
  delete: 'text-danger',
  update: 'text-warning',
  login: 'text-primary',
  logout: 'text-text-secondary',
}

export default function AuditLog() {
  const showError = useApiError()
  const [entries, setEntries] = useState<AuditEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')

  useEffect(() => {
    loadEntries()
  }, [])

  const loadEntries = async () => {
    try {
      const res = await api.get('/audit-log')
      setEntries(res.data.data || [])
    } catch (err: any) {
      showError(err, 'Failed to load audit log')
    } finally {
      setLoading(false)
    }
  }

  const filteredEntries = search
    ? entries.filter((e) =>
        e.action.toLowerCase().includes(search.toLowerCase()) ||
        e.resource.toLowerCase().includes(search.toLowerCase()) ||
        e.user.toLowerCase().includes(search.toLowerCase())
      )
    : entries

  const getIcon = (action: string) => {
    const Icon = actionIcons[action] || Clock
    return Icon
  }

  const getColor = (action: string) => {
    for (const [key, color] of Object.entries(actionColors)) {
      if (action.includes(key)) return color
    }
    return 'text-text-secondary'
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Audit Log"
        description="Track all administrative actions and user activity"
        breadcrumbs={[{ label: 'Audit Log' }]}
      />

      <Card padding="none">
        <div className="p-4 border-b border-border">
          <SearchInput
            value={search}
            onChange={setSearch}
            placeholder="Search by action, resource, or user..."
          />
        </div>
        {loading ? (
          <div className="p-8 text-center text-text-secondary">Loading...</div>
        ) : filteredEntries.length === 0 ? (
          <div className="p-8 text-center text-text-secondary text-sm">
            <Activity className="w-10 h-10 mx-auto text-text-secondary/50 mb-3" />
            <p>No audit entries yet</p>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {filteredEntries.map((entry) => {
              const Icon = getIcon(entry.action)
              const color = getColor(entry.action)
              return (
                <div key={entry.id} className="flex items-start gap-4 p-4 hover:bg-accent/30 transition-colors">
                  <div className={cn('w-8 h-8 rounded flex items-center justify-center bg-accent/50 mt-0.5', color)}>
                    <Icon className="w-4 h-4" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="font-medium text-sm">{entry.action}</span>
                      <span className="text-text-secondary text-xs">on</span>
                      <span className="text-sm text-text-secondary">{entry.resource}</span>
                    </div>
                    <p className="text-xs text-text-secondary mt-1">{entry.details}</p>
                    <div className="flex items-center gap-3 mt-2 text-xs text-text-secondary">
                      <span className="flex items-center gap-1">
                        <User className="w-3 h-3" />
                        {entry.user}
                      </span>
                      <span className="flex items-center gap-1">
                        <Clock className="w-3 h-3" />
                        {timeAgo(entry.timestamp)}
                      </span>
                      {entry.ip && <span>{entry.ip}</span>}
                    </div>
                  </div>
                  <span className="text-xs text-text-secondary whitespace-nowrap">
                    {formatDate(entry.timestamp)}
                  </span>
                </div>
              )
            })}
          </div>
        )}
      </Card>
    </div>
  )
}
