import { Link, useLocation } from 'react-router-dom'
import {
  LayoutDashboard,
  Globe,
  Database,
  Mail,
  FolderOpen,
  Shield,
  Archive,
  Clock,
  Activity,
  Terminal,
  Settings,
  Trash2,
} from 'lucide-react'

const nav = [
  { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/websites', label: 'Websites', icon: Globe },
  { to: '/websites/trash', label: 'Trash', icon: Trash2, indent: true },
  { to: '/databases', label: 'Databases', icon: Database },
  { to: '/email', label: 'Email', icon: Mail },
  { to: '/email/webmail', label: 'Webmail', icon: Mail, indent: true },
  { to: '/files', label: 'Files', icon: FolderOpen },
  { to: '/firewall', label: 'Firewall', icon: Shield },
  { to: '/backups', label: 'Backups', icon: Archive },
  { to: '/cron', label: 'Cron Jobs', icon: Clock },
  { to: '/metrics', label: 'Metrics', icon: Activity },
  { to: '/terminal', label: 'Terminal', icon: Terminal },
  { to: '/settings', label: 'Settings', icon: Settings },
]

export default function Sidebar() {
  const location = useLocation()
  return (
    <aside className="w-60 border-r border-border bg-card flex flex-col">
      <div className="h-14 flex items-center px-4 border-b border-border">
        <span className="font-semibold text-lg tracking-tight">Juvia</span>
      </div>
      <nav className="flex-1 py-2">
        {nav.map((item) => {
          const active = location.pathname.startsWith(item.to)
          return (
            <Link
              key={item.to}
              to={item.to}
              className={
                'flex items-center gap-3 px-4 py-2 text-sm font-medium transition-colors ' +
                (item.indent ? 'pl-10 ' : '') +
                (active
                  ? 'bg-primary text-primary-foreground'
                  : 'text-foreground hover:bg-accent')
              }
            >
<item.icon size={16} />
              {item.label}
            </Link>
          )
        })}
      </nav>
    </aside>
  )
}
