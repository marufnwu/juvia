import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'

interface Action {
  id: string
  label: string
  shortcut: string
  action: () => void
}

export default function CommandPalette() {
  const [isOpen, setIsOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [selectedIndex, setSelectedIndex] = useState(0)
  const navigate = useNavigate()

  const actions: Action[] = [
    { id: 'dashboard', label: 'Go to Dashboard', shortcut: '→ Dashboard', action: () => navigate('/dashboard') },
    { id: 'websites', label: 'View Websites', shortcut: '→ Websites', action: () => navigate('/websites') },
    { id: 'add-website', label: 'Add New Website', shortcut: '→ Websites', action: () => navigate('/websites/create') },
    { id: 'databases', label: 'View Databases', shortcut: '→ Databases', action: () => navigate('/databases') },
    { id: 'create-database', label: 'Create Database', shortcut: '→ Databases', action: () => navigate('/databases/create') },
    { id: 'email', label: 'View Email', shortcut: '→ Email', action: () => navigate('/email') },
    { id: 'files', label: 'File Manager', shortcut: '→ Files', action: () => navigate('/files') },
    { id: 'firewall', label: 'View Firewall', shortcut: '→ Firewall', action: () => navigate('/firewall') },
    { id: 'backups', label: 'View Backups', shortcut: '→ Backups', action: () => navigate('/backups') },
    { id: 'cron', label: 'View Cron Jobs', shortcut: '→ Cron Jobs', action: () => navigate('/cron') },
    { id: 'alerts', label: 'View Alerts', shortcut: '→ Alerts', action: () => navigate('/alerts') },
    { id: 'settings', label: 'Open Settings', shortcut: '→ Settings', action: () => navigate('/settings') },
    { id: 'terminal', label: 'Terminal', shortcut: '→ Terminal', action: () => navigate('/terminal') },
    { id: 'terminal-recordings', label: 'Terminal Recordings', shortcut: '→ Terminal', action: () => navigate('/terminal/recordings') },
  ]

  const filteredActions = actions.filter((action) =>
    action.label.toLowerCase().includes(query.toLowerCase())
  )

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault()
      setIsOpen(true)
    }
    if (e.key === 'Escape') {
      setIsOpen(false)
    }
  }, [])

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  useEffect(() => {
    if (isOpen) {
      setQuery('')
      setSelectedIndex(0)
    }
  }, [isOpen])

  const handleSelect = (action: Action) => {
    action.action()
    setIsOpen(false)
  }

  const handleListKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setSelectedIndex((i) => Math.min(i + 1, filteredActions.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setSelectedIndex((i) => Math.max(i - 1, 0))
    } else if (e.key === 'Enter' && filteredActions[selectedIndex]) {
      handleSelect(filteredActions[selectedIndex])
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50">
      <div
        className="absolute inset-0 bg-black/50"
        onClick={() => setIsOpen(false)}
      />
      <div className="absolute top-[20%] left-1/2 -translate-x-1/2 w-full max-w-lg">
        <div className="bg-background border rounded-lg shadow-lg overflow-hidden">
          <div className="p-3 border-b">
            <input
              type="text"
              placeholder="Type a command or search..."
              value={query}
              onChange={(e) => {
                setQuery(e.target.value)
                setSelectedIndex(0)
              }}
              onKeyDown={handleListKeyDown}
              className="w-full px-3 py-2 bg-background outline-none"
              autoFocus
            />
          </div>
          <div className="max-h-80 overflow-auto">
            {filteredActions.length === 0 ? (
              <div className="p-4 text-center text-muted-foreground">
                No commands found
              </div>
            ) : (
              <div className="py-2">
                {filteredActions.map((action, index) => (
                  <button
                    key={action.id}
                    onClick={() => handleSelect(action)}
                    className={`w-full px-4 py-2 text-left flex items-center justify-between hover:bg-accent ${
                      index === selectedIndex ? 'bg-accent' : ''
                    }`}
                  >
                    <span>{action.label}</span>
                    <span className="text-xs text-muted-foreground">{action.shortcut}</span>
                  </button>
                ))}
              </div>
            )}
          </div>
          <div className="p-2 border-t bg-muted/50 text-xs text-muted-foreground flex gap-4">
            <span>↑↓ Navigate</span>
            <span>↵ Select</span>
            <span>Esc Close</span>
          </div>
        </div>
      </div>
    </div>
  )
}