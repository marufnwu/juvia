import { useEffect, useState } from 'react'
import { Mail, ExternalLink, Download, Package, Trash2, RefreshCw, CheckCircle } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { PageHeader } from '../../components/ui/Misc'
import { ConfirmModal } from '../../components/ui/Modal'
import api from '../../lib/api'
import { cn } from '../../lib/utils'
import { useApiError } from '../../hooks/useToast'

export default function Webmail() {
  const showError = useApiError()
  const [status, setStatus] = useState<'installed' | 'not_installed' | 'installing'>('not_installed')
  const [webmailUrl, setWebmailUrl] = useState('')
  const [loading, setLoading] = useState(true)
  const [confirmUninstall, setConfirmUninstall] = useState(false)
  const [uninstalling, setUninstalling] = useState(false)

  useEffect(() => {
    loadStatus()
  }, [])

  const loadStatus = async () => {
    try {
      const res = await api.get('/webmail/status')
      setStatus(res.data.data?.installed ? 'installed' : 'not_installed')
      setWebmailUrl(res.data.data?.url || '')
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const handleInstall = async () => {
    setStatus('installing')
    try {
      await api.post('/webmail/install')
      loadStatus()
    } catch (err: any) {
      showError(err, 'Failed to install')
      setStatus('not_installed')
    }
  }

  const handleUninstall = () => {
    setConfirmUninstall(true)
  }

  const doUninstall = async () => {
    setConfirmUninstall(false)
    setUninstalling(true)
    try {
      await api.post('/webmail/uninstall')
      loadStatus()
    } catch (err: any) {
      showError(err, 'Failed to uninstall')
    } finally {
      setUninstalling(false)
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Webmail"
        description="Browser-based email client (Roundcube)"
        breadcrumbs={[
          { label: 'Email', href: '/email' },
          { label: 'Webmail' },
        ]}
      />

      {loading ? (
        <Card className="flex items-center justify-center h-48">
          <RefreshCw className="w-6 h-6 animate-spin text-text-secondary" />
        </Card>
      ) : status === 'installed' ? (
        <div className="space-y-6">
          <Card>
            <div className="flex items-center gap-4">
              <div className="p-3 bg-success/10 rounded">
                <CheckCircle className="w-8 h-8 text-success" />
              </div>
              <div className="flex-1">
                <CardTitle>Webmail Installed</CardTitle>
                <CardDescription>Roundcube is ready to use</CardDescription>
              </div>
              <Button
                onClick={() => window.open(webmailUrl, '_blank')}
              >
                <ExternalLink className="w-4 h-4 mr-2" />
                Open Webmail
              </Button>
            </div>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Actions</CardTitle>
            </CardHeader>
            <div className="flex gap-3">
              <Button
                variant="outline"
                onClick={() => window.open(webmailUrl, '_blank')}
              >
                <ExternalLink className="w-4 h-4 mr-2" />
                Open Webmail
              </Button>
              <Button
                variant="outline"
                onClick={handleUninstall}
              >
                <Trash2 className="w-4 h-4 mr-2" />
                Uninstall
              </Button>
            </div>
          </Card>
        </div>
      ) : status === 'installing' ? (
        <Card className="flex items-center justify-center h-48">
          <div className="text-center">
            <RefreshCw className="w-8 h-8 mx-auto animate-spin text-primary mb-3" />
            <p className="text-text-secondary">Installing webmail...</p>
          </div>
        </Card>
      ) : (
        <div className="space-y-6">
          <Card>
            <div className="flex items-center gap-4">
              <div className="p-3 bg-accent/50 rounded">
                <Mail className="w-8 h-8 text-text-secondary" />
              </div>
              <div className="flex-1">
                <CardTitle>Webmail Not Installed</CardTitle>
                <CardDescription>
                  Install Roundcube to access email through your browser. Requires at least 512MB RAM.
                </CardDescription>
              </div>
              <Button onClick={handleInstall}>
                <Package className="w-4 h-4 mr-2" />
                Install Webmail
              </Button>
            </div>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>About Webmail</CardTitle>
            </CardHeader>
            <div className="space-y-3 text-sm text-text-secondary">
              <p>
                Webmail provides a full-featured email client accessible through your browser.
                It's ideal for checking email when you don't have a desktop email client configured.
              </p>
              <p>
                We install Roundcube, a popular open-source webmail application that works with
                your existing mail server configuration.
              </p>
            </div>
          </Card>
        </div>
      )}

      <ConfirmModal
        open={confirmUninstall}
        onClose={() => setConfirmUninstall(false)}
        onConfirm={doUninstall}
        title="Uninstall Webmail"
        description="Remove the browser-based email client (Roundcube)? This cannot be undone."
        confirmLabel="Uninstall"
        variant="danger"
        loading={uninstalling}
      />
    </div>
  )
}