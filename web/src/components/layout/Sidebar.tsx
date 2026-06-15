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
  ChevronLeft,
  ChevronRight,
  Server,
  Users,
  Bell,
  Network,
} from 'lucide-react'
import { cn } from '../../lib/utils'
import { useUIStore } from '../../stores/uiStore'
import { useAuthStore } from '../../stores/authStore'

const navSections = [
  {
    label: 'Server',
    items: [
      { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard, tooltip: 'Server overview' },
      { to: '/metrics', label: 'Metrics', icon: Activity, tooltip: 'How your server is performing' },
      { to: '/terminal', label: 'Terminal', icon: Terminal, tooltip: 'Command-line for server management' },
      { to: '/terminal/recordings', label: 'Recordings', icon: Clock, tooltip: 'Saved terminal playback' },
    ],
  },
  {
    label: 'Hosting',
    items: [
      { to: '/websites', label: 'Websites', icon: Globe, tooltip: 'Manage your websites' },
      { to: '/websites/trash', label: 'Trash', icon: Trash2, tooltip: 'Deleted websites (30-day retention)' },
      { to: '/dns', label: 'DNS', icon: Network, tooltip: 'Authoritative DNS zones and nameservers' },
      { to: '/databases', label: 'Databases', icon: Database, tooltip: 'Structured data storage' },
      { to: '/email', label: 'Email', icon: Mail, tooltip: 'Mailboxes, aliases, and forwarding' },
      { to: '/email/webmail', label: 'Webmail', icon: Mail, tooltip: 'Browser-based email client' },
      { to: '/files', label: 'Files', icon: FolderOpen, tooltip: 'Manage website files' },
    ],
  },
  {
    label: 'Security',
    items: [
      { to: '/firewall', label: 'Firewall', icon: Shield, tooltip: 'Traffic allow/deny rules' },
      { to: '/backups', label: 'Backups', icon: Archive, tooltip: 'Save and restore data' },
      { to: '/cron', label: 'Cron Jobs', icon: Clock, tooltip: 'Scheduled automated tasks' },
      { to: '/alerts', label: 'Alerts', icon: Bell, tooltip: 'Monitoring notifications' },
    ],
  },
  {
    label: 'System',
    items: [
      { to: '/settings', label: 'Settings', icon: Settings, tooltip: 'Server configuration' },
      { to: '/users', label: 'Users', icon: Users, tooltip: 'Manage panel users' },
      { to: '/audit-log', label: 'Audit Log', icon: Activity, tooltip: 'History of all changes' },
    ],
  },
]

export default function Sidebar() {
  const location = useLocation()
  const { sidebarCollapsed, toggleSidebar } = useUIStore()
  const user = useAuthStore((s) => s.user)

  return (
    <aside
      className={cn(
        'h-screen bg-surface border-r border-border flex flex-col transition-all duration-200',
        sidebarCollapsed ? 'w-16' : 'w-60'
      )}
    >
      <div className="h-14 flex items-center justify-between px-4 border-b border-border">
        {!sidebarCollapsed && (
          <div className="flex items-center gap-2">
            <Server className="w-5 h-5 text-primary" />
            <span className="font-semibold text-base text-foreground tracking-tight">Juvia</span>
          </div>
        )}
        <button
          onClick={toggleSidebar}
          className="p-1.5 text-text-secondary hover:text-foreground hover:bg-accent rounded transition-colors"
        >
          {sidebarCollapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
        </button>
      </div>

      <nav className="flex-1 overflow-y-auto py-3">
        {navSections.map((section) => (
          <div key={section.label} className="mb-4">
            {!sidebarCollapsed && (
              <div className="px-4 mb-1">
                <span className="text-[10px] font-semibold uppercase tracking-wider text-text-secondary">
                  {section.label}
                </span>
              </div>
            )}
            {section.items.map((item) => {
              const active = location.pathname === item.to ||
                (location.pathname.startsWith(item.to + '/') &&
                 !section.items.some(sibling => sibling.to !== item.to && location.pathname.startsWith(sibling.to)))
              return (
                <Link
                  key={item.to}
                  to={item.to}
                  className={cn(
                    'flex items-center gap-3 mx-2 px-3 py-2 text-sm font-medium rounded transition-colors group',
                    active
                      ? 'bg-primary/10 text-primary border-l-2 border-primary ml-0'
                      : 'text-text-secondary hover:text-foreground hover:bg-accent/50'
                  )}
                  title={sidebarCollapsed ? item.label : item.tooltip}
                >
                  <item.icon
                    size={18}
                    className={cn(active ? 'text-primary' : 'text-text-secondary group-hover:text-foreground')}
                  />
                  {!sidebarCollapsed && <span>{item.label}</span>}
                </Link>
              )
            })}
          </div>
        ))}
      </nav>

      {!sidebarCollapsed && user && (
        <div className="p-3 border-t border-border">
          <div className="flex items-center gap-3 px-2">
            <div className="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center">
              <span className="text-xs font-semibold text-primary">
                {user.username?.charAt(0).toUpperCase()}
              </span>
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-foreground truncate">{user.username}</p>
              <p className="text-xs text-text-secondary truncate">{user.role}</p>
            </div>
          </div>
        </div>
      )}
    </aside>
  )
}