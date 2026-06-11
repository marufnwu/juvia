import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import api from '../../lib/api'

export default function CreateDatabase() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [engine, setEngine] = useState<'mysql' | 'postgresql'>('mysql')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    if (!name.match(/^[a-zA-Z0-9_]{1,64}$/)) {
      setError('Name must be alphanumeric or underscore, max 64 characters')
      return
    }

    setLoading(true)
    try {
      await api.post('/databases', { name, engine })
      navigate('/databases')
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to create database')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-6 max-w-md">
      <h1 className="text-lg font-semibold mb-6">Create Database</h1>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium mb-1">Database Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full border border-border px-3 py-2 text-sm"
            placeholder="my_database"
            pattern="^[a-zA-Z0-9_]{1,64}$"
            required
          />
          <p className="text-xs text-muted-foreground mt-1">
            Letters, numbers, and underscores only. Max 64 characters.
          </p>
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">Database Type</label>
          <div className="flex gap-4">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                name="engine"
                value="mysql"
                checked={engine === 'mysql'}
                onChange={() => setEngine('mysql')}
                className="cursor-pointer"
              />
              <span className="text-sm">MySQL</span>
            </label>
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                name="engine"
                value="postgresql"
                checked={engine === 'postgresql'}
                onChange={() => setEngine('postgresql')}
                className="cursor-pointer"
              />
              <span className="text-sm">PostgreSQL</span>
            </label>
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            MySQL · Most popular database for web applications. PostgreSQL · Advanced database with more features.
          </p>
        </div>

        {error && <div className="text-red-600 text-sm">{error}</div>}

        <div className="flex gap-2 pt-2">
          <button
            type="submit"
            disabled={loading}
            className="bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90 disabled:opacity-50"
          >
            {loading ? 'Creating...' : 'Create Database'}
          </button>
          <button
            type="button"
            onClick={() => navigate('/databases')}
            className="border border-border px-4 py-2 text-sm hover:bg-muted"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  )
}