import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Server, Eye, EyeOff, AlertCircle, CheckCircle2, ArrowRight } from 'lucide-react'
import api from '../../lib/api'
import { cn } from '../../lib/utils'
import { Input } from '../../components/ui/Input'
import { Button } from '../../components/ui/Button'

interface SetupStatus {
  setup_required: boolean
}

export default function Setup() {
  const navigate = useNavigate()
  const [step, setStep] = useState<'check' | 'form' | 'done' | 'error'>('check')
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [serverName, setServerName] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    const checkStatus = async () => {
      try {
        const res = await api.get('/setup/status')
        const data = res.data.data as SetupStatus
        if (!data.setup_required) {
          navigate('/login', { replace: true })
        } else {
          setStep('form')
        }
      } catch {
        setStep('form')
      }
    }
    checkStatus()
  }, [navigate])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (password !== confirmPassword) {
      setError('Passwords do not match')
      return
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters')
      return
    }

    setLoading(true)

    try {
      await api.post('/setup/first-run', {
        username,
        email,
        password,
        server_name: serverName || undefined,
      })
      setStep('done')
      setTimeout(() => navigate('/login'), 2000)
    } catch (err: any) {
      setError(err.response?.data?.error?.user_message || 'Setup failed. Please try again.')
      setStep('error')
    } finally {
      setLoading(false)
    }
  }

  if (step === 'check') {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <div className="flex flex-col items-center gap-3">
          <div className="w-8 h-8 border-2 border-primary border-t-transparent rounded-full animate-spin" />
          <p className="text-sm text-text-secondary">Checking setup status...</p>
        </div>
      </div>
    )
  }

  if (step === 'done') {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <div className="w-full max-w-sm">
          <div className="flex flex-col items-center mb-8">
            <div className="w-12 h-12 rounded-lg bg-success/10 flex items-center justify-center mb-4">
              <CheckCircle2 className="w-6 h-6 text-success" />
            </div>
            <h1 className="text-xl font-semibold text-foreground">Setup Complete</h1>
            <p className="text-sm text-text-secondary mt-1">Redirecting to login...</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="w-full max-w-sm">
        <div className="flex flex-col items-center mb-8">
          <div className="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4">
            <Server className="w-6 h-6 text-primary" />
          </div>
          <h1 className="text-xl font-semibold text-foreground">Welcome to Juvia</h1>
          <p className="text-sm text-text-secondary mt-1">Create your admin account to get started</p>
        </div>

        <div className="bg-surface border border-border rounded-card p-6">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">
                Server Name
              </label>
              <Input
                type="text"
                value={serverName}
                onChange={(e) => setServerName(e.target.value)}
                placeholder="e.g. Production Server"
              />
              <p className="text-xs text-text-secondary mt-1">Optional — a friendly name for this server</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">
                Admin Username
              </label>
              <Input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Choose a username"
                required
                autoFocus
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">
                Email Address
              </label>
              <Input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="admin@example.com"
                required
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
                  placeholder="At least 8 characters"
                  required
                  minLength={8}
                  className="pr-10"
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-text-secondary hover:text-foreground"
                >
                  {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </Button>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">
                Confirm Password
              </label>
              <Input
                type={showPassword ? 'text' : 'password'}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="Repeat your password"
                required
                minLength={8}
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
              loading={loading}
            >
              {loading ? 'Setting up...' : 'Create Account'}
              {!loading && <ArrowRight className="w-4 h-4" />}
            </Button>
          </form>
        </div>

        <p className="text-center text-xs text-text-secondary mt-6">
          This account will have full administrative access
        </p>
      </div>
    </div>
  )
}
