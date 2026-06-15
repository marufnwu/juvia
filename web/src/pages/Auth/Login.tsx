import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Server, Eye, EyeOff, AlertCircle } from 'lucide-react'
import api from '../../lib/api'
import { useAuthStore } from '../../stores/authStore'
import { cn } from '../../lib/utils'
import { Input } from '../../components/ui/Input'
import { Button } from '../../components/ui/Button'

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
          if (res.data.data.csrf_token) {
            localStorage.setItem('csrf_token', res.data.data.csrf_token)
          }
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
                <Input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
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
                  <Input
                    type={showPassword ? 'text' : 'password'}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="pr-10"
                    placeholder="Enter password"
                    required
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-1 top-1/2 -translate-y-1/2 text-text-secondary hover:text-foreground"
                  >
                    {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </Button>
                </div>
              </div>

              {error && (
                <div className="flex items-center gap-2 p-3 bg-danger/10 border border-danger/20 rounded text-sm text-danger">
                  <AlertCircle className="w-4 h-4 flex-shrink-0" />
                  {error}
                </div>
              )}

              <Button
                type="submit"
                variant="primary"
                className="w-full"
                disabled={loading}
                loading={loading}
              >
                {loading ? 'Signing in...' : 'Sign in'}
              </Button>
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
                <Input
                  type="text"
                  value={twoFactorCode}
                  onChange={(e) => setTwoFactorCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                  className="text-center tracking-widest font-mono"
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

              <Button
                type="submit"
                variant="primary"
                className="w-full"
                disabled={loading || twoFactorCode.length !== 6}
                loading={loading}
              >
                {loading ? 'Verifying...' : 'Verify'}
              </Button>

              <Button
                type="button"
                variant="ghost"
                className="w-full"
                onClick={() => {
                  setRequires2FA(false)
                  setTwoFactorCode('')
                  setError('')
                }}
              >
                Back to login
              </Button>
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