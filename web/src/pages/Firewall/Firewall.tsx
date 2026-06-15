import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Shield, Plus, RefreshCw, Trash2, CheckCircle, XCircle, Globe, Lock, Server, Pencil } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge, StatusBadge } from '../../components/ui/Badge'
import { Table } from '../../components/ui/Table'
import { PageHeader } from '../../components/ui/Misc'
import { Modal, ConfirmModal } from '../../components/ui/Modal'
import { Input, Label, Select, FormGroup, Switch } from '../../components/ui/Input'
import api from '../../lib/api'
import { formatDate } from '../../lib/utils'
import { cn } from '../../lib/utils'
import { useApiError } from '../../hooks/useToast'

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
  const showError = useApiError()
  const [rules, setRules] = useState<FirewallRule[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [formData, setFormData] = useState({
    name: '', action: 'allow', protocol: 'tcp', port: '', source: '0.0.0.0/0',
  })
  const [saving, setSaving] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [editingRule, setEditingRule] = useState<FirewallRule | null>(null)
  const [editFormData, setEditFormData] = useState({ name: '', action: 'allow' as 'allow' | 'deny', protocol: 'tcp' as 'tcp' | 'udp' | 'all', port: '', source: '' })
  const [confirmDelete, setConfirmDelete] = useState<number | null>(null)
  const [deleting, setDeleting] = useState(false)

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
      showError(err, 'Failed to create rule')
    } finally {
      setSaving(false)
    }
  }

  const handleToggle = async (id: number, enabled: boolean) => {
    try {
      await api.put(`/firewall/rules/${id}`, { enabled: !enabled })
      loadRules()
    } catch (err: any) {
      showError(err, 'Failed to update rule')
    }
  }

  const handleDelete = (id: number) => {
    setConfirmDelete(id)
  }

  const doDelete = async () => {
    if (!confirmDelete) return
    const id = confirmDelete
    setConfirmDelete(null)
    setDeleting(true)
    try {
      await api.delete(`/firewall/rules/${id}`)
      setRules(rules.filter((r) => r.id !== id))
    } catch (err: any) {
      showError(err, 'Failed to delete')
    } finally {
      setDeleting(false)
    }
  }

  const openEditModal = (rule: FirewallRule) => {
    setEditingRule(rule)
    setEditFormData({
      name: rule.name,
      action: rule.action,
      protocol: rule.protocol,
      port: rule.port,
      source: rule.source,
    })
    setShowEditModal(true)
  }

  const handleEdit = async () => {
    if (!editingRule || !editFormData.name || !editFormData.port) return
    setSaving(true)
    try {
      await api.put(`/firewall/rules/${editingRule.id}`, editFormData)
      setShowEditModal(false)
      setEditingRule(null)
      loadRules()
    } catch (err: any) {
      showError(err, 'Failed to update rule')
    } finally {
      setSaving(false)
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
        <Switch
          checked={rule.enabled}
          onChange={() => handleToggle(rule.id, rule.enabled)}
        />
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
      width: '120px',
      render: (rule: FirewallRule) => (
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => openEditModal(rule)}
            title="Edit"
          >
            <Pencil className="w-3.5 h-3.5" />
          </Button>
          <Button
            variant="danger"
            size="icon"
            onClick={() => handleDelete(rule.id)}
            title="Delete"
          >
            <Trash2 className="w-3.5 h-3.5" />
          </Button>
        </div>
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
          emptyMessage={
            <div className="p-8 text-center">
              <Shield className="w-12 h-12 mx-auto text-text-secondary/50 mb-3" />
              <p className="text-sm text-text-secondary mb-3">No custom firewall rules</p>
              <Button variant="primary" size="sm" onClick={() => setShowModal(true)}>Add your first rule</Button>
            </div>
          }
        />
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card className="text-center">
          <Globe className="w-8 h-8 mx-auto text-success mb-2" />
          <p className="font-medium">HTTP/HTTPS · Web traffic ports (80, 443)</p>
        </Card>
        <Card className="text-center">
          <Server className="w-8 h-8 mx-auto text-primary mb-2" />
          <p className="font-medium">SSH · Secure remote login (port 22)</p>
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
              <option value="tcp">TCP · Reliable (web, email)</option>
              <option value="udp">UDP · Fast (DNS, streaming)</option>
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

      <Modal open={showEditModal} onClose={() => setShowEditModal(false)} title="Edit Firewall Rule" size="sm">
        <div className="space-y-4">
          {editingRule && (
            <>
              <FormGroup>
                <Label>Rule Name</Label>
                <Input
                  value={editFormData.name}
                  onChange={(e) => setEditFormData({ ...editFormData, name: e.target.value })}
                  placeholder="SSH Access"
                />
              </FormGroup>
              <FormGroup>
                <Label>Action</Label>
                <Select
                  value={editFormData.action}
                  onChange={(e) => setEditFormData({ ...editFormData, action: e.target.value as 'allow' | 'deny' })}
                >
                  <option value="allow">Allow</option>
                  <option value="deny">Deny</option>
                </Select>
              </FormGroup>
              <FormGroup>
            <Label>Protocol · Type of traffic</Label>
                <Select
                  value={editFormData.protocol}
                  onChange={(e) => setEditFormData({ ...editFormData, protocol: e.target.value as 'tcp' | 'udp' | 'all' })}
                >
                  <option value="tcp">TCP</option>
                  <option value="udp">UDP</option>
                  <option value="all">All</option>
                </Select>
              </FormGroup>
              <FormGroup>
            <Label>Port · Traffic channel number</Label>
                <Input
                  value={editFormData.port}
                  onChange={(e) => setEditFormData({ ...editFormData, port: e.target.value })}
                  placeholder="22 or 8000-9000"
                />
              </FormGroup>
              <FormGroup>
            <Label>Source IP · Which addresses this rule applies to</Label>
                <Input
                  value={editFormData.source}
                  onChange={(e) => setEditFormData({ ...editFormData, source: e.target.value })}
                  placeholder="0.0.0.0/0 or 192.168.1.0/24"
                />
              </FormGroup>
              <div className="flex justify-end gap-3 pt-2">
                <Button variant="outline" onClick={() => setShowEditModal(false)}>Cancel</Button>
                <Button onClick={handleEdit} loading={saving}>Save Changes</Button>
              </div>
            </>
          )}
        </div>
      </Modal>

      <ConfirmModal
        open={!!confirmDelete}
        onClose={() => setConfirmDelete(null)}
        onConfirm={doDelete}
        title="Delete Firewall Rule"
        description="Remove this firewall rule? The change takes effect immediately."
        confirmLabel="Delete"
        variant="danger"
        loading={deleting}
      />
    </div>
  )
}
