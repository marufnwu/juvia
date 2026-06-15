import { useEffect, useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, Globe, Database, Lock, Check, AlertTriangle, Server, Copy } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { PageHeader } from '../../components/ui/Misc'
import { Modal } from '../../components/ui/Modal'
import api from '../../lib/api'
import { Input, Switch } from '../../components/ui/Input'
import { cn } from '../../lib/utils'

export default function CreateWebsite() {
  const navigate = useNavigate()
  const [step, setStep] = useState(1)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [createdWebsite, setCreatedWebsite] = useState<{ id: number; domain: string } | null>(null)
  const [serverIP, setServerIP] = useState('')
  const [nameservers, setNameservers] = useState<string[]>([])

  useEffect(() => {
    loadNetworkInfo()
  }, [])

  const loadNetworkInfo = async () => {
    try {
      const [ipRes, nsRes] = await Promise.allSettled([
        api.get('/server/ip'),
        api.get('/dns/nameservers'),
      ])
      if (ipRes.status === 'fulfilled') {
        setServerIP(ipRes.value.data.data?.ip || '')
      }
      if (nsRes.status === 'fulfilled') {
        const nsData = nsRes.value.data.data
        setNameservers([nsData?.ns1_hostname, nsData?.ns2_hostname].filter(Boolean))
      }
    } catch {
      // ignore
    }
  }
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
        const domain = res.data.data?.domain || res.data.data?.website?.domain || formData.domain
        setCreatedWebsite({ id: websiteId, domain })
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
              <label className="block text-sm font-medium mb-1.5">Domain Name · Your website address (e.g. example.com)</label>
              <Input
                type="text"
                value={formData.domain}
                onChange={(e) => {
                  setFormData({ ...formData, domain: e.target.value })
                  if (validation.errors.length) validateDomain(e.target.value)
                }}
                placeholder="example.com"
                error={validation.errors.length > 0}
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
              <label className="block text-sm font-medium mb-1.5">Document Root · Where website files are stored</label>
              <Input
                type="text"
                value={formData.document_root}
                onChange={(e) => setFormData({ ...formData, document_root: e.target.value })}
                placeholder="/public_html"
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
              <label className="block text-sm font-medium mb-3">Web Server · Software serving pages</label>
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
              <label className="block text-sm font-medium mb-3">PHP Version · The language your code runs on</label>
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
                PHP-FPM · Runs PHP code efficiently
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
            <Switch
              checked={formData.enable_ssl}
              onChange={(e) => setFormData({ ...formData, enable_ssl: e.target.checked })}
              label="Enable SSL · Add security certificate for HTTPS"
            />

            <Switch
              checked={formData.create_database}
              onChange={(e) => setFormData({ ...formData, create_database: e.target.checked })}
              label="Create Database"
            />

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

      {createdWebsite && (
        <Modal open={!!createdWebsite} onClose={() => navigate(`/websites/${createdWebsite.id}`)} title="Website Created" size="lg">
          <div className="space-y-5">
            <div className="p-4 bg-success/10 border border-success/20 rounded flex items-center gap-3">
              <Check className="w-5 h-5 text-success flex-shrink-0" />
              <div>
                <p className="font-medium text-foreground">{createdWebsite.domain} is ready</p>
                <p className="text-sm text-text-secondary">Your website has been created on the server.</p>
              </div>
            </div>

            {serverIP && (
              <div className="space-y-3">
                <h4 className="text-sm font-medium flex items-center gap-2">
                  <Server className="w-4 h-4" />
                  Point your domain to this server
                </h4>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="space-y-1">
                    <p className="text-sm text-text-secondary">A Record for @</p>
                    <div className="flex items-center gap-2">
                      <code className="px-2 py-1 bg-accent rounded text-sm font-mono flex-1">{serverIP}</code>
                      <Button variant="ghost" size="sm" onClick={() => navigator.clipboard.writeText(serverIP)}>
                        <Copy className="w-3.5 h-3.5" />
                      </Button>
                    </div>
                  </div>
                  {nameservers.length > 0 && (
                    <div className="space-y-1">
                      <p className="text-sm text-text-secondary">Or use nameservers</p>
                      <div className="flex flex-wrap gap-2">
                        {nameservers.map((ns) => (
                          <code key={ns} className="px-2 py-1 bg-accent rounded text-sm font-mono">{ns}</code>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}

            <div className="flex justify-end gap-3 pt-2">
              <Link to={`/websites/${createdWebsite.id}`}>
                <Button>Go to Website</Button>
              </Link>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}