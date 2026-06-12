import { useState } from 'react'
import { Settings, User, Shield, Bell, Database, Globe, RefreshCw, Save, Download, Upload } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card'
import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Input, Label, Select, FormGroup, Switch } from '../../components/ui/Input'
import { Tabs, TabPanel } from '../../components/ui/Tabs'
import { PageHeader } from '../../components/ui/Misc'
import api from '../../lib/api'
import { cn } from '../../lib/utils'

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState('general')

  const tabs = [
    { id: 'general', label: 'General' },
    { id: 'security', label: 'Security' },
    { id: 'email', label: 'Email' },
    { id: 'updates', label: 'Updates' },
    { id: 'users', label: 'Users' },
    { id: 'backup', label: 'Backup' },
    { id: 'notifications', label: 'Notifications' },
  ]

  return (
    <div className="space-y-6">
      <PageHeader
        title="Settings"
        description="Panel configuration and preferences"
        breadcrumbs={[{ label: 'Settings' }]}
      />

      <Tabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />

      {activeTab === 'general' && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Server Identity</CardTitle>
              <CardDescription>Basic server information</CardDescription>
            </CardHeader>
            <div className="space-y-4">
              <FormGroup>
                <Label>Server Name</Label>
                <Input defaultValue="juvia-server" />
              </FormGroup>
              <FormGroup>
                <Label>Hostname</Label>
                <Input defaultValue="panel.example.com" />
              </FormGroup>
              <FormGroup>
                <Label>Timezone</Label>
                <Select defaultValue="UTC">
                  <option value="UTC">UTC</option>
                  <option value="America/New_York">America/New_York</option>
                  <option value="Europe/London">Europe/London</option>
                  <option value="Asia/Dhaka">Asia/Dhaka</option>
                </Select>
              </FormGroup>
              <Button>
                <Save className="w-4 h-4 mr-2" />
                Save Changes
              </Button>
            </div>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>SSH Settings</CardTitle>
              <CardDescription>Configure SSH access</CardDescription>
            </CardHeader>
            <div className="space-y-4">
              <FormGroup>
                <Label>SSH Port</Label>
                <Input defaultValue="22" type="number" className="w-32" />
              </FormGroup>
              <FormGroup>
                <Label>Root Login</Label>
                <Switch />
              </FormGroup>
              <Button>Save SSH Settings</Button>
            </div>
          </Card>
        </div>
      )}

      {activeTab === 'security' && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Two-Factor Authentication</CardTitle>
              <CardDescription>Add an extra layer of security to your account</CardDescription>
            </CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">TOTP Authenticator</p>
                <p className="text-xs text-text-secondary mt-0.5">Use an authenticator app for 2FA</p>
              </div>
              <Button variant="outline">Enable 2FA</Button>
            </div>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Session Management</CardTitle>
              <CardDescription>Manage active sessions</CardDescription>
            </CardHeader>
            <div className="space-y-3">
              <div className="flex items-center justify-between p-3 bg-accent/30 rounded">
                <div>
                  <p className="font-medium text-sm">Current Session</p>
                  <p className="text-xs text-text-secondary">Last active: just now</p>
                </div>
                <Badge variant="success">Active</Badge>
              </div>
            </div>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Login History</CardTitle>
              <CardDescription>Recent login attempts</CardDescription>
            </CardHeader>
            <div className="text-sm text-text-secondary text-center py-4">
              No recent login activity
            </div>
          </Card>
        </div>
      )}

      {activeTab === 'email' && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>SMTP Configuration</CardTitle>
              <CardDescription>Email sending settings</CardDescription>
            </CardHeader>
            <div className="space-y-4">
              <FormGroup>
                <Label>SMTP Host</Label>
                <Input defaultValue="smtp.example.com" />
              </FormGroup>
              <FormGroup>
                <Label>SMTP Port</Label>
                <Input defaultValue="587" type="number" className="w-32" />
              </FormGroup>
              <FormGroup>
                <Label>SMTP Username</Label>
                <Input defaultValue="noreply@example.com" />
              </FormGroup>
              <FormGroup>
                <Label>SMTP Password</Label>
                <Input type="password" defaultValue="••••••••" />
              </FormGroup>
              <Button>
                <Save className="w-4 h-4 mr-2" />
                Save SMTP Settings
              </Button>
            </div>
          </Card>
        </div>
      )}

      {activeTab === 'updates' && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Software Updates</CardTitle>
              <CardDescription>Keep your panel up to date</CardDescription>
            </CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Current Version</p>
                <p className="text-xs text-text-secondary mt-0.5">v0.1.0 — You are up to date</p>
              </div>
              <Button variant="outline">
                <RefreshCw className="w-4 h-4 mr-2" />
                Check for Updates
              </Button>
            </div>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Auto-Update</CardTitle>
              <CardDescription>Automatically download and install updates</CardDescription>
            </CardHeader>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm">Enable auto-updates</span>
                <Switch />
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm">Update channel</span>
                <Select defaultValue="stable" className="w-32">
                  <option value="stable">Stable</option>
                  <option value="beta">Beta</option>
                </Select>
              </div>
            </div>
          </Card>
        </div>
      )}

      {activeTab === 'users' && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>User Management</CardTitle>
              <CardDescription>Manage panel users</CardDescription>
            </CardHeader>
            <div className="space-y-3">
              <div className="flex items-center justify-between p-3 bg-accent/30 rounded">
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center">
                    <span className="text-xs font-semibold text-primary">A</span>
                  </div>
                  <div>
                    <p className="font-medium">admin</p>
                    <p className="text-xs text-text-secondary">admin@example.com · Admin</p>
                  </div>
                </div>
                <Badge variant="success">Active</Badge>
              </div>
            </div>
            <Button variant="outline" className="mt-4">
              <User className="w-4 h-4 mr-2" />
              Add User
            </Button>
          </Card>
        </div>
      )}

      {activeTab === 'backup' && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Backup Configuration</CardTitle>
              <CardDescription>Configure backup storage</CardDescription>
            </CardHeader>
            <div className="space-y-4">
              <FormGroup>
                <Label>Backup Storage</Label>
                <Select defaultValue="local">
                  <option value="local">Local Disk</option>
                  <option value="s3">Amazon S3</option>
                  <option value="r2">Cloudflare R2</option>
                  <option value="b2">Backblaze B2</option>
                </Select>
              </FormGroup>
              <FormGroup>
                <Label>Retention Period (days)</Label>
                <Input defaultValue="30" type="number" className="w-32" />
              </FormGroup>
              <Button>
                <Save className="w-4 h-4 mr-2" />
                Save Backup Settings
              </Button>
            </div>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Export / Import</CardTitle>
              <CardDescription>Export or import panel configuration</CardDescription>
            </CardHeader>
            <div className="flex gap-3">
              <Button variant="outline">
                <Download className="w-4 h-4 mr-2" />
                Export Config
              </Button>
              <Button variant="outline">
                <Upload className="w-4 h-4 mr-2" />
                Import Config
              </Button>
            </div>
          </Card>
        </div>
      )}

      {activeTab === 'notifications' && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Notification Preferences</CardTitle>
              <CardDescription>Configure alert notifications</CardDescription>
            </CardHeader>
            <div className="space-y-4">
              {[
                { label: 'Email on alert', desc: 'Send email when an alert is triggered' },
                { label: 'SSL expiry warnings', desc: 'Alert 14 days before SSL expires' },
                { label: 'Backup failures', desc: 'Alert when a backup fails' },
                { label: 'Website down alerts', desc: 'Alert when a website goes offline' },
              ].map((item) => (
                <div key={item.label} className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium">{item.label}</p>
                    <p className="text-xs text-text-secondary">{item.desc}</p>
                  </div>
                  <Switch defaultChecked />
                </div>
              ))}
              <Button>Save Notification Settings</Button>
            </div>
          </Card>
        </div>
      )}
    </div>
  )
}