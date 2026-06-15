import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, Database, Check } from 'lucide-react'
import { Card, CardHeader, CardTitle } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { PageHeader } from '../../components/ui/Misc'
import api from '../../lib/api'
import { Input, Switch } from '../../components/ui/Input'
import { cn } from '../../lib/utils'

export default function CreateDatabase() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [formData, setFormData] = useState({
    name: '',
    engine: 'mysql',
    create_user: true,
    username: '',
    password: '',
    grant_all: true,
  })

  const handleSubmit = async () => {
    if (!formData.name) {
      setError('Database name is required')
      return
    }

    setLoading(true)
    setError('')

    try {
      const res = await api.post('/databases', {
        name: formData.name,
        engine: formData.engine,
        create_user: formData.create_user,
        username: formData.create_user ? formData.username : undefined,
        password: formData.create_user ? formData.password : undefined,
        grant_all: formData.grant_all,
      })

      if (res.data.success) {
        navigate('/databases')
      }
    } catch (err: any) {
      setError(err.response?.data?.error?.user_message || 'Failed to create database')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-xl mx-auto space-y-6">
      <PageHeader
        title="Create Database"
        description="Add a new MySQL or PostgreSQL database"
        breadcrumbs={[
          { label: 'Databases', href: '/databases' },
          { label: 'Create' },
        ]}
        actions={
          <Link
            to="/databases"
            className="inline-flex items-center gap-2 px-4 py-2 border border-border text-sm rounded hover:bg-accent transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            Cancel
          </Link>
        }
      />

      {error && (
        <div className="p-4 bg-danger/10 border border-danger/20 rounded text-sm text-danger">
          {error}
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Database Details</CardTitle>
        </CardHeader>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1.5">Database Name</label>
            <Input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value.replace(/\s/g, '_').toLowerCase() })}
              placeholder="my_database"
              className="font-mono"
            />
            <p className="text-xs text-text-secondary mt-1">Only letters, numbers, and underscores allowed</p>
          </div>

          <div>
            <label className="block text-sm font-medium mb-3">Database Engine · Software managing your data</label>
            <div className="grid grid-cols-2 gap-3">
              {[
                { id: 'mysql', label: 'MySQL 8.0', desc: 'Most popular for web applications' },
                { id: 'postgresql', label: 'PostgreSQL 15', desc: 'Advanced features and performance' },
              ].map((engine) => (
                <button
                  key={engine.id}
                  onClick={() => setFormData({ ...formData, engine: engine.id })}
                  className={cn(
                    'p-4 border rounded text-left transition-colors',
                    formData.engine === engine.id
                      ? 'border-primary bg-primary/10'
                      : 'border-border hover:border-primary/50'
                  )}
                >
                  <p className="font-medium">{engine.label}</p>
                  <p className="text-xs text-text-secondary mt-0.5">{engine.desc}</p>
                </button>
              ))}
            </div>
          </div>
        </div>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Database User</CardTitle>
        </CardHeader>
        <div className="space-y-4">
          <Switch
            checked={formData.create_user}
            onChange={(e) => setFormData({ ...formData, create_user: e.target.checked })}
            label="Create a dedicated user for this database"
          />

          {formData.create_user && (
            <div className="space-y-3 pl-7">
              <div>
                <label className="block text-sm font-medium mb-1.5">Username</label>
                <Input
                  type="text"
                  value={formData.username}
                  onChange={(e) => setFormData({ ...formData, username: e.target.value.replace(/\s/g, '_').toLowerCase() })}
                  placeholder="db_user"
                  className="font-mono"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1.5">Password</label>
                <Input
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  placeholder="Strong password"
                />
              </div>
              <Switch
                checked={formData.grant_all}
                onChange={(e) => setFormData({ ...formData, grant_all: e.target.checked })}
                label="Grant all privileges · Give full database control"
              />
            </div>
          )}
        </div>
      </Card>

      <Button
        onClick={handleSubmit}
        loading={loading}
        disabled={!formData.name || (formData.create_user && !formData.username)}
        className="w-full"
      >
        Create Database
      </Button>
    </div>
  )
}