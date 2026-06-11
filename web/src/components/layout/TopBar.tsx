import { Bell, User } from 'lucide-react'

export default function TopBar() {
  return (
    <header className="h-14 border-b border-border flex items-center justify-between px-6 bg-card">
      <div className="text-sm text-muted-foreground">Server Panel</div>
      <div className="flex items-center gap-4">
        <button className="p-2 hover:bg-accent">
          <Bell size={16} />
        </button>
        <button className="p-2 hover:bg-accent">
          <User size={16} />
        </button>
      </div>
    </header>
  )
}
