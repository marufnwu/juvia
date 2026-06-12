import { useEffect, useState } from 'react'
import {
  Plus, Mail, CheckCircle, XCircle, AlertTriangle, Copy, RefreshCw,
  Users, Forward, AtSign, Shield, Globe
} from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge, StatusBadge } from '../../components/ui/Badge'
import { Tabs, TabPanel } from '../../components/ui/Tabs'
import { Table } from '../../components/ui/Table'
import { PageHeader } from '../../components/ui/Misc'
import { CopyButton } from '../../components/ui/Misc'
import { Modal } from '../../components/ui/Modal'
import { Input, Label, FormGroup, FormError } from '../../components/ui/Input'
import api from '../../lib/api'
import { formatDate, formatBytes } from '../../lib/utils'
import { cn } from '../../lib/utils'

interface Mailbox {
  id: number
  email: string
  display_name: string
  quota: number
  forward_to: string
  status: string
  created_at: string
}

interface Alias {
  id: number
  domain: string
  source: string
  destination: string
  created_at: string
}

interface Forwarder {
  id: number
  domain: string
  source: string
  destination: string
  created_at: string
}

interface DeliverabilityResult {
  spf: { status: boolean; record: string; message: string }
  dkim: { status: boolean; record: string; message: string }
  dmarc: { status: boolean; record: string; message: string }
  ptr: { status: boolean; ip: string; message: string }
}

export default function EmailDashboard() {
  const [activeTab, setActiveTab] = useState('mailboxes')
  const [mailboxes, setMailboxes] = useState<Mailbox[]>([])
  const [aliases, setAliases] = useState<Alias[]>([])
  const [forwarders, setForwarders] = useState<Forwarder[]>([])
  const [showModal, setShowModal] = useState(false)
  const [modalType, setModalType] = useState<'mailbox' | 'alias' | 'forwarder'>('mailbox')
  const [loading, setLoading] = useState(true)
  const [deliverabilityDomain, setDeliverabilityDomain] = useState('')
  const [deliverabilityResult, setDeliverabilityResult] = useState<DeliverabilityResult | null>(null)
  const [formData, setFormData] = useState({
    email: '', password: '', quota: 10737418240, display_name: '',
    domain: '', source: '', destination: '', forward_to: '',
  })
  const [formError, setFormError] = useState('')
  const [saving, setSaving] = useState(false)

  const tabs = [
    { id: 'mailboxes', label: 'Mailboxes', count: mailboxes.length },
    { id: 'aliases', label: 'Aliases', count: aliases.length },
    { id: 'forwarders', label: 'Forwarders', count: forwarders.length },
    { id: 'catchall', label: 'Catch-All' },
    { id: 'deliverability', label: 'Deliverability' },
  ]

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    try {
      const [mboxRes, aliasRes, fwdRes] = await Promise.allSettled([
        api.get('/email/mailboxes'),
        api.get('/email/aliases'),
        api.get('/email/forwarders'),
      ])
      if (mboxRes.status === 'fulfilled') setMailboxes(mboxRes.value.data.data || [])
      if (aliasRes.status === 'fulfilled') setAliases(aliasRes.value.data.data || [])
      if (fwdRes.status === 'fulfilled') setForwarders(fwdRes.value.data.data || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const openModal = (type: 'mailbox' | 'alias' | 'forwarder') => {
    setModalType(type)
    setFormError('')
    setFormData({ email: '', password: '', quota: 10737418240, display_name: '', domain: '', source: '', destination: '', forward_to: '' })
    setShowModal(true)
  }

  const handleCreate = async () => {
    setSaving(true)
    setFormError('')
    try {
      if (modalType === 'mailbox') {
        await api.post('/email/mailboxes', {
          email: formData.email,
          password: formData.password,
          quota: formData.quota,
          display_name: formData.display_name,
        })
      } else if (modalType === 'alias') {
        await api.post('/email/aliases', {
          domain: formData.domain,
          source: formData.source,
          destination: formData.destination,
        })
      } else {
        await api.post('/email/forwarders', {
          domain: formData.domain,
          source: formData.source,
          destination: formData.destination,
        })
      }
      setShowModal(false)
      loadData()
    } catch (err: any) {
      setFormError(err.response?.data?.error?.user_message || 'Failed to create')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (type: string, id: number) => {
    if (!confirm('Delete this item?')) return
    try {
      if (type === 'mailbox') await api.delete(`/email/mailboxes/${id}`)
      else if (type === 'alias') await api.delete(`/email/aliases/${id}`)
      else await api.delete(`/email/forwarders/${id}`)
      loadData()
    } catch (err) {
      console.error(err)
    }
  }

  const checkDeliverability = async () => {
    if (!deliverabilityDomain) return
    try {
      const res = await api.get(`/email/deliverability?domain=${encodeURIComponent(deliverabilityDomain)}`)
      setDeliverabilityResult(res.data.data)
    } catch (err) {
      console.error(err)
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Email"
        description="Manage mailboxes, aliases, and email delivery"
        breadcrumbs={[{ label: 'Email' }]}
      />

      <Tabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />

      {activeTab === 'mailboxes' && (
        <Card padding="none">
          <div className="p-4 flex items-center justify-between border-b border-border">
            <p className="text-sm text-text-secondary">Mailbox · An email account that stores messages on your server</p>
            <Button onClick={() => openModal('mailbox')}>
              <Plus className="w-4 h-4 mr-2" />
              Create Mailbox
            </Button>
          </div>
          <table className="w-full">
            <thead className="bg-accent/50">
              <tr>
                <th className="text-left p-3 text-xs font-medium text-text-secondary uppercase">Email</th>
                <th className="text-left p-3 text-xs font-medium text-text-secondary uppercase">Display Name</th>
                <th className="text-left p-3 text-xs font-medium text-text-secondary uppercase">Quota</th>
                <th className="text-left p-3 text-xs font-medium text-text-secondary uppercase">Status</th>
                <th className="text-left p-3 text-xs font-medium text-text-secondary uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {mailboxes.map((m) => (
                <tr key={m.id} className="hover:bg-accent/30 transition-colors">
                  <td className="p-3">
                    <div className="flex items-center gap-2">
                      <Mail className="w-4 h-4 text-text-secondary" />
                      <span className="font-medium">{m.email}</span>
                    </div>
                  </td>
                  <td className="p-3 text-text-secondary">{m.display_name || '—'}</td>
                  <td className="p-3 text-text-secondary">{formatBytes(m.quota)}</td>
                  <td className="p-3"><StatusBadge status={m.status} /></td>
                  <td className="p-3">
                    <button onClick={() => handleDelete('mailbox', m.id)} className="text-xs text-danger hover:underline">Delete</button>
                  </td>
                </tr>
              ))}
              {mailboxes.length === 0 && (
                <tr>
                  <td colSpan={5} className="p-8 text-center text-text-secondary text-sm">No mailboxes</td>
                </tr>
              )}
            </tbody>
          </table>
        </Card>
      )}

      {activeTab === 'aliases' && (
        <Card padding="none">
          <div className="p-4 flex items-center justify-between border-b border-border">
            <p className="text-sm text-text-secondary">Alias · Redirects emails from one address to another on the same domain</p>
            <Button onClick={() => openModal('alias')}>
              <Plus className="w-4 h-4 mr-2" />
              Create Alias
            </Button>
          </div>
          <table className="w-full">
            <thead className="bg-accent/50">
              <tr>
                <th className="text-left p-3 text-xs font-medium text-text-secondary uppercase">Source</th>
                <th className="text-left p-3 text-xs font-medium text-text-secondary uppercase">Destination</th>
                <th className="text-left p-3 text-xs font-medium text-text-secondary uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {aliases.map((a) => (
                <tr key={a.id} className="hover:bg-accent/30 transition-colors">
                  <td className="p-3 font-mono text-sm">{a.source}@{a.domain}</td>
                  <td className="p-3 text-text-secondary">{a.destination}</td>
                  <td className="p-3">
                    <button onClick={() => handleDelete('alias', a.id)} className="text-xs text-danger hover:underline">Delete</button>
                  </td>
                </tr>
              ))}
              {aliases.length === 0 && (
                <tr>
                  <td colSpan={3} className="p-8 text-center text-text-secondary text-sm">No aliases</td>
                </tr>
              )}
            </tbody>
          </table>
        </Card>
      )}

      {activeTab === 'forwarders' && (
        <Card padding="none">
          <div className="p-4 flex items-center justify-between border-b border-border">
            <p className="text-sm text-text-secondary">Forwarder · Sends copies of emails to external addresses</p>
            <Button onClick={() => openModal('forwarder')}>
              <Plus className="w-4 h-4 mr-2" />
              Create Forwarder
            </Button>
          </div>
          <table className="w-full">
            <thead className="bg-accent/50">
              <tr>
                <th className="text-left p-3 text-xs font-medium text-text-secondary uppercase">Source</th>
                <th className="text-left p-3 text-xs font-medium text-text-secondary uppercase">Destination</th>
                <th className="text-left p-3 text-xs font-medium text-text-secondary uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {forwarders.map((f) => (
                <tr key={f.id} className="hover:bg-accent/30 transition-colors">
                  <td className="p-3 font-mono text-sm">{f.source}@{f.domain}</td>
                  <td className="p-3 text-text-secondary">{f.destination}</td>
                  <td className="p-3">
                    <button onClick={() => handleDelete('forwarder', f.id)} className="text-xs text-danger hover:underline">Delete</button>
                  </td>
                </tr>
              ))}
              {forwarders.length === 0 && (
                <tr>
                  <td colSpan={3} className="p-8 text-center text-text-secondary text-sm">No forwarders</td>
                </tr>
              )}
            </tbody>
          </table>
        </Card>
      )}

      {activeTab === 'catchall' && (
        <Card>
          <CardHeader>
            <CardTitle>Catch-All Address</CardTitle>
            <CardDescription>Catch-All · Receives email sent to any address at your domain that doesn't exist</CardDescription>
          </CardHeader>
          <div className="flex gap-3">
            <input
              type="text"
              placeholder="yourdomain.com"
              className="flex-1 h-10 px-3 bg-background border border-border rounded text-sm"
              value={deliverabilityDomain}
              onChange={(e) => setFormData({ ...formData, domain: e.target.value })}
            />
            <input
              type="text"
              placeholder="catchall@external.com"
              className="flex-1 h-10 px-3 bg-background border border-border rounded text-sm"
              value={formData.forward_to}
              onChange={(e) => setFormData({ ...formData, forward_to: e.target.value })}
            />
            <Button>Set Catch-All</Button>
          </div>
        </Card>
      )}

      {activeTab === 'deliverability' && (
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Email Deliverability Check</CardTitle>
              <CardDescription>Check if your domain can receive emails from major providers</CardDescription>
            </CardHeader>
            <div className="flex gap-3">
              <Input
                type="text"
                placeholder="yourdomain.com"
                value={deliverabilityDomain}
                onChange={(e) => setDeliverabilityDomain(e.target.value)}
                className="flex-1"
              />
              <Button onClick={checkDeliverability}>Check</Button>
            </div>
          </Card>

          {deliverabilityResult && (
            <div className="grid grid-cols-2 gap-4">
              {[
                { title: 'SPF', result: deliverabilityResult.spf },
                { title: 'DKIM', result: deliverabilityResult.dkim },
                { title: 'DMARC', result: deliverabilityResult.dmarc },
                { title: 'PTR', result: deliverabilityResult.ptr },
              ].map(({ title, result }) => (
                <Card key={title}>
                  <div className="flex items-center gap-3 mb-3">
                    {result.status ? (
                      <CheckCircle className="w-5 h-5 text-success" />
                    ) : (
                      <XCircle className="w-5 h-5 text-danger" />
                    )}
                    <span className="font-semibold">{title}</span>
                  </div>
                  <p className="text-sm text-text-secondary mb-2">{result.message}</p>
                  {'record' in result && result.record && (
                    <div className="flex items-center gap-2">
                      <code className="flex-1 text-xs bg-accent/50 p-2 rounded font-mono">{result.record}</code>
                      <CopyButton text={result.record} />
                    </div>
                  )}
                </Card>
              ))}
            </div>
          )}
        </div>
      )}

      <Modal
        open={showModal}
        onClose={() => setShowModal(false)}
        title={modalType === 'mailbox' ? 'Create Mailbox' : modalType === 'alias' ? 'Create Alias' : 'Create Forwarder'}
        size="sm"
      >
        <div className="space-y-4">
          {formError && (
            <div className="p-3 bg-danger/10 border border-danger/20 rounded text-sm text-danger">{formError}</div>
          )}

          {modalType === 'mailbox' && (
            <>
              <FormGroup>
                <Label>Email Address</Label>
                <input
                  type="email"
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  placeholder="user@domain.com"
                  className="w-full h-10 px-3 bg-background border border-border rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
                />
              </FormGroup>
              <FormGroup>
                <Label>Password</Label>
                <input
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  className="w-full h-10 px-3 bg-background border border-border rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
                />
              </FormGroup>
              <FormGroup>
                <Label>Display Name</Label>
                <input
                  type="text"
                  value={formData.display_name}
                  onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
                  placeholder="John Doe"
                  className="w-full h-10 px-3 bg-background border border-border rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
                />
              </FormGroup>
            </>
          )}

          {(modalType === 'alias' || modalType === 'forwarder') && (
            <>
              <FormGroup>
                <Label>Domain</Label>
                <input
                  type="text"
                  value={formData.domain}
                  onChange={(e) => setFormData({ ...formData, domain: e.target.value })}
                  placeholder="domain.com"
                  className="w-full h-10 px-3 bg-background border border-border rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
                />
              </FormGroup>
              <FormGroup>
                <Label>Source (local part)</Label>
                <input
                  type="text"
                  value={formData.source}
                  onChange={(e) => setFormData({ ...formData, source: e.target.value })}
                  placeholder="info"
                  className="w-full h-10 px-3 bg-background border border-border rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
                />
              </FormGroup>
              <FormGroup>
                <Label>Destination</Label>
                <input
                  type="text"
                  value={formData.destination}
                  onChange={(e) => setFormData({ ...formData, destination: e.target.value })}
                  placeholder={modalType === 'forwarder' ? 'external@example.com' : 'user@domain.com'}
                  className="w-full h-10 px-3 bg-background border border-border rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
                />
              </FormGroup>
            </>
          )}

          <div className="flex justify-end gap-3 pt-2">
            <Button variant="outline" onClick={() => setShowModal(false)}>Cancel</Button>
            <Button onClick={handleCreate} loading={saving}>
              Create {modalType === 'mailbox' ? 'Mailbox' : modalType === 'alias' ? 'Alias' : 'Forwarder'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}