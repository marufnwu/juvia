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
} from 'lucide-react'
import { cn } from '../../lib/utils'
import { useUIStore } from '../../stores/uiStore'
import { useAuthStore } from '../../stores/authStore'

const navSections = [
  {
    label: 'Server',
    items: [
      { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
      { to: '/metrics', label: 'Metrics', icon: Activity },
      { to: '/terminal', label: 'Terminal', icon: Terminal },
    ],
  },
  {
    label: 'Hosting',
    items: [
      { to: '/websites', label: 'Websites', icon: Globe },
      { to: '/databases', label: 'Databases', icon: Database },
      { to: '/email', label: 'Email', icon: Mail },
      { to: '/files', label: 'Files', icon: FolderOpen },
    ],
  },
  {
    label: 'Security',
    items: [
      { to: '/firewall', label: 'Firewall', icon: Shield },
      { to: '/backups', label: 'Backups', icon: Archive },
      { to: '/cron', label: 'Cron Jobs', icon: Clock },
      { to: '/alerts', label: 'Alerts', icon: Bell },
    ],
  },
  {
    label: 'System',
    items: [
      { to: '/settings', label: 'Settings', icon: Settings },
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
              const active = location.pathname.startsWith(item.to)
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
                  title={sidebarCollapsed ? item.label : undefined}
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