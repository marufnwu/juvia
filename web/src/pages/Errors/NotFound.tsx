import { Link } from 'react-router-dom'
import { Home, AlertTriangle } from 'lucide-react'
import { Button } from '../../components/ui/Button'

export default function NotFound() {
  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-surface border border-border rounded-card p-6 text-center">
        <div className="w-12 h-12 rounded-full bg-warning/10 flex items-center justify-center mx-auto mb-4">
          <AlertTriangle className="w-6 h-6 text-warning" />
        </div>
        <h1 className="text-4xl font-bold text-foreground mb-2">404</h1>
        <p className="text-lg font-medium text-foreground mb-2">Page not found</p>
        <p className="text-sm text-text-secondary mb-6">
          The page you are looking for does not exist or has been moved.
        </p>
        <Link to="/dashboard">
          <Button>
            <Home className="w-4 h-4 mr-2" />
            Back to Dashboard
          </Button>
        </Link>
      </div>
    </div>
  )
}
