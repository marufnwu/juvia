import { useEffect, useState } from 'react'
import { Mail, ExternalLink, Trash2, CheckCircle, AlertCircle } from 'lucide-react'
import api from '../../lib/api'

export default function Webmail() {
  const [installed, setInstalled] = useState(false)
  const [loading, setLoading] = useState(true)
  const [installing, setInstalling] = useState(false)
  const [uninstalling, setUninstalling] = useState(false)
  const [webmailURL, setWebmailURL] = useState('')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetchStatus()
  }, [])

  const fetchStatus = async () => {
    try {
      const res = await api.get('/webmail/status')
      const data = res.data
      if (data.success) {
        setInstalled(data.data.installed)
      }

      const urlRes = await api.get('/webmail/url')
      const urlData = urlRes.data
      if (urlData.success) {
        setWebmailURL(urlData.data.url)
      }
    } catch (err) {
      console.error('Failed to fetch webmail status:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleInstall = async () => {
    setInstalling(true)
    setError(null)
    try {
      const res = await api.post('/webmail/install')
      const data = res.data
      if (data.success) {
        setInstalled(true)
        fetchStatus()
      } else {
        setError(data.error?.message || 'Failed to install webmail')
      }
    } catch (err: any) {
      setError(err?.response?.data?.error?.message || 'Failed to install webmail')
    } finally {
      setInstalling(false)
    }
  }

  const handleUninstall = async () => {
    if (!confirm('Are you sure you want to uninstall webmail? This will remove all Roundcube data.')) {
      return
    }
    setUninstalling(true)
    setError(null)
    try {
      const res = await api.post('/webmail/uninstall')
      const data = res.data
      if (data.success) {
        setInstalled(false)
        setWebmailURL('')
      } else {
        setError(data.error?.message || 'Failed to uninstall webmail')
      }
    } catch (err: any) {
      setError(err?.response?.data?.error?.message || 'Failed to uninstall webmail')
    } finally {
      setUninstalling(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-muted-foreground">Loading...</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Webmail</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Access your emails through a web browser — no email client needed
        </p>
      </div>

      {error && (
        <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-lg flex items-start gap-3">
          <AlertCircle className="h-5 w-5 text-red-500 shrink-0 mt-0.5" />
          <div>
            <p className="font-medium text-red-500">Error</p>
            <p className="text-sm text-red-400">{error}</p>
          </div>
        </div>
      )}

      <div className="border rounded-lg p-6">
        <div className="flex items-start gap-4">
          <div className="p-3 bg-primary/10 rounded-lg">
            <Mail className="h-8 w-8 text-primary" />
          </div>
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <h2 className="text-lg font-semibold">Roundcube Webmail</h2>
              {installed ? (
                <span className="flex items-center gap-1 text-xs text-green-500 bg-green-500/10 px-2 py-0.5 rounded">
                  <CheckCircle size={12} /> Installed
                </span>
              ) : (
                <span className="text-xs text-muted-foreground bg-muted px-2 py-0.5 rounded">
                  Not installed
                </span>
              )}
            </div>
            <p className="text-sm text-muted-foreground mt-1">
              Roundcube is a browser-based email client that lets you read, compose, and send emails
              from any device with a web browser. It's the easiest way to access your email when
              you're on the go.
            </p>

            {installed && webmailURL && (
              <a
                href={webmailURL}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-2 mt-4 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90"
              >
                <ExternalLink size={16} />
                Open Webmail
              </a>
            )}
          </div>
        </div>

        <div className="mt-6 pt-6 border-t">
          {!installed ? (
            <div>
              <p className="text-sm text-muted-foreground mb-4">
                Installing Roundcube will set up the webmail application on your server. This may take
                a few minutes.
              </p>
              <button
                onClick={handleInstall}
                disabled={installing}
                className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50"
              >
                {installing ? 'Installing...' : 'Install Roundcube'}
              </button>
            </div>
          ) : (
            <div className="flex items-center justify-between">
              <p className="text-sm text-muted-foreground">
                Roundcube is installed and ready to use.
              </p>
              <button
                onClick={handleUninstall}
                disabled={uninstalling}
                className="inline-flex items-center gap-2 px-4 py-2 text-red-500 border border-red-500 rounded-lg hover:bg-red-500/10 disabled:opacity-50"
              >
                <Trash2 size={16} />
                {uninstalling ? 'Uninstalling...' : 'Uninstall'}
              </button>
            </div>
          )}
        </div>
      </div>

      <div className="border rounded-lg p-4 bg-muted/50">
        <h3 className="font-medium text-sm mb-2">What is Webmail?</h3>
        <p className="text-sm text-muted-foreground">
          Webmail allows you to access your email accounts through a web browser (like Chrome, Firefox,
          or Safari) instead of using a desktop email application like Outlook or Apple Mail. This means
          you can check your email from any computer, tablet, or phone with an internet connection —
          no additional software required.
        </p>
      </div>
    </div>
  )
}