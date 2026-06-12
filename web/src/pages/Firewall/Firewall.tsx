import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Shield, Plus, RefreshCw, Trash2, CheckCircle, XCircle, Globe, Lock, Server } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge, StatusBadge } from '../../components/ui/Badge'
import { Table } from '../../components/ui/Table'
import { PageHeader } from '../../components/ui/Misc'
import { Modal } from '../../components/ui/Modal'
import { Input, Label, Select, FormGroup } from '../../components/ui/Input'
import api from '../../lib/api'
import { formatDate } from '../../lib/utils'
import { cn } from '../../lib/utils'

interface FirewallRule {
  id: number
  name: string
  action: 'allow' | 'deny'
  protocol: 'tcp' | 'udp' | 'all'
  port: string
  source: string
  enabled: boolean
  created_at: string
}

export default function Firewall() {
  const [rules, setRules] = useState<FirewallRule[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [formData, setFormData] = useState({
    name: '', action: 'allow', protocol: 'tcp', port: '', source: '0.0.0.0/0',
  })
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    loadRules()
  }, [])

  const loadRules = async () => {
    try {
      const res = await api.get('/firewall/rules')
      setRules(res.data.data || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = async () => {
    if (!formData.name || !formData.port) return
    setSaving(true)
    try {
      await api.post('/firewall/rules', formData)
      setShowModal(false)
      setFormData({ name: '', action: 'allow', protocol: 'tcp', port: '', source: '0.0.0.0/0' })
      loadRules()
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to create rule')
    } finally {
      setSaving(false)
    }
  }

  const handleToggle = async (id: number, enabled: boolean) => {
    try {
      await api.put(`/firewall/rules/${id}`, { enabled: !enabled })
      loadRules()
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to update rule')
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this rule?')) return
    try {
      await api.delete(`/firewall/rules/${id}`)
      setRules(rules.filter((r) => r.id !== id))
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to delete')
    }
  }

  const columns = [
    {
      key: 'name',
      header: 'Rule',
      render: (rule: FirewallRule) => (
        <div className="flex items-center gap-3">
          <div className={cn('w-8 h-8 rounded flex items-center justify-center',
            rule.action === 'allow' ? 'bg-success/10' : 'bg-danger/10')}>
            {rule.action === 'allow' ? (
              <CheckCircle className="w-4 h-4 text-success" />
            ) : (
              <XCircle className="w-4 h-4 text-danger" />
            )}
          </div>
          <div>
            <p className="font-medium text-sm">{rule.name}</p>
            <p className="text-xs text-text-secondary capitalize">
              {rule.action} {rule.protocol.toUpperCase()} port {rule.port}
            </p>
          </div>
        </div>
      ),
    },
    {
      key: 'source',
      header: 'Source',
      render: (rule: FirewallRule) => (
        <span className="text-xs font-mono text-text-secondary">{rule.source}</span>
      ),
    },
    {
      key: 'enabled',
      header: 'Status',
      render: (rule: FirewallRule) => (
        <button
          onClick={() => handleToggle(rule.id, rule.enabled)}
          className={cn(
            'relative w-10 h-5 rounded-full transition-colors',
            rule.enabled ? 'bg-success' : 'bg-border'
          )}
        >
          <span className={cn(
            'absolute top-0.5 w-4 h-4 bg-white rounded-full transition-transform',
            rule.enabled ? 'left-5.5' : 'left-0.5'
          )} />
        </button>
      ),
    },
    {
      key: 'created',
      header: 'Created',
      render: (rule: FirewallRule) => (
        <span className="text-xs text-text-secondary">{formatDate(rule.created_at)}</span>
      ),
    },
    {
      key: 'actions',
      header: '',
      width: '80px',
      render: (rule: FirewallRule) => (
        <button
          onClick={() => handleDelete(rule.id)}
          className="p-1.5 text-text-secondary hover:text-danger hover:bg-danger/10 rounded transition-colors"
        >
          <Trash2 className="w-3.5 h-3.5" />
        </button>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <PageHeader
        title="Firewall"
        description="Control incoming and outgoing network traffic"
        breadcrumbs={[{ label: 'Firewall' }]}
        actions={
          <Button onClick={() => setShowModal(true)}>
            <Plus className="w-4 h-4 mr-2" />
            Add Rule
          </Button>
        }
      />

      <Card padding="none">
        <div className="p-4 border-b border-border">
          <div className="flex items-center gap-4">
            <div className="p-2 bg-primary/10 rounded">
              <Shield className="w-5 h-5 text-primary" />
            </div>
            <div>
              <p className="font-medium">{rules.filter(r => r.enabled).length} active rules</p>
              <p className="text-xs text-text-secondary">Protecting your server</p>
            </div>
          </div>
        </div>
        <Table
          columns={columns}
          data={rules}
          keyField="id"
          loading={loading}
          emptyMessage="No firewall rules configured"
        />
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card className="text-center">
          <Globe className="w-8 h-8 mx-auto text-success mb-2" />
          <p className="font-medium">HTTP/HTTPS</p>
          <p className="text-xs text-text-secondary mt-1">Ports 80, 443</p>
        </Card>
        <Card className="text-center">
          <Server className="w-8 h-8 mx-auto text-primary mb-2" />
          <p className="font-medium">SSH</p>
          <p className="text-xs text-text-secondary mt-1">Port 22 (rate limited)</p>
        </Card>
        <Card className="text-center">
          <Lock className="w-8 h-8 mx-auto text-warning mb-2" />
          <p className="font-medium">Panel</p>
          <p className="text-xs text-text-secondary mt-1">Port 8080</p>
        </Card>
      </div>

      <Modal
        open={showModal}
        onClose={() => setShowModal(false)}
        title="Add Firewall Rule"
        size="sm"
      >
        <div className="space-y-4">
          <FormGroup>
            <Label>Rule Name</Label>
            <Input
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="SSH Access"
            />
          </FormGroup>
          <FormGroup>
            <Label>Action</Label>
            <Select
              value={formData.action}
              onChange={(e) => setFormData({ ...formData, action: e.target.value })}
            >
              <option value="allow">Allow</option>
              <option value="deny">Deny</option>
            </Select>
          </FormGroup>
          <FormGroup>
            <Label>Protocol</Label>
            <Select
              value={formData.protocol}
              onChange={(e) => setFormData({ ...formData, protocol: e.target.value })}
            >
              <option value="tcp">TCP</option>
              <option value="udp">UDP</option>
              <option value="all">All</option>
            </Select>
          </FormGroup>
          <FormGroup>
            <Label>Port</Label>
            <Input
              value={formData.port}
              onChange={(e) => setFormData({ ...formData, port: e.target.value })}
              placeholder="22 or 8000-9000"
            />
          </FormGroup>
          <FormGroup>
            <Label>Source IP / CIDR</Label>
            <Input
              value={formData.source}
              onChange={(e) => setFormData({ ...formData, source: e.target.value })}
              placeholder="0.0.0.0/0 or 192.168.1.0/24"
            />
          </FormGroup>
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="outline" onClick={() => setShowModal(false)}>Cancel</Button>
            <Button onClick={handleCreate} loading={saving}>Create Rule</Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}