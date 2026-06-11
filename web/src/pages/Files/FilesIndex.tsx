import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { FolderOpen } from 'lucide-react'
import api from '../../lib/api'

interface Website {
  id: number
  domain: string
  document_root: string
  status: string
}

export default function FilesIndex() {
  const [websites, setWebsites] = useState<Website[]>([])
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    fetchWebsites()
  }, [])

  const fetchWebsites = async () => {
    try {
      const res = await api.get('/websites')
      const data = res.data
      if (data.success) {
        setWebsites(data.data)
      }
    } catch (err) {
      console.error('Failed to fetch websites:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleSelectWebsite = (website: Website) => {
    navigate(`/files/${website.id}`)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-muted-foreground">Loading websites...</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">File Manager</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Browse and manage files for your websites
        </p>
      </div>

      {websites.length === 0 ? (
        <div className="text-center py-12 border rounded-lg">
          <FolderOpen className="mx-auto h-12 w-12 text-muted-foreground/50" />
          <h3 className="mt-4 text-lg font-medium">No websites found</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Create a website first to start managing files
          </p>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {websites.map((website) => (
            <button
              key={website.id}
              onClick={() => handleSelectWebsite(website)}
              className="border rounded-lg p-4 text-left hover:bg-accent/50 transition-colors group"
            >
              <div className="flex items-start gap-3">
                <FolderOpen className="h-8 w-8 text-primary mt-1" />
                <div className="flex-1 min-w-0">
                  <p className="font-medium truncate">{website.domain}</p>
                  <p className="text-sm text-muted-foreground truncate">
                    {website.document_root}
                  </p>
                  <div className="mt-2">
                    <span className={`text-xs px-2 py-0.5 rounded ${
                      website.status === 'active'
                        ? 'bg-green-500/10 text-green-500'
                        : 'bg-yellow-500/10 text-yellow-500'
                    }`}>
                      {website.status}
                    </span>
                  </div>
                </div>
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}