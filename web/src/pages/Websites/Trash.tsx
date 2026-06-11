import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import api from '../../lib/api'

interface Website {
  id: number
  domain: string
  deleted_at: string
}

export default function Trash() {
  const [websites, setWebsites] = useState<Website[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get('/websites/trash')
      .then((res) => setWebsites(res.data.data || []))
      .catch((err) => console.error(err))
      .finally(() => setLoading(false))
  }, [])

  const handleRestore = async (id: number) => {
    if (!confirm('Restore this website?')) return
    try {
      await api.post(`/websites/${id}/restore`)
      setWebsites(websites.filter(w => w.id !== id))
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to restore')
    }
  }

  const handlePermanentDelete = async (id: number) => {
    if (!confirm('Permanently delete this website? This cannot be undone.')) return
    try {
      await api.delete(`/websites/${id}/permanent`)
      setWebsites(websites.filter(w => w.id !== id))
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to delete')
    }
  }

  const handlePurgeAll = async () => {
    if (!confirm('Permanently delete all trashed websites? This cannot be undone.')) return
    try {
      await api.delete('/websites/trash/purge')
      setWebsites([])
    } catch (err: any) {
      alert(err.response?.data?.error?.user_message || 'Failed to purge')
    }
  }

  if (loading) {
    return <div className="p-6 text-muted-foreground">Loading...</div>
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-lg font-semibold">Trash</h1>
        {websites.length > 0 && (
          <button
            onClick={handlePurgeAll}
            className="text-sm text-red-600 hover:underline"
          >
            Purge All
          </button>
        )}
      </div>

      {websites.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          <p>Trash is empty</p>
        </div>
      ) : (
        <div className="border border-border">
          <table className="w-full text-sm">
            <thead className="bg-muted/50 border-b border-border">
              <tr>
                <th className="text-left p-3 font-medium">Domain</th>
                <th className="text-left p-3 font-medium">Deleted</th>
                <th className="text-left p-3 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {websites.map((site) => (
                <tr key={site.id} className="border-b border-border">
                  <td className="p-3">
                    <Link to={`/websites/${site.id}`} className="text-primary hover:underline">
                      {site.domain}
                    </Link>
                  </td>
                  <td className="p-3 text-muted-foreground">
                    {new Date(site.deleted_at).toLocaleDateString()}
                  </td>
                  <td className="p-3">
                    <div className="flex gap-2">
                      <button
                        onClick={() => handleRestore(site.id)}
                        className="text-green-600 hover:underline text-xs"
                      >
                        Restore
                      </button>
                      <button
                        onClick={() => handlePermanentDelete(site.id)}
                        className="text-red-600 hover:underline text-xs"
                      >
                        Delete Permanently
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
