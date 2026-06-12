import { create } from 'zustand'

interface UIState {
  sidebarCollapsed: boolean
  theme: 'dark' | 'light'
  commandPaletteOpen: boolean
  toasts: ToastItem[]
  toggleSidebar: () => void
  setSidebarCollapsed: (collapsed: boolean) => void
  setTheme: (theme: 'dark' | 'light') => void
  toggleCommandPalette: () => void
  setCommandPaletteOpen: (open: boolean) => void
}

interface ToastItem {
  id: string
  variant: 'success' | 'error' | 'warning' | 'info'
  title: string
  message?: string
}

export const useUIStore = create<UIState>((set) => ({
  sidebarCollapsed: false,
  theme: 'dark',
  commandPaletteOpen: false,
  toasts: [],

  toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
  setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),
  setTheme: (theme) => set({ theme }),
  toggleCommandPalette: () => set((state) => ({ commandPaletteOpen: !state.commandPaletteOpen })),
  setCommandPaletteOpen: (open) => set({ commandPaletteOpen: open }),
}))