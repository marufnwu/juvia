import { useEffect, useState } from 'react'
import { Plus, CheckCircle, XCircle } from 'lucide-react'
import api from '../../lib/api'

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
  spf: { status: boolean; record: string; message: string; recommendation: string }
  dkim: { status: boolean; record: string; message: string; recommendation: string }
  dmarc: { status: boolean; record: string; message: string; recommendation: string }
  ptr: { status: boolean; ip: string; message: string; recommendation: string }
}

export default function EmailDashboard() {
  const [activeTab, setActiveTab] = useState<'mailboxes' | 'aliases' | 'forwarders' | 'catchall' | 'deliverability'>('mailboxes')
  const [mailboxes, setMailboxes] = useState<Mailbox[]>([])
  const [aliases, setAliases] = useState<Alias[]>([])
  const [forwarders, setForwarders] = useState<Forwarder[]>([])
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [createType, setCreateType] = useState<'mailbox' | 'alias' | 'forwarder'>('mailbox')
  const [deliverabilityDomain, setDeliverabilityDomain] = useState('')
  const [deliverabilityResult, setDeliverabilityResult] = useState<DeliverabilityResult | null>(null)

  const [formData, setFormData] = useState({
    email: '',
    password: '',
    quota: 10737418240,
    display_name: '',
    domain: '',
    source: '',
    destination: '',
    forward_to: '',
  })

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    try {
      const [mboxRes, aliasRes, fwdRes] = await Promise.all([
        api.get('/email/mailboxes'),
        api.get('/email/aliases'),
        api.get('/email/forwarders'),
      ])
      setMailboxes(mboxRes.data.data || [])
      setAliases(aliasRes.data.data || [])
      setForwarders(fwdRes.data.data || [])
    } catch (err) {
      console.error(err)
    }
  }

  const createMailbox = async () => {
    try {
      await api.post('/email/mailboxes', {
        email: formData.email,
        password: formData.password,
        quota: formData.quota,
        display_name: formData.display_name,
      })
      setShowCreateModal(false)
      loadData()
    } catch (err) {
      console.error(err)
    }
  }

  const createAlias = async () => {
    try {
      await api.post('/email/aliases', {
        domain: formData.domain,
        source: formData.source,
        destination: formData.destination,
      })
      setShowCreateModal(false)
      loadData()
    } catch (err) {
      console.error(err)
    }
  }

  const createForwarder = async () => {
    try {
      await api.post('/email/forwarders', {
        domain: formData.domain,
        source: formData.source,
        destination: formData.destination,
      })
      setShowCreateModal(false)
      loadData()
    } catch (err) {
      console.error(err)
    }
  }

  const deleteMailbox = async (id: number) => {
    if (!confirm('Delete this mailbox?')) return
    try {
      await api.delete(`/email/mailboxes/${id}`)
      loadData()
    } catch (err) {
      console.error(err)
    }
  }

  const deleteAlias = async (id: number) => {
    if (!confirm('Delete this alias?')) return
    try {
      await api.delete(`/email/aliases/${id}`)
      loadData()
    } catch (err) {
      console.error(err)
    }
  }

  const deleteForwarder = async (id: number) => {
    if (!confirm('Delete this forwarder?')) return
    try {
      await api.delete(`/email/forwarders/${id}`)
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

  const openCreateModal = (type: 'mailbox' | 'alias' | 'forwarder') => {
    setCreateType(type)
    setFormData({ email: '', password: '', quota: 10737418240, display_name: '', domain: '', source: '', destination: '', forward_to: '' })
    setShowCreateModal(true)
  }

  const formatSize = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-lg font-semibold">Email</h1>
      </div>

      <div className="flex border-b border-border mb-6">
        <button
          onClick={() => setActiveTab('mailboxes')}
          className={`px-4 py-2 text-sm font-medium ${activeTab === 'mailboxes' ? 'border-b-2 border-primary' : 'text-muted-foreground'}`}
        >
          Mailboxes
        </button>
        <button
          onClick={() => setActiveTab('aliases')}
          className={`px-4 py-2 text-sm font-medium ${activeTab === 'aliases' ? 'border-b-2 border-primary' : 'text-muted-foreground'}`}
        >
          Aliases
        </button>
        <button
          onClick={() => setActiveTab('forwarders')}
          className={`px-4 py-2 text-sm font-medium ${activeTab === 'forwarders' ? 'border-b-2 border-primary' : 'text-muted-foreground'}`}
        >
          Forwarders
        </button>
        <button
          onClick={() => setActiveTab('catchall')}
          className={`px-4 py-2 text-sm font-medium ${activeTab === 'catchall' ? 'border-b-2 border-primary' : 'text-muted-foreground'}`}
        >
          Catch-All
        </button>
        <button
          onClick={() => setActiveTab('deliverability')}
          className={`px-4 py-2 text-sm font-medium ${activeTab === 'deliverability' ? 'border-b-2 border-primary' : 'text-muted-foreground'}`}
        >
          Deliverability
        </button>
      </div>

      {activeTab === 'mailboxes' && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <p className="text-sm text-muted-foreground">
              Mailbox · An email account that stores messages on your server
            </p>
            <button
              onClick={() => openCreateModal('mailbox')}
              className="flex items-center gap-2 bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90"
            >
              <Plus size={16} />
              Create Mailbox
            </button>
          </div>

          <div className="border border-border">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 border-b border-border">
                <tr>
                  <th className="text-left p-3 font-medium">Email</th>
                  <th className="text-left p-3 font-medium">Display Name</th>
                  <th className="text-left p-3 font-medium">Quota</th>
                  <th className="text-left p-3 font-medium">Forward To</th>
                  <th className="text-left p-3 font-medium">Status</th>
                  <th className="text-left p-3 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {mailboxes.map((m) => (
                  <tr key={m.id} className="border-b border-border">
                    <td className="p-3">{m.email}</td>
                    <td className="p-3 text-muted-foreground">{m.display_name || '—'}</td>
                    <td className="p-3 text-muted-foreground">{formatSize(m.quota)}</td>
                    <td className="p-3 text-muted-foreground">{m.forward_to || '—'}</td>
                    <td className="p-3">
                      <span className={`text-xs px-2 py-1 rounded ${m.status === 'active' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'}`}>
                        {m.status}
                      </span>
                    </td>
                    <td className="p-3">
                      <button onClick={() => deleteMailbox(m.id)} className="text-red-600 hover:underline text-xs">
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
                {mailboxes.length === 0 && (
                  <tr>
                    <td colSpan={6} className="p-8 text-center text-muted-foreground">No mailboxes</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {activeTab === 'aliases' && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <p className="text-sm text-muted-foreground">
              Alias · Redirects emails from one address to another on the same domain
            </p>
            <button
              onClick={() => openCreateModal('alias')}
              className="flex items-center gap-2 bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90"
            >
              <Plus size={16} />
              Create Alias
            </button>
          </div>

          <div className="border border-border">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 border-b border-border">
                <tr>
                  <th className="text-left p-3 font-medium">Source</th>
                  <th className="text-left p-3 font-medium">Destination</th>
                  <th className="text-left p-3 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {aliases.map((a) => (
                  <tr key={a.id} className="border-b border-border">
                    <td className="p-3">{a.source}@{a.domain}</td>
                    <td className="p-3 text-muted-foreground">{a.destination}</td>
                    <td className="p-3">
                      <button onClick={() => deleteAlias(a.id)} className="text-red-600 hover:underline text-xs">
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
                {aliases.length === 0 && (
                  <tr>
                    <td colSpan={3} className="p-8 text-center text-muted-foreground">No aliases</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {activeTab === 'forwarders' && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <p className="text-sm text-muted-foreground">
              Forwarder · Sends copies of emails to external addresses
            </p>
            <button
              onClick={() => openCreateModal('forwarder')}
              className="flex items-center gap-2 bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90"
            >
              <Plus size={16} />
              Create Forwarder
            </button>
          </div>

          <div className="border border-border">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 border-b border-border">
                <tr>
                  <th className="text-left p-3 font-medium">Source</th>
                  <th className="text-left p-3 font-medium">Destination</th>
                  <th className="text-left p-3 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {forwarders.map((f) => (
                  <tr key={f.id} className="border-b border-border">
                    <td className="p-3">{f.source}@{f.domain}</td>
                    <td className="p-3 text-muted-foreground">{f.destination}</td>
                    <td className="p-3">
                      <button onClick={() => deleteForwarder(f.id)} className="text-red-600 hover:underline text-xs">
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
                {forwarders.length === 0 && (
                  <tr>
                    <td colSpan={3} className="p-8 text-center text-muted-foreground">No forwarders</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {activeTab === 'catchall' && (
        <div>
          <p className="text-sm text-muted-foreground mb-4">
            Catch-All · Receives email sent to any address at your domain that doesn't exist
          </p>

          <div className="border border-border p-4">
            <p className="text-sm text-muted-foreground mb-2">Set a catch-all address for your domain:</p>
            <div className="flex gap-2">
              <input
                type="text"
                placeholder="domain.com"
                className="border border-border px-3 py-2 text-sm w-64"
                value={deliverabilityDomain}
                onChange={(e) => setDeliverabilityDomain(e.target.value)}
              />
              <input
                type="text"
                placeholder="forward@example.com"
                className="border border-border px-3 py-2 text-sm flex-1"
                value={formData.forward_to}
                onChange={(e) => setFormData({ ...formData, forward_to: e.target.value })}
              />
              <button className="bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90">
                Set Catch-All
              </button>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'deliverability' && (
        <div>
          <p className="text-sm text-muted-foreground mb-4">
            Deliverability · Check if your domain can receive emails from major providers
          </p>

          <div className="flex gap-2 mb-6">
            <input
              type="text"
              placeholder="yourdomain.com"
              className="border border-border px-3 py-2 text-sm w-64"
              value={deliverabilityDomain}
              onChange={(e) => setDeliverabilityDomain(e.target.value)}
            />
            <button
              onClick={checkDeliverability}
              className="bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90"
            >
              Check
            </button>
          </div>

          {deliverabilityResult && (
            <div className="grid grid-cols-2 gap-4">
              <DeliverabilityCard title="SPF" result={deliverabilityResult.spf} />
              <DeliverabilityCard title="DKIM" result={deliverabilityResult.dkim} />
              <DeliverabilityCard title="DMARC" result={deliverabilityResult.dmarc} />
              <DeliverabilityCard title="PTR" result={deliverabilityResult.ptr} />
            </div>
          )}
        </div>
      )}

      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card border border-border p-6 w-96">
            <h3 className="font-medium mb-4">
              Create {createType === 'mailbox' ? 'Mailbox' : createType === 'alias' ? 'Alias' : 'Forwarder'}
            </h3>

            {createType === 'mailbox' && (
              <div className="space-y-3">
                <div>
                  <label className="block text-sm mb-1">Email Address</label>
                  <input
                    type="email"
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                    className="w-full border border-border px-3 py-2 text-sm"
                    placeholder="user@domain.com"
                  />
                </div>
                <div>
                  <label className="block text-sm mb-1">Password</label>
                  <input
                    type="password"
                    value={formData.password}
                    onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                    className="w-full border border-border px-3 py-2 text-sm"
                    placeholder="Min 8 characters"
                  />
                </div>
                <div>
                  <label className="block text-sm mb-1">Display Name</label>
                  <input
                    type="text"
                    value={formData.display_name}
                    onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
                    className="w-full border border-border px-3 py-2 text-sm"
                    placeholder="John Doe"
                  />
                </div>
                <button
                  onClick={createMailbox}
                  className="w-full bg-primary text-primary-foreground py-2 text-sm font-medium hover:opacity-90"
                >
                  Create Mailbox
                </button>
              </div>
            )}

            {(createType === 'alias' || createType === 'forwarder') && (
              <div className="space-y-3">
                <div>
                  <label className="block text-sm mb-1">Domain</label>
                  <input
                    type="text"
                    value={formData.domain}
                    onChange={(e) => setFormData({ ...formData, domain: e.target.value })}
                    className="w-full border border-border px-3 py-2 text-sm"
                    placeholder="domain.com"
                  />
                </div>
                <div>
                  <label className="block text-sm mb-1">Source (local part)</label>
                  <input
                    type="text"
                    value={formData.source}
                    onChange={(e) => setFormData({ ...formData, source: e.target.value })}
                    className="w-full border border-border px-3 py-2 text-sm"
                    placeholder="info"
                  />
                </div>
                <div>
                  <label className="block text-sm mb-1">Destination</label>
                  <input
                    type="text"
                    value={formData.destination}
                    onChange={(e) => setFormData({ ...formData, destination: e.target.value })}
                    className="w-full border border-border px-3 py-2 text-sm"
                    placeholder={createType === 'forwarder' ? 'external@example.com' : 'user@domain.com'}
                  />
                </div>
                <button
                  onClick={createType === 'alias' ? createAlias : createForwarder}
                  className="w-full bg-primary text-primary-foreground py-2 text-sm font-medium hover:opacity-90"
                >
                  Create {createType === 'alias' ? 'Alias' : 'Forwarder'}
                </button>
              </div>
            )}

            <button
              onClick={() => setShowCreateModal(false)}
              className="mt-3 w-full border border-border py-2 text-sm hover:bg-muted"
            >
              Cancel
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

function DeliverabilityCard({ title, result }: { title: string; result: any }) {
  const StatusIcon = result.status ? CheckCircle : XCircle
  const statusColor = result.status ? 'text-green-600' : 'text-red-600'

  return (
    <div className="border border-border p-4">
      <div className="flex items-center gap-2 mb-2">
        <StatusIcon size={18} className={statusColor} />
        <span className="font-medium">{title}</span>
        <span className="text-sm text-muted-foreground">·</span>
        <span className="text-sm text-muted-foreground">{result.message}</span>
      </div>
      {result.record && (
        <p className="text-xs font-mono bg-muted p-2 mb-2">{result.record}</p>
      )}
      {result.recommendation && (
        <p className="text-xs text-muted-foreground">{result.recommendation}</p>
      )}
    </div>
  )
}