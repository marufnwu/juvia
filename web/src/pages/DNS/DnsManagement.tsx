import { useEffect, useState, useCallback } from 'react'
import { Link } from 'react-router-dom'
import {
  Globe, Server, RefreshCw, Loader2, AlertTriangle, Copy, Check, ExternalLink, Database, Activity, Plus, Trash2
} from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { Modal } from '../../components/ui/Modal'
import { ConfirmModal } from '../../components/ui/Modal'
import { PageHeader } from '../../components/ui/Misc'
import { TableHeader, TableBody, TableRow, TableHead, TableCell } from '../../components/ui/Table'
import api from '../../lib/api'
import { formatDate } from '../../lib/utils'
import { useApiError } from '../../hooks/useToast'

interface Zone {
  id: number
  website_id: number | null
  domain: string
  serial: number
  record_count: number
  created_at: string
  updated_at: string
}

interface Nameservers {
  server_ip: string
  ns_brand_domain: string
  ns1_hostname: string
  ns2_hostname: string
}

export default function DnsManagement() {
  const showError = useApiError()
  const [zones, setZones] = useState<Zone[]>([])
  const [nameservers, setNameservers] = useState<Nameservers | null>(null)
  const [loading, setLoading] = useState(true)
  const [restarting, setRestarting] = useState(false)
  const [copied, setCopied] = useState('')
  const [health, setHealth] = useState<any>(null)
  const [checkingHealth, setCheckingHealth] = useState(false)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [newZoneDomain, setNewZoneDomain] = useState('')
  const [creating, setCreating] = useState(false)
  const [confirmDeleteZone, setConfirmDeleteZone] = useState<Zone | null>(null)
  const [deleting, setDeleting] = useState(false)

  const checkHealth = useCallback(async () => {
    setCheckingHealth(true)
    try {
      const res = await api.get('/dns/health')
      setHealth(res.data.data)
    } catch (err: any) {
      showError(err, 'Failed to check DNS health')
    } finally {
      setCheckingHealth(false)
    }
  }, [showError])

  useEffect(() => {
    loadData()
  }, [])

  useEffect(() => {
    if (nsConfigured && !health) {
      checkHealth()
    }
  }, [nameservers])

  const loadData = async () => {
    setLoading(true)
    try {
      const [zonesRes, nsRes] = await Promise.allSettled([
        api.get('/dns/zones'),
        api.get('/dns/nameservers'),
      ])
      if (zonesRes.status === 'fulfilled') {
        setZones(zonesRes.value.data.data || [])
      }
      if (nsRes.status === 'fulfilled') {
        setNameservers(nsRes.value.data.data)
      }
    } catch (err: any) {
      showError(err, 'Failed to load DNS data')
    } finally {
      setLoading(false)
    }
  }

  const handleRestart = async () => {
    setRestarting(true)
    try {
      await api.post('/dns/restart')
      loadData()
    } catch (err: any) {
      showError(err, 'Failed to restart BIND')
    } finally {
      setRestarting(false)
    }
  }

  const handleCreateZone = async () => {
    if (!newZoneDomain.trim()) return
    setCreating(true)
    try {
      await api.post('/dns/zones', { domain: newZoneDomain.trim() })
      setNewZoneDomain('')
      setShowCreateModal(false)
      loadData()
    } catch (err: any) {
      showError(err, 'Failed to create zone')
    } finally {
      setCreating(false)
    }
  }

  const handleDeleteZone = async () => {
    if (!confirmDeleteZone) return
    setDeleting(true)
    try {
      await api.delete(`/dns/zones/${confirmDeleteZone.id}`)
      setConfirmDeleteZone(null)
      loadData()
    } catch (err: any) {
      showError(err, 'Failed to delete zone')
    } finally {
      setDeleting(false)
    }
  }

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    setCopied(label)
    setTimeout(() => setCopied(''), 2000)
  }

  const nsConfigured = nameservers?.ns1_hostname && nameservers?.ns2_hostname

  return (
    <div className="space-y-6">
      <PageHeader
        title="DNS & Nameservers"
        description="Manage your authoritative DNS zones and nameserver settings"
        actions={
          <Button variant="outline" onClick={handleRestart} disabled={restarting}>
            {restarting ? (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            ) : (
              <RefreshCw className="w-4 h-4 mr-2" />
            )}
            Restart BIND
          </Button>
        }
      />

      {nameservers && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>Server IP · Your server's address</CardTitle>
            </CardHeader>
            <div className="p-4">
              <div className="flex items-center gap-2">
                <code className="text-lg font-mono">{nameservers.server_ip || 'Not set'}</code>
                {nameservers.server_ip && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => copyToClipboard(nameservers.server_ip, 'ip')}
                  >
                    {copied === 'ip' ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                  </Button>
                )}
              </div>
              <p className="text-sm text-muted-foreground mt-2">
                Point your domain's A record to this IP address
              </p>
            </div>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Nameservers · Your DNS server hostnames</CardTitle>
            </CardHeader>
            <div className="p-4 space-y-3">
              <div className="flex items-center justify-between">
                <div>
                  <span className="text-sm text-muted-foreground">NS1</span>
                  {nameservers.ns1_hostname ? (
                    <p className="font-mono text-sm">{nameservers.ns1_hostname}</p>
                  ) : (
                    <p className="text-sm text-warning">Not configured — set in Settings → Nameservers</p>
                  )}
                </div>
                {nameservers.ns1_hostname && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => copyToClipboard(nameservers.ns1_hostname, 'ns1')}
                  >
                    {copied === 'ns1' ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                  </Button>
                )}
              </div>
              <div className="flex items-center justify-between">
                <div>
                  <span className="text-sm text-muted-foreground">NS2</span>
                  {nameservers.ns2_hostname ? (
                    <p className="font-mono text-sm">{nameservers.ns2_hostname}</p>
                  ) : (
                    <p className="text-sm text-warning">Not configured — set in Settings → Nameservers</p>
                  )}
                </div>
                {nameservers.ns2_hostname && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => copyToClipboard(nameservers.ns2_hostname, 'ns2')}
                  >
                    {copied === 'ns2' ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                  </Button>
                )}
              </div>
              {nameservers.ns_brand_domain && (
                <p className="text-xs text-muted-foreground">
                  Brand domain: {nameservers.ns_brand_domain}
                </p>
              )}
              {!nsConfigured && (
                <Link to="/settings" className="inline-flex items-center gap-1 text-xs text-primary hover:underline mt-1">
                  Configure nameservers →
                </Link>
              )}
            </div>
          </Card>
        </div>
      )}

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>DNS Health · Check if your DNS is working</CardTitle>
              <CardDescription className="mt-1">
                Verify BIND is running and nameservers resolve to this server
              </CardDescription>
            </div>
            <Button variant="outline" size="sm" onClick={checkHealth} disabled={checkingHealth}>
              {checkingHealth ? (
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <Activity className="w-4 h-4 mr-2" />
              )}
              Check Health
            </Button>
          </div>
        </CardHeader>
        {health ? (
          <div className="p-4 space-y-3">
            <div className="flex items-center gap-2">
              {health.bind_active ? (
                <Badge variant="success">BIND Active</Badge>
              ) : (
                <Badge variant="danger">BIND Inactive</Badge>
              )}
              {!nsConfigured ? (
                <Badge variant="warning">Nameservers Not Configured</Badge>
              ) : health.all_ok ? (
                <Badge variant="success">All OK</Badge>
              ) : (
                <Badge variant="warning">Issues Found</Badge>
              )}
            </div>
            {!nsConfigured && (
              <div className="bg-warning/10 border border-warning/20 rounded-lg p-3 text-sm">
                <p className="font-medium text-warning">Nameserver settings are not configured.</p>
                <p className="text-muted-foreground mt-1">
                  Go to <strong>Settings → General → Nameserver Settings</strong> to set your brand domain and NS hostnames.
                </p>
              </div>
            )}
            {health.nameservers?.map((ns: any) => (
              <div key={ns.label} className="flex items-center gap-3">
                {ns.ok ? (
                  <Check className="w-4 h-4 text-green-500" />
                ) : (
                  <AlertTriangle className="w-4 h-4 text-yellow-500" />
                )}
                <span className="font-mono text-sm">{ns.hostname}</span>
                <span className="text-muted-foreground text-sm">
                  {ns.a_record ? `→ ${ns.a_record}` : 'Not resolving'}
                </span>
                {ns.resolves_to_server ? (
                  <Badge variant="success">Points here</Badge>
                ) : ns.a_record ? (
                  <Badge variant="danger">Wrong IP</Badge>
                ) : null}
              </div>
            ))}
            {health.nameservers?.some((ns: any) => !ns.ok) && (
              <div className="bg-warning/10 border border-warning/20 rounded-lg p-3 text-sm mt-2">
                <p className="font-medium text-warning">Glue records missing</p>
                <p className="text-muted-foreground mt-1">
                  Your nameserver hostnames (ns1/ns2) don't resolve yet. At your domain registrar:
                </p>
                <ol className="list-decimal list-inside text-muted-foreground mt-2 space-y-1">
                  <li>Add glue records: <code>{nameservers?.ns1_hostname || 'ns1.<brand>'}</code> → <code>{health.server_ip}</code></li>
                  <li>Add glue records: <code>{nameservers?.ns2_hostname || 'ns2.<brand>'}</code> → <code>{health.server_ip}</code></li>
                  <li>Set your domain's nameservers to ns1 and ns2 above</li>
                </ol>
                <p className="text-muted-foreground mt-2 text-xs">
                  Glue records tell the internet where your nameservers are. Without them, DNS lookups fail.
                </p>
              </div>
            )}
          </div>
        ) : (
          <div className="p-4 text-sm text-muted-foreground">
            Click "Check Health" to verify your nameservers are working
          </div>
        )}
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>DNS Zones · Domains managed by this server</CardTitle>
              <CardDescription className="mt-1">
                Each zone is a domain whose DNS is served by BIND on this server
              </CardDescription>
            </div>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => setShowCreateModal(true)}>
                <Plus className="w-4 h-4 mr-2" />
                Create Zone
              </Button>
              <Button variant="outline" size="sm" onClick={loadData}>
                <RefreshCw className="w-4 h-4 mr-2" />
                Refresh
              </Button>
            </div>
          </div>
        </CardHeader>

        {zones.length === 0 ? (
          <div className="p-8 text-center">
            <Globe className="w-10 h-10 mx-auto text-muted-foreground mb-3" />
            <p className="text-sm text-muted-foreground mb-1">No DNS zones yet</p>
            <p className="text-xs text-muted-foreground">
              Zones are created automatically when you add a website
            </p>
          </div>
        ) : (
          <TableHeader>
            <TableRow>
              <TableHead>Domain</TableHead>
              <TableHead>Records</TableHead>
              <TableHead>Serial</TableHead>
              <TableHead>Website</TableHead>
              <TableHead>Last Updated</TableHead>
              <TableHead className="w-[80px]"></TableHead>
            </TableRow>
          </TableHeader>
        )}

        {zones.length > 0 && (
          <TableBody>
            {zones.map((z) => (
              <TableRow key={z.id}>
                <TableCell>
                  <div className="flex items-center gap-2">
                    <Globe className="w-4 h-4 text-muted-foreground" />
                    <span className="font-medium">{z.domain}</span>
                  </div>
                </TableCell>
                <TableCell>
                  <Badge variant="info">{z.record_count} records</Badge>
                </TableCell>
                <TableCell className="text-muted-foreground font-mono text-sm">
                  {z.serial}
                </TableCell>
                <TableCell>
                  {z.website_id ? (
                    <Link
                      to={`/websites/${z.website_id}`}
                      className="text-primary hover:underline inline-flex items-center gap-1"
                    >
                      View
                      <ExternalLink className="w-3 h-3" />
                    </Link>
                  ) : (
                    <span className="text-muted-foreground">—</span>
                  )}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {formatDate(z.updated_at)}
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-1">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => copyToClipboard(z.domain + '. zone: ' + z.serial, 'zone-' + z.id)}
                    >
                      {copied === 'zone-' + z.id ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setConfirmDeleteZone(z)}
                      title="Delete zone"
                    >
                      <Trash2 className="w-4 h-4 text-danger" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        )}
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>How DNS Works · Quick reference</CardTitle>
        </CardHeader>
        <div className="p-4 space-y-4 text-sm text-muted-foreground">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <h4 className="font-medium text-foreground">To point a domain here:</h4>
              <ol className="list-decimal list-inside space-y-1">
                <li>Set your domain's nameservers to NS1 and NS2 above</li>
                <li>Or create an A record pointing to the server IP</li>
                <li>Wait for DNS propagation (up to 48 hours)</li>
              </ol>
            </div>
            <div className="space-y-2">
              <h4 className="font-medium text-foreground">Records explained:</h4>
              <ul className="space-y-1">
                <li><strong>A</strong> — Maps domain to an IP address</li>
                <li><strong>CNAME</strong> — Alias pointing to another domain</li>
                <li><strong>MX</strong> — Mail server for the domain</li>
                <li><strong>TXT</strong> — Text records (SPF, DKIM, DMARC)</li>
              </ul>
            </div>
          </div>
        </div>
      </Card>

      <Modal open={showCreateModal} onClose={() => { setShowCreateModal(false); setNewZoneDomain('') }} title="Create DNS Zone">
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Create a standalone DNS zone not linked to any website. You can manage its records after creation.
          </p>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">
              Domain · The domain name for this zone
            </label>
            <input
              type="text"
              value={newZoneDomain}
              onChange={(e) => setNewZoneDomain(e.target.value)}
              placeholder="example.com"
              className="w-full h-10 px-3 bg-background border border-border rounded text-sm"
              onKeyDown={(e) => e.key === 'Enter' && handleCreateZone()}
              autoFocus
            />
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="outline" onClick={() => { setShowCreateModal(false); setNewZoneDomain('') }}>
              Cancel
            </Button>
            <Button onClick={handleCreateZone} disabled={!newZoneDomain.trim() || creating}>
              {creating ? 'Creating...' : 'Create Zone'}
            </Button>
          </div>
        </div>
      </Modal>

      <ConfirmModal
        open={!!confirmDeleteZone}
        onClose={() => setConfirmDeleteZone(null)}
        onConfirm={handleDeleteZone}
        title="Delete DNS Zone"
        description={`Delete the zone for ${confirmDeleteZone?.domain}? This will remove all DNS records and cannot be undone.`}
        confirmLabel="Delete"
        variant="danger"
        loading={deleting}
      />
    </div>
  )
}
