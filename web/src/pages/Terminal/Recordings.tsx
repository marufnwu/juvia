import { Link } from 'react-router-dom'
import { Play, Trash2, Clock } from 'lucide-react'
import { Card } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { PageHeader } from '../../components/ui/Misc'
import { cn } from '../../lib/utils'

interface Recording {
  id: string
  name: string
  duration: number
  created_at: string
  size: number
}

const mockRecordings: Recording[] = []

export default function Recordings() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Session Recordings"
        description="Recorded terminal sessions for review and audit"
        breadcrumbs={[
          { label: 'Terminal', href: '/terminal' },
          { label: 'Recordings' },
        ]}
      />

      <Card padding="none">
        <div className="p-8 text-center">
          <Play className="w-12 h-12 mx-auto text-text-secondary/50 mb-3" />
          <p className="text-text-secondary text-sm mb-3">No recordings yet</p>
          <p className="text-xs text-text-secondary">
            Terminal sessions are automatically recorded when enabled
          </p>
        </div>
      </Card>
    </div>
  )
}