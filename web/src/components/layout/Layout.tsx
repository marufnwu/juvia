import { Outlet } from 'react-router-dom'
import Sidebar from './Sidebar'
import TopBar from './TopBar'
import CommandPalette from '../CommandPalette'
import { useUIStore } from '../../stores/uiStore'
import { cn } from '../../lib/utils'

export default function Layout() {
  const { sidebarCollapsed } = useUIStore()

  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      <div className="flex flex-col flex-1 min-w-0">
        <TopBar />
        <main className="flex-1 overflow-auto p-6 bg-background">
          <Outlet />
        </main>
      </div>
      <CommandPalette />
    </div>
  )
}