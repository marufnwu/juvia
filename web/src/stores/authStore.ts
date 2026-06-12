import { create } from 'zustand'
import api from '../lib/api'

interface User {
  id: number
  username: string
  email: string
  role: string
  two_factor_enabled: boolean
}

interface Session {
  id: string
  ip: string
  user_agent: string
  created_at: string
  last_active: string
}

interface AuthState {
  user: User | null
  accessToken: string | null
  isAuthenticated: boolean
  sessions: Session[]
  setUser: (user: User | null) => void
  setAccessToken: (token: string | null) => void
  setSessions: (sessions: Session[]) => void
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: localStorage.getItem('access_token'),
  isAuthenticated: !!localStorage.getItem('access_token'),
  sessions: [],

  setUser: (user) => set({ user }),

  setAccessToken: (token) => {
    if (token) {
      localStorage.setItem('access_token', token)
    } else {
      localStorage.removeItem('access_token')
    }
    set({ accessToken: token, isAuthenticated: !!token })
  },

  setSessions: (sessions) => set({ sessions }),

  logout: async () => {
    try {
      await api.post('/auth/logout')
    } catch (err) {
      console.error('Logout error:', err)
    }
    localStorage.removeItem('access_token')
    set({ user: null, accessToken: null, isAuthenticated: false, sessions: [] })
    window.location.href = '/login'
  },

  refreshUser: async () => {
    try {
      const res = await api.get('/auth/me')
      set({ user: res.data.data.user, sessions: res.data.data.sessions || [] })
    } catch (err) {
      console.error('Failed to refresh user:', err)
      get().logout()
    }
  },
}))