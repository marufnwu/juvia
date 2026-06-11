import { useEffect, useState } from 'react'
import { Shield, Plus, X } from 'lucide-react'
import api from '../../lib/api'

interface FirewallRule {
  id: number
  name: string
  description: string
  action: string
  port: string
  protocol: string
  source: string
  created_at: string
}

export default function Firewall() {
  const [rules, setRules] = useState<FirewallRule[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    action: 'allow',
    port: '',
    protocol: 'tcp',
    source: 'any',
  })

  useEffect(() => {
    loadRules()
  }, [])

  const loadRules = async () => {
    setLoading(true)
    try {
      const res = await api.get('/firewall/rules')
      setRules(res.data.data || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const createRule = async () => {
    try {
      await api.post('/firewall/rules', formData)
      setShowModal(false)
      setFormData({ name: '', description: '', action: 'allow', port: '', protocol: 'tcp', source: 'any' })
      loadRules()
    } catch (err) {
      console.error(err)
    }
  }

  const deleteRule = async (id: number) => {
    if (!confirm('Delete this firewall rule?')) return
    try {
      await api.delete(`/firewall/rules/${id}`)
      loadRules()
    } catch (err) {
      console.error(err)
    }
  }

  const getPortDescription = (port: string, _protocol: string): string => {
    const portMap: Record<string, string> = {
      '22': 'SSH · Secure shell for server access',
      '80': 'HTTP · Standard web traffic',
      '443': 'HTTPS · Encrypted web traffic',
      '8080': 'HTTP Alt · Alternative web port',
      '25': 'SMTP · Email sending',
      '587': 'SMTP Secure · Email submission',
      '993': 'IMAPS · Secure email retrieval',
      '995': 'POP3S · Secure email retrieval',
      '3306': 'MySQL · Database access',
      '5432': 'PostgreSQL · Database access',
    }
    const desc = portMap[port] || ''
    return desc
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Firewall</h1>
          <p className="text-sm text-muted-foreground">
            Firewall Rule · Controls which network traffic can reach your server
          </p>
        </div>
        <button
          onClick={() => setShowModal(true)}
          className="flex items-center gap-2 bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90"
        >
          <Plus size={16} />
          Create Rule
        </button>
      </div>

      {loading ? (
        <div className="text-muted-foreground">Loading rules...</div>
      ) : rules.length === 0 ? (
        <div className="text-center py-12 border border-border">
          <Shield size={48} className="mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">No firewall rules configured</p>
        </div>
      ) : (
        <div className="border border-border">
          <table className="w-full text-sm">
            <thead className="bg-muted/50 border-b border-border">
              <tr>
                <th className="text-left p-3 font-medium">Name</th>
                <th className="text-left p-3 font-medium">Action</th>
                <th className="text-left p-3 font-medium">Port</th>
                <th className="text-left p-3 font-medium">Source</th>
                <th className="text-left p-3 font-medium">Description</th>
                <th className="text-left p-3 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {rules.map((rule) => (
                <tr key={rule.id} className="border-b border-border hover:bg-muted/30">
                  <td className="p-3 font-medium">{rule.name}</td>
                  <td className="p-3">
                    <span className={`text-xs px-2 py-1 rounded ${rule.action === 'allow' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
                      {rule.action}
                    </span>
                  </td>
                  <td className="p-3">
                    {rule.port}/{rule.protocol}
                    {getPortDescription(rule.port, rule.protocol) && (
                      <span className="text-xs text-muted-foreground ml-2">
                        · {getPortDescription(rule.port, rule.protocol)}
                      </span>
                    )}
                  </td>
                  <td className="p-3 text-muted-foreground">{rule.source === 'any' ? 'Any IP' : rule.source}</td>
                  <td className="p-3 text-muted-foreground text-xs">{rule.description || '—'}</td>
                  <td className="p-3">
                    <button onClick={() => deleteRule(rule.id)} className="text-red-600 hover:underline text-xs">
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card border border-border p-6 w-96">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-medium">Create Firewall Rule</h3>
              <button onClick={() => setShowModal(false)} className="text-muted-foreground hover:text-foreground">
                <X size={20} />
              </button>
            </div>

            <div className="space-y-3">
              <div>
                <label className="block text-sm mb-1">Rule Name</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full border border-border px-3 py-2 text-sm"
                  placeholder="SSH Access"
                />
              </div>
              <div>
                <label className="block text-sm mb-1">Action</label>
                <select
                  value={formData.action}
                  onChange={(e) => setFormData({ ...formData, action: e.target.value })}
                  className="w-full border border-border px-3 py-2 text-sm"
                >
                  <option value="allow">Allow</option>
                  <option value="deny">Deny</option>
                </select>
              </div>
              <div className="flex gap-2">
                <div className="flex-1">
                  <label className="block text-sm mb-1">Port</label>
                  <input
                    type="text"
                    value={formData.port}
                    onChange={(e) => setFormData({ ...formData, port: e.target.value })}
                    className="w-full border border-border px-3 py-2 text-sm"
                    placeholder="22"
                  />
                </div>
                <div className="w-24">
                  <label className="block text-sm mb-1">Protocol</label>
                  <select
                    value={formData.protocol}
                    onChange={(e) => setFormData({ ...formData, protocol: e.target.value })}
                    className="w-full border border-border px-3 py-2 text-sm"
                  >
                    <option value="tcp">TCP</option>
                    <option value="udp">UDP</option>
                    <option value="both">Both</option>
                  </select>
                </div>
              </div>
              <div>
                <label className="block text-sm mb-1">Source IP</label>
                <input
                  type="text"
                  value={formData.source}
                  onChange={(e) => setFormData({ ...formData, source: e.target.value })}
                  className="w-full border border-border px-3 py-2 text-sm"
                  placeholder="any or 192.168.1.0/24"
                />
                <p className="text-xs text-muted-foreground mt-1">Use 'any' for any IP, or CIDR notation like '192.168.1.0/24'</p>
              </div>
              <div>
                <label className="block text-sm mb-1">Description (optional)</label>
                <input
                  type="text"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  className="w-full border border-border px-3 py-2 text-sm"
                  placeholder="Allow SSH from office network"
                />
              </div>

              <button
                onClick={createRule}
                className="w-full bg-primary text-primary-foreground py-2 text-sm font-medium hover:opacity-90"
              >
                Create Rule
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}