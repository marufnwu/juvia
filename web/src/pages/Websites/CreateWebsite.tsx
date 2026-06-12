import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, Globe, Database, Lock, Check, AlertTriangle } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { PageHeader } from '../../components/ui/Misc'
import api from '../../lib/api'
import { cn } from '../../lib/utils'

export default function CreateWebsite() {
  const navigate = useNavigate()
  const [step, setStep] = useState(1)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [formData, setFormData] = useState({
    domain: '',
    php_version: '8.2',
    web_server: 'nginx',
    create_database: false,
    enable_ssl: true,
    document_root: '',
  })
  const [validation, setValidation] = useState<{ valid: boolean; errors: string[] }>({
    valid: true,
    errors: [],
  })

  const validateDomain = (domain: string): boolean => {
    const errors: string[] = []
    if (!domain) {
      errors.push('Domain is required')
    } else if (!/^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/.test(domain)) {
      errors.push('Invalid domain format')
    }
    setValidation({ valid: errors.length === 0, errors })
    return errors.length === 0
  }

  const handleSubmit = async () => {
    if (!validateDomain(formData.domain)) return

    setLoading(true)
    setError('')

    try {
      const res = await api.post('/websites', {
        domain: formData.domain,
        php_version: formData.php_version,
        web_server: formData.web_server,
        document_root: formData.document_root || undefined,
        enable_ssl: formData.enable_ssl,
        create_database: formData.create_database,
      })

      if (res.data.success) {
        const websiteId = res.data.data?.id || res.data.data?.website?.id
        navigate(`/websites/${websiteId}`)
      }
    } catch (err: any) {
      setError(err.response?.data?.error?.user_message || 'Failed to create website')
    } finally {
      setLoading(false)
    }
  }

  const steps = [
    { id: 1, label: 'Domain' },
    { id: 2, label: 'Runtime' },
    { id: 3, label: 'Options' },
    { id: 4, label: 'Review' },
  ]

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <PageHeader
        title="Create Website"
        description="Add a new website to your server"
        breadcrumbs={[
          { label: 'Websites', href: '/websites' },
          { label: 'Create' },
        ]}
        actions={
          <Link
            to="/websites"
            className="inline-flex items-center gap-2 px-4 py-2 border border-border text-sm rounded hover:bg-accent transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            Cancel
          </Link>
        }
      />

      <div className="flex items-center gap-4 mb-6">
        {steps.map((s, i) => (
          <div key={s.id} className="flex items-center">
            <button
              onClick={() => step > s.id && setStep(s.id)}
              className={cn(
                'w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium transition-colors',
                step >= s.id
                  ? 'bg-primary text-white'
                  : 'bg-accent text-text-secondary hover:bg-accent/70'
              )}
            >
              {step > s.id ? <Check className="w-4 h-4" /> : s.id}
            </button>
            <span className={cn('ml-2 text-sm', step >= s.id ? 'text-foreground' : 'text-text-secondary')}>
              {s.label}
            </span>
            {i < steps.length - 1 && <div className="w-8 h-px bg-border ml-4" />}
          </div>
        ))}
      </div>

      {error && (
        <div className="p-4 bg-danger/10 border border-danger/20 rounded flex items-center gap-3">
          <AlertTriangle className="w-4 h-4 text-danger flex-shrink-0" />
          <span className="text-sm text-danger">{error}</span>
        </div>
      )}

      <Card>
        {step === 1 && (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1.5">Domain Name</label>
              <input
                type="text"
                value={formData.domain}
                onChange={(e) => {
                  setFormData({ ...formData, domain: e.target.value })
                  if (validation.errors.length) validateDomain(e.target.value)
                }}
                placeholder="example.com"
                className={cn(
                  'w-full h-10 px-3 bg-background border rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary/50',
                  validation.errors.length ? 'border-danger' : 'border-border'
                )}
                autoFocus
              />
              <p className="text-xs text-text-secondary mt-1.5">
                The primary domain for this website
              </p>
              {validation.errors.length > 0 && (
                <div className="mt-2 space-y-1">
                  {validation.errors.map((err, i) => (
                    <p key={i} className="text-xs text-danger">{err}</p>
                  ))}
                </div>
              )}
            </div>

            <div>
              <label className="block text-sm font-medium mb-1.5">Document Root (optional)</label>
              <input
                type="text"
                value={formData.document_root}
                onChange={(e) => setFormData({ ...formData, document_root: e.target.value })}
                placeholder="/public_html"
                className="w-full h-10 px-3 bg-background border border-border rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
              />
              <p className="text-xs text-text-secondary mt-1.5">
                Leave empty to use default public_html
              </p>
            </div>

            <Button onClick={() => validateDomain(formData.domain) && setStep(2)} className="w-full">
              Continue
            </Button>
          </div>
        )}

        {step === 2 && (
          <div className="space-y-6">
            <div>
              <label className="block text-sm font-medium mb-3">Web Server</label>
              <div className="grid grid-cols-2 gap-3">
                {['nginx', 'apache'].map((server) => (
                  <button
                    key={server}
                    onClick={() => setFormData({ ...formData, web_server: server })}
                    className={cn(
                      'p-4 border rounded text-center transition-colors',
                      formData.web_server === server
                        ? 'border-primary bg-primary/10 text-primary'
                        : 'border-border hover:border-primary/50'
                    )}
                  >
                    <Globe className="w-6 h-6 mx-auto mb-2" />
                    <span className="font-medium capitalize">{server}</span>
                  </button>
                ))}
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium mb-3">PHP Version</label>
              <div className="grid grid-cols-4 gap-2">
                {['8.3', '8.2', '8.1', '8.0'].map((version) => (
                  <button
                    key={version}
                    onClick={() => setFormData({ ...formData, php_version: version })}
                    className={cn(
                      'p-3 border rounded text-center transition-colors',
                      formData.php_version === version
                        ? 'border-primary bg-primary/10 text-primary'
                        : 'border-border hover:border-primary/50'
                    )}
                  >
                    <span className="font-medium">PHP {version}</span>
                  </button>
                ))}
              </div>
              <p className="text-xs text-text-secondary mt-2">
                PHP-FPM pool will be created with the selected version
              </p>
            </div>

            <div className="flex gap-3">
              <Button variant="outline" onClick={() => setStep(1)}>Back</Button>
              <Button onClick={() => setStep(3)} className="flex-1">Continue</Button>
            </div>
          </div>
        )}

        {step === 3 && (
          <div className="space-y-6">
            <div>
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={formData.enable_ssl}
                  onChange={(e) => setFormData({ ...formData, enable_ssl: e.target.checked })}
                  className="w-4 h-4 rounded border-border"
                />
                <div>
                  <span className="font-medium">Enable SSL</span>
                  <p className="text-xs text-text-secondary">Issue a free Let's Encrypt certificate</p>
                </div>
              </label>
            </div>

            <div>
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={formData.create_database}
                  onChange={(e) => setFormData({ ...formData, create_database: e.target.checked })}
                  className="w-4 h-4 rounded border-border"
                />
                <div>
                  <span className="font-medium">Create Database</span>
                  <p className="text-xs text-text-secondary">Create a MySQL database for this website</p>
                </div>
              </label>
            </div>

            <div className="flex gap-3">
              <Button variant="outline" onClick={() => setStep(2)}>Back</Button>
              <Button onClick={() => setStep(4)} className="flex-1">Continue</Button>
            </div>
          </div>
        )}

        {step === 4 && (
          <div className="space-y-6">
            <Card className="bg-accent/30">
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-text-secondary">Domain</span>
                  <span className="font-medium">{formData.domain || '—'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-text-secondary">Web Server</span>
                  <span className="font-medium capitalize">{formData.web_server}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-text-secondary">PHP Version</span>
                  <span className="font-medium">PHP {formData.php_version}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-text-secondary">SSL</span>
                  <span className="font-medium">{formData.enable_ssl ? 'Enabled' : 'Disabled'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-text-secondary">Database</span>
                  <span className="font-medium">{formData.create_database ? 'Will be created' : 'None'}</span>
                </div>
              </div>
            </Card>

            <div className="flex gap-3">
              <Button variant="outline" onClick={() => setStep(3)}>Back</Button>
              <Button
                onClick={handleSubmit}
                loading={loading}
                className="flex-1"
              >
                Create Website
              </Button>
            </div>
          </div>
        )}
      </Card>
    </div>
  )
}