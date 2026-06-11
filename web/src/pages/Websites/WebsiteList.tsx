import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import api from '../../lib/api'

interface Website {
  id: number
  domain: string
  php_version: string
  web_server: string
  ssl_enabled: boolean
  ssl_expiry: string | null
  status: string
  created_at: string
}

interface ListResponse {
  success: boolean
  data: Website[]
  meta: {
    total: number
    page: number
    limit: number
    pages: number
  }
}

export default function WebsiteList() {
  const [websites, setWebsites] = useState<Website[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const navigate = useNavigate()

  useEffect(() => {
    api.get('/websites')
      .then((res) => {
        const data = res.data as ListResponse
        setWebsites(data.data || [])
      })
      .catch((err) => {
        setError(err.response?.data?.error?.user_message || 'Failed to load websites')
      })
      .finally(() => setLoading(false))
  }, [])

  const handleSuspend = async (id: number) => {
    if (!confirm('Suspend this website?')) return
    try {
      await api.post(`/websites/${id}/suspend`)
      navigate(0)
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to suspend')
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Move this website to trash?')) return
    try {
      await api.delete(`/websites/${id}`)
      setWebsites(websites.filter(w => w.id !== id))
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to delete')
    }
  }

  if (loading) {
    return <div className="p-6 text-muted-foreground">Loading websites...</div>
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-lg font-semibold">Websites</h1>
        <Link
          to="/websites/create"
          className="bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90"
        >
          Create Website
        </Link>
      </div>

      {error && (
        <div className="mb-4 text-sm text-red-600 border border-red-200 bg-red-50 p-3">
          {error}
        </div>
      )}

      {websites.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          <p className="mb-2">No websites yet</p>
          <Link to="/websites/create" className="text-primary hover:underline">
            Create your first website
          </Link>
        </div>
      ) : (
        <div className="border border-border">
          <table className="w-full text-sm">
            <thead className="bg-muted/50 border-b border-border">
              <tr>
                <th className="text-left p-3 font-medium">Domain</th>
                <th className="text-left p-3 font-medium">PHP Version</th>
                <th className="text-left p-3 font-medium">Web Server</th>
                <th className="text-left p-3 font-medium">SSL</th>
                <th className="text-left p-3 font-medium">Status</th>
                <th className="text-left p-3 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {websites.map((site) => (
                <tr key={site.id} className="border-b border-border hover:bg-muted/30">
                  <td className="p-3">
                    <Link to={`/websites/${site.id}`} className="text-primary hover:underline">
                      {site.domain}
                    </Link>
                  </td>
                  <td className="p-3 text-muted-foreground">
                    PHP {site.php_version}
                    <span className="block text-xs text-muted-foreground">
                      The programming language your website's code runs on
                    </span>
                  </td>
                  <td className="p-3 text-muted-foreground capitalize">{site.web_server}</td>
                  <td className="p-3">
                    {site.ssl_enabled ? (
                      <span className="text-green-600">Active</span>
                    ) : (
                      <span className="text-muted-foreground">None</span>
                    )}
                  </td>
                  <td className="p-3">
                    <span className={`px-2 py-0.5 text-xs rounded ${
                      site.status === 'active' ? 'bg-green-100 text-green-700' :
                      site.status === 'suspended' ? 'bg-yellow-100 text-yellow-700' :
                      'bg-gray-100 text-gray-700'
                    }`}>
                      {site.status}
                    </span>
                  </td>
                  <td className="p-3">
                    <div className="flex gap-2">
                      <Link
                        to={`/websites/${site.id}`}
                        className="text-primary hover:underline text-xs"
                      >
                        View
                      </Link>
                      {site.status === 'active' && (
                        <button
                          onClick={() => handleSuspend(site.id)}
                          className="text-yellow-600 hover:underline text-xs"
                        >
                          Suspend
                        </button>
                      )}
                      <button
                        onClick={() => handleDelete(site.id)}
                        className="text-red-600 hover:underline text-xs"
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
