import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Server, Eye, EyeOff, AlertCircle } from 'lucide-react'
import api from '../../lib/api'
import { useAuthStore } from '../../stores/authStore'
import { cn } from '../../lib/utils'

export default function Login() {
  const navigate = useNavigate()
  const setAccessToken = useAuthStore((s) => s.setAccessToken)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [requires2FA, setRequires2FA] = useState(false)
  const [twoFactorCode, setTwoFactorCode] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      if (requires2FA) {
        const res = await api.post('/auth/2fa/verify', { code: twoFactorCode })
        if (res.data.success) {
          setAccessToken(res.data.data.access_token)
          navigate('/dashboard')
        }
      } else {
        const res = await api.post('/auth/login', { username, password })
        if (res.data.data?.access_token) {
          setAccessToken(res.data.data.access_token)
          navigate('/dashboard')
        } else if (res.data.data?.requires_2fa) {
          setRequires2FA(true)
        }
      }
    } catch (err: any) {
      setError(err.response?.data?.error?.user_message || 'Login failed. Please check your credentials.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="w-full max-w-sm">
        <div className="flex flex-col items-center mb-8">
          <div className="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4">
            <Server className="w-6 h-6 text-primary" />
          </div>
          <h1 className="text-xl font-semibold text-foreground">Juvia</h1>
          <p className="text-sm text-text-secondary mt-1">Server Control Panel</p>
        </div>

        <div className="bg-surface border border-border rounded-card p-6">
          {!requires2FA ? (
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">
                  Username
                </label>
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  className="w-full h-10 px-3 bg-background border border-border rounded text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-colors"
                  placeholder="Enter username"
                  required
                  autoFocus
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">
                  Password
                </label>
                <div className="relative">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="w-full h-10 px-3 pr-10 bg-background border border-border rounded text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-colors"
                    placeholder="Enter password"
                    required
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-text-secondary hover:text-foreground"
                  >
                    {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
              </div>

              {error && (
                <div className="flex items-center gap-2 p-3 bg-danger/10 border border-danger/20 rounded text-sm text-danger">
                  <AlertCircle className="w-4 h-4 flex-shrink-0" />
                  {error}
                </div>
              )}

              <button
                type="submit"
                disabled={loading}
                className="w-full h-10 bg-primary text-white font-medium rounded hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {loading ? 'Signing in...' : 'Sign in'}
              </button>
            </form>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="text-center mb-4">
                <p className="text-sm text-text-secondary">
                  Enter the 6-digit code from your authenticator app
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">
                  2FA Code
                </label>
                <input
                  type="text"
                  value={twoFactorCode}
                  onChange={(e) => setTwoFactorCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                  className="w-full h-10 px-3 bg-background border border-border rounded text-sm text-foreground text-center tracking-widest font-mono focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-colors"
                  placeholder="000000"
                  maxLength={6}
                  required
                  autoFocus
                />
              </div>

              {error && (
                <div className="flex items-center gap-2 p-3 bg-danger/10 border border-danger/20 rounded text-sm text-danger">
                  <AlertCircle className="w-4 h-4 flex-shrink-0" />
                  {error}
                </div>
              )}

              <button
                type="submit"
                disabled={loading || twoFactorCode.length !== 6}
                className="w-full h-10 bg-primary text-white font-medium rounded hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {loading ? 'Verifying...' : 'Verify'}
              </button>

              <button
                type="button"
                onClick={() => {
                  setRequires2FA(false)
                  setTwoFactorCode('')
                  setError('')
                }}
                className="w-full text-sm text-text-secondary hover:text-foreground transition-colors"
              >
                Back to login
              </button>
            </form>
          )}
        </div>

        <p className="text-center text-xs text-text-secondary mt-6">
          Protected by two-factor authentication
        </p>
      </div>
    </div>
  )
}