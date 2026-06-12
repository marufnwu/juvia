import { useLocation, Link } from 'react-router-dom'
import { Bell, Search, Menu, ChevronRight, LogOut, User, Shield, Moon, Sun } from 'lucide-react'
import { useState, useRef, useEffect } from 'react'
import { cn } from '../../lib/utils'
import { useUIStore } from '../../stores/uiStore'
import { useAuthStore } from '../../stores/authStore'

const breadcrumbsMap: Record<string, string> = {
  dashboard: 'Dashboard',
  websites: 'Websites',
  databases: 'Databases',
  email: 'Email',
  files: 'Files',
  firewall: 'Firewall',
  backups: 'Backups',
  cron: 'Cron Jobs',
  alerts: 'Alerts',
  metrics: 'Metrics',
  terminal: 'Terminal',
  settings: 'Settings',
  trash: 'Trash',
  create: 'Create',
  webmail: 'Webmail',
  recordings: 'Recordings',
}

function getBreadcrumbs(pathname: string) {
  const parts = pathname.split('/').filter(Boolean)
  const crumbs = []

  let path = ''
  for (let i = 0; i < parts.length; i++) {
    path += '/' + parts[i]
    const label = breadcrumbsMap[parts[i]] || parts[i]
    crumbs.push({ label, path })
  }

  return crumbs
}

export default function TopBar() {
  const location = useLocation()
  const { theme, setTheme, toggleCommandPalette } = useUIStore()
  const { user, logout, sessions } = useAuthStore()
  const [showUserMenu, setShowUserMenu] = useState(false)
  const [showNotifications, setShowNotifications] = useState(false)
  const userMenuRef = useRef<HTMLDivElement>(null)
  const notifRef = useRef<HTMLDivElement>(null)

  const breadcrumbs = getBreadcrumbs(location.pathname)

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (userMenuRef.current && !userMenuRef.current.contains(e.target as Node)) {
        setShowUserMenu(false)
      }
      if (notifRef.current && !notifRef.current.contains(e.target as Node)) {
        setShowNotifications(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  return (
    <header className="h-14 border-b border-border bg-surface flex items-center px-4 gap-4">
      <div className="flex-1">
        {breadcrumbs.length > 1 && (
          <div className="flex items-center gap-2 text-sm">
            {breadcrumbs.map((crumb, i) => (
              <span key={crumb.path} className="flex items-center gap-2">
                {i > 0 && <ChevronRight className="w-3 h-3 text-text-secondary" />}
                {i === breadcrumbs.length - 1 ? (
                  <span className="font-medium text-foreground">{crumb.label}</span>
                ) : (
                  <Link to={crumb.path} className="text-text-secondary hover:text-foreground transition-colors">
                    {crumb.label}
                  </Link>
                )}
              </span>
            ))}
          </div>
        )}
      </div>

      <button
        onClick={toggleCommandPalette}
        className="flex items-center gap-2 px-3 py-1.5 text-sm text-text-secondary bg-accent/50 border border-border rounded hover:bg-accent transition-colors"
      >
        <Search className="w-3.5 h-3.5" />
        <span className="hidden sm:inline">Search or command...</span>
        <kbd className="hidden sm:inline ml-2 px-1.5 py-0.5 text-[10px] bg-accent rounded">Ctrl+K</kbd>
      </button>

      <button
        onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
        className="p-2 text-text-secondary hover:text-foreground hover:bg-accent rounded transition-colors"
        title="Toggle theme"
      >
        {theme === 'dark' ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
      </button>

      <div className="relative" ref={notifRef}>
        <button
          onClick={() => setShowNotifications(!showNotifications)}
          className="p-2 text-text-secondary hover:text-foreground hover:bg-accent rounded transition-colors relative"
        >
          <Bell className="w-4 h-4" />
          <span className="absolute top-1 right-1 w-2 h-2 bg-danger rounded-full" />
        </button>

        {showNotifications && (
          <div className="absolute right-0 top-full mt-2 w-80 bg-surface border border-border rounded-card shadow-xl z-50">
            <div className="p-3 border-b border-border">
              <h3 className="font-semibold text-sm">Notifications</h3>
            </div>
            <div className="max-h-64 overflow-y-auto">
              <div className="p-4 text-center text-text-secondary text-sm">
                No new notifications
              </div>
            </div>
            <Link
              to="/alerts"
              className="block p-3 text-center text-xs text-primary hover:bg-accent/50 border-t border-border"
              onClick={() => setShowNotifications(false)}
            >
              View all alerts
            </Link>
          </div>
        )}
      </div>

      <div className="relative" ref={userMenuRef}>
        <button
          onClick={() => setShowUserMenu(!showUserMenu)}
          className="flex items-center gap-2 p-1.5 hover:bg-accent rounded transition-colors"
        >
          <div className="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center">
            <span className="text-xs font-semibold text-primary">
              {user?.username?.charAt(0).toUpperCase() || 'U'}
            </span>
          </div>
        </button>

        {showUserMenu && (
          <div className="absolute right-0 top-full mt-2 w-56 bg-surface border border-border rounded-card shadow-xl z-50">
            <div className="p-3 border-b border-border">
              <p className="text-sm font-medium">{user?.username}</p>
              <p className="text-xs text-text-secondary">{user?.email}</p>
            </div>
            <div className="py-1">
              <Link
                to="/settings"
                className="flex items-center gap-3 px-3 py-2 text-sm text-text-secondary hover:text-foreground hover:bg-accent/50"
                onClick={() => setShowUserMenu(false)}
              >
                <User className="w-4 h-4" />
                Profile & Sessions
              </Link>
              <Link
                to="/settings/security"
                className="flex items-center gap-3 px-3 py-2 text-sm text-text-secondary hover:text-foreground hover:bg-accent/50"
                onClick={() => setShowUserMenu(false)}
              >
                <Shield className="w-4 h-4" />
                Security
              </Link>
            </div>
            <div className="border-t border-border py-1">
              <button
                onClick={logout}
                className="flex items-center gap-3 px-3 py-2 text-sm text-danger hover:bg-danger/10 w-full"
              >
                <LogOut className="w-4 h-4" />
                Sign out
              </button>
            </div>
          </div>
        )}
      </div>
    </header>
  )
}