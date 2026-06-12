import { cn } from '../../lib/utils'

interface TabsProps {
  tabs: { id: string; label: string; count?: number }[]
  activeTab: string
  onChange: (id: string) => void
  className?: string
}

export function Tabs({ tabs, activeTab, onChange, className }: TabsProps) {
  return (
    <div className={cn('flex border-b border-border', className)}>
      {tabs.map((tab) => (
        <button
          key={tab.id}
          onClick={() => onChange(tab.id)}
          className={cn(
            'px-4 py-2.5 text-sm font-medium border-b-2 transition-colors relative',
            activeTab === tab.id
              ? 'border-primary text-primary'
              : 'border-transparent text-text-secondary hover:text-foreground'
          )}
        >
          <span>{tab.label}</span>
          {tab.count !== undefined && (
            <span
              className={cn(
                'ml-2 px-1.5 py-0.5 text-xs rounded',
                activeTab === tab.id
                  ? 'bg-primary/10 text-primary'
                  : 'bg-accent text-text-secondary'
              )}
            >
              {tab.count}
            </span>
          )}
        </button>
      ))}
    </div>
  )
}

interface TabPanelProps {
  children: React.ReactNode
  className?: string
}

export function TabPanel({ children, className }: TabPanelProps) {
  return <div className={cn('mt-4', className)}>{children}</div>
}