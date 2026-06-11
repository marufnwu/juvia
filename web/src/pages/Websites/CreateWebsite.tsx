import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import api from '../../lib/api'

export default function CreateWebsite() {
  const [domain, setDomain] = useState('')
  const [phpVersion, setPhpVersion] = useState('8.2')
  const [webServer, setWebServer] = useState('nginx')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    if (!domain) {
      setError('Domain is required')
      setLoading(false)
      return
    }

    try {
      const res = await api.post('/websites', {
        domain,
        php_version: phpVersion,
        web_server: webServer,
      })
      if (res.data.data?.website_id) {
        navigate(`/websites/${res.data.data.website_id}`)
      }
    } catch (err: any) {
      setError(err.response?.data?.error?.user_message || 'Failed to create website')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-6 max-w-xl">
      <h1 className="text-lg font-semibold mb-6">Create Website</h1>

      {error && (
        <div className="mb-4 text-sm text-red-600 border border-red-200 bg-red-50 p-3">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        <div>
          <label className="block text-sm font-medium mb-1">Domain Name</label>
          <input
            type="text"
            value={domain}
            onChange={(e) => setDomain(e.target.value)}
            placeholder="example.com"
            className="w-full border border-border px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            required
          />
          <p className="mt-1 text-xs text-muted-foreground">
            Your website address (e.g., mysite.com)
          </p>
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">PHP Version</label>
          <select
            value={phpVersion}
            onChange={(e) => setPhpVersion(e.target.value)}
            className="w-full border border-border px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          >
            <option value="8.1">PHP 8.1</option>
            <option value="8.2">PHP 8.2</option>
            <option value="8.3">PHP 8.3</option>
          </select>
          <p className="mt-1 text-xs text-muted-foreground">
            The programming language your website's code runs on
          </p>
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Web Server</label>
          <select
            value={webServer}
            onChange={(e) => setWebServer(e.target.value)}
            className="w-full border border-border px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          >
            <option value="nginx">Nginx</option>
            <option value="apache">Apache</option>
          </select>
          <p className="mt-1 text-xs text-muted-foreground">
            The web server software that serves your website files
          </p>
        </div>

        <div className="flex gap-3 pt-4">
          <button
            type="submit"
            disabled={loading}
            className="bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90 disabled:opacity-50"
          >
            {loading ? 'Creating...' : 'Create Website'}
          </button>
          <button
            type="button"
            onClick={() => navigate('/websites')}
            className="border border-border px-4 py-2 text-sm hover:bg-muted"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  )
}
