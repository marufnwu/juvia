import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Globe, FolderOpen, Search } from 'lucide-react'
import { Card } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { SearchInput } from '../../components/ui/SearchInput'
import { PageHeader } from '../../components/ui/Misc'
import api from '../../lib/api'
import { formatDate } from '../../lib/utils'
import { cn } from '../../lib/utils'

interface Website {
  id: number
  domain: string
}

export default function FilesIndex() {
  const navigate = useNavigate()
  const [websites, setWebsites] = useState<Website[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')

  useEffect(() => {
    loadWebsites()
  }, [])

  const loadWebsites = async () => {
    try {
      const res = await api.get('/websites')
      setWebsites(res.data.data || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const filteredWebsites = websites.filter((w) =>
    w.domain.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <div className="space-y-6">
      <PageHeader
        title="File Manager"
        description="Select a website to manage its files"
        breadcrumbs={[{ label: 'Files' }]}
      />

      <div className="w-64">
        <SearchInput
          value={search}
          onChange={setSearch}
          placeholder="Search websites..."
        />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {loading ? (
          [...Array(6)].map((_, i) => (
            <Card key={i} className="p-4">
              <div className="h-12 skeleton rounded mb-2" />
              <div className="h-4 w-24 skeleton rounded" />
            </Card>
          ))
        ) : filteredWebsites.length === 0 ? (
          <div className="col-span-full text-center py-12 text-text-secondary">
            No websites found
          </div>
        ) : (
          filteredWebsites.map((site) => (
            <Card
              key={site.id}
              hover
              className="cursor-pointer"
              onClick={() => navigate(`/files/${site.id}`)}
            >
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded bg-accent/50 flex items-center justify-center">
                  <Globe className="w-5 h-5 text-text-secondary" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="font-medium truncate">{site.domain}</p>
                  <p className="text-xs text-text-secondary">Click to manage files</p>
                </div>
                <FolderOpen className="w-5 h-5 text-text-secondary" />
              </div>
            </Card>
          ))
        )}
      </div>
    </div>
  )
}