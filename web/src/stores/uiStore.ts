import { create } from 'zustand'
import { persist } from 'zustand/middleware'

type Theme = 'dark' | 'light'

interface UIState {
  sidebarCollapsed: boolean
  theme: Theme
  commandPaletteOpen: boolean
  toggleSidebar: () => void
  setSidebarCollapsed: (collapsed: boolean) => void
  setTheme: (theme: Theme) => void
  toggleTheme: () => void
  toggleCommandPalette: () => void
  setCommandPaletteOpen: (open: boolean) => void
}

function applyThemeClass(theme: Theme) {
  const root = document.documentElement
  root.classList.remove('dark', 'light')
  root.classList.add(theme)
}

const initialTheme: Theme = 'dark'

export const useUIStore = create<UIState>()(
  persist(
    (set, get) => ({
      sidebarCollapsed: false,
      theme: initialTheme,
      commandPaletteOpen: false,

      toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
      setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),
      setTheme: (theme) => {
        applyThemeClass(theme)
        set({ theme })
      },
      toggleTheme: () => {
        const next = get().theme === 'dark' ? 'light' : 'dark'
        applyThemeClass(next)
        set({ theme: next })
      },
      toggleCommandPalette: () => set((state) => ({ commandPaletteOpen: !state.commandPaletteOpen })),
      setCommandPaletteOpen: (open) => set({ commandPaletteOpen: open }),
    }),
    {
      name: 'juvia-ui',
      partialize: (state) => ({ sidebarCollapsed: state.sidebarCollapsed, theme: state.theme }),
      onRehydrateStorage: () => (state) => {
        if (state?.theme) {
          applyThemeClass(state.theme)
        } else {
          applyThemeClass(initialTheme)
        }
      },
    }
  )
)

if (typeof window !== 'undefined') {
  applyThemeClass(useUIStore.getState().theme)
}