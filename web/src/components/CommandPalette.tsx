import React, { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { Search, Globe, Database, Mail, Terminal, Server, Plus, Settings, LogOut } from 'lucide-react'
import { cn } from '../lib/utils'
import { useUIStore } from '../stores/uiStore'
import { useAuthStore } from '../stores/authStore'

interface Command {
  id: string
  label: string
  description?: string
  icon: React.ReactNode
  action: () => void
  category: string
}

export default function CommandPalette() {
  const navigate = useNavigate()
  const { commandPaletteOpen, setCommandPaletteOpen } = useUIStore()
  const logout = useAuthStore((s) => s.logout)
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<Command[]>([])
  const [selectedIndex, setSelectedIndex] = useState(0)

  const commands: Command[] = [
    {
      id: 'dashboard',
      label: 'Go to Dashboard',
      description: 'View server overview',
      icon: <Server className="w-4 h-4" />,
      action: () => navigate('/dashboard'),
      category: 'Navigation',
    },
    {
      id: 'websites',
      label: 'Go to Websites',
      description: 'Manage websites',
      icon: <Globe className="w-4 h-4" />,
      action: () => navigate('/websites'),
      category: 'Navigation',
    },
    {
      id: 'databases',
      label: 'Go to Databases',
      description: 'Manage databases',
      icon: <Database className="w-4 h-4" />,
      action: () => navigate('/databases'),
      category: 'Navigation',
    },
    {
      id: 'email',
      label: 'Go to Email',
      description: 'Manage mailboxes',
      icon: <Mail className="w-4 h-4" />,
      action: () => navigate('/email'),
      category: 'Navigation',
    },
    {
      id: 'terminal',
      label: 'Open Terminal',
      description: 'Access server terminal',
      icon: <Terminal className="w-4 h-4" />,
      action: () => navigate('/terminal'),
      category: 'Navigation',
    },
    {
      id: 'create-website',
      label: 'Create Website',
      description: 'Add a new website',
      icon: <Plus className="w-4 h-4" />,
      action: () => navigate('/websites/create'),
      category: 'Actions',
    },
    {
      id: 'create-database',
      label: 'Create Database',
      description: 'Add a new database',
      icon: <Plus className="w-4 h-4" />,
      action: () => navigate('/databases/create'),
      category: 'Actions',
    },
    {
      id: 'settings',
      label: 'Go to Settings',
      description: 'Panel configuration',
      icon: <Settings className="w-4 h-4" />,
      action: () => navigate('/settings'),
      category: 'Navigation',
    },
    {
      id: 'logout',
      label: 'Sign Out',
      description: 'Log out of panel',
      icon: <LogOut className="w-4 h-4" />,
      action: logout,
      category: 'Actions',
    },
  ]

  useEffect(() => {
    if (commandPaletteOpen) {
      setQuery('')
      setResults(commands)
      setSelectedIndex(0)
    }
  }, [commandPaletteOpen])

  useEffect(() => {
    if (!query) {
      setResults(commands)
      return
    }
    const filtered = commands.filter(
      (c) =>
        c.label.toLowerCase().includes(query.toLowerCase()) ||
        c.description?.toLowerCase().includes(query.toLowerCase())
    )
    setResults(filtered)
    setSelectedIndex(0)
  }, [query])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Escape') {
        setCommandPaletteOpen(false)
      } else if (e.key === 'ArrowDown') {
        e.preventDefault()
        setSelectedIndex((i) => Math.min(i + 1, results.length - 1))
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        setSelectedIndex((i) => Math.max(i - 1, 0))
      } else if (e.key === 'Enter' && results[selectedIndex]) {
        e.preventDefault()
        results[selectedIndex].action()
        setCommandPaletteOpen(false)
      }
    },
    [results, selectedIndex, setCommandPaletteOpen]
  )

  useEffect(() => {
    const handleGlobalKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault()
        setCommandPaletteOpen(!commandPaletteOpen)
      }
    }
    document.addEventListener('keydown', handleGlobalKeyDown)
    return () => document.removeEventListener('keydown', handleGlobalKeyDown)
  }, [commandPaletteOpen, setCommandPaletteOpen])

  if (!commandPaletteOpen) return null

  const groupedResults = results.reduce((acc, cmd) => {
    if (!acc[cmd.category]) acc[cmd.category] = []
    acc[cmd.category].push(cmd)
    return acc
  }, {} as Record<string, Command[]>)

  let globalIndex = 0

  return (
    <div className="fixed inset-0 z-[200] flex items-start justify-center pt-[15vh]">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={() => setCommandPaletteOpen(false)} />
      <div className="relative w-full max-w-lg bg-surface border border-border rounded-card shadow-2xl overflow-hidden">
        <div className="flex items-center gap-3 px-4 border-b border-border">
          <Search className="w-4 h-4 text-text-secondary" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Search or type a command..."
            className="flex-1 h-12 bg-transparent text-sm text-foreground placeholder:text-text-secondary focus:outline-none"
            autoFocus
          />
          <kbd className="px-2 py-1 text-[10px] bg-accent rounded text-text-secondary">ESC</kbd>
        </div>

        <div className="max-h-80 overflow-y-auto">
          {Object.entries(groupedResults).map(([category, cmds]) => (
            <div key={category}>
              <div className="px-4 py-2">
                <span className="text-[10px] font-semibold uppercase tracking-wider text-text-secondary">
                  {category}
                </span>
              </div>
              {cmds.map((cmd) => {
                const isSelected = globalIndex === selectedIndex
                const currentIndex = globalIndex
                globalIndex++
                return (
                  <button
                    key={cmd.id}
                    onClick={() => {
                      cmd.action()
                      setCommandPaletteOpen(false)
                    }}
                    className={cn(
                      'w-full flex items-center gap-3 px-4 py-2.5 text-sm transition-colors',
                      isSelected ? 'bg-primary/10 text-primary' : 'text-foreground hover:bg-accent/50'
                    )}
                  >
                    <span className={cn('text-text-secondary', isSelected && 'text-primary')}>
                      {cmd.icon}
                    </span>
                    <div className="text-left">
                      <span>{cmd.label}</span>
                      {cmd.description && (
                        <span className="ml-2 text-xs text-text-secondary">{cmd.description}</span>
                      )}
                    </div>
                  </button>
                )
              })}
            </div>
          ))}

          {results.length === 0 && (
            <div className="p-8 text-center text-text-secondary text-sm">
              No results found
            </div>
          )}
        </div>

        <div className="flex items-center gap-4 px-4 py-2 border-t border-border text-xs text-text-secondary">
          <span className="flex items-center gap-1">
            <kbd className="px-1.5 py-0.5 bg-accent rounded">↑↓</kbd> Navigate
          </span>
          <span className="flex items-center gap-1">
            <kbd className="px-1.5 py-0.5 bg-accent rounded">Enter</kbd> Select
          </span>
          <span className="flex items-center gap-1">
            <kbd className="px-1.5 py-0.5 bg-accent rounded">Esc</kbd> Close
          </span>
        </div>
      </div>
    </div>
  )
}