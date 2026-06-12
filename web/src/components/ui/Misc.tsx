import { useState } from 'react'
import { Copy, Check } from 'lucide-react'
import { cn } from '../../lib/utils'

interface CopyButtonProps {
  text: string
  className?: string
}

export function CopyButton({ text, className }: CopyButtonProps) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }

  return (
    <button
      onClick={handleCopy}
      className={cn(
        'p-1.5 border border-border rounded hover:bg-accent transition-colors',
        className
      )}
      title="Copy to clipboard"
    >
      {copied ? (
        <Check className="w-3.5 h-3.5 text-success" />
      ) : (
        <Copy className="w-3.5 h-3.5 text-text-secondary" />
      )}
    </button>
  )
}

interface ProgressBarProps {
  value: number
  max?: number
  variant?: 'primary' | 'success' | 'warning' | 'danger'
  size?: 'sm' | 'md'
  showLabel?: boolean
  className?: string
}

export function ProgressBar({
  value,
  max = 100,
  variant = 'primary',
  size = 'md',
  showLabel,
  className,
}: ProgressBarProps) {
  const percentage = Math.min((value / max) * 100, 100)

  const variantClasses = {
    primary: 'bg-primary',
    success: 'bg-success',
    warning: 'bg-warning',
    danger: 'bg-danger',
  }

  const sizeClasses = {
    sm: 'h-1',
    md: 'h-2',
  }

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <div className={cn('flex-1 bg-accent rounded-full overflow-hidden', sizeClasses[size])}>
        <div
          className={cn('h-full rounded-full transition-all duration-300', variantClasses[variant])}
          style={{ width: `${percentage}%` }}
        />
      </div>
      {showLabel && (
        <span className="text-xs text-text-secondary min-w-[3rem] text-right">
          {Math.round(percentage)}%
        </span>
      )}
    </div>
  )
}

interface HealthIndicatorProps {
  status: 'healthy' | 'warning' | 'critical' | 'unknown'
  label?: string
  pulse?: boolean
}

export function HealthIndicator({ status, label, pulse }: HealthIndicatorProps) {
  const statusConfig = {
    healthy: { color: 'bg-success', label: 'Healthy' },
    warning: { color: 'bg-warning', label: 'Warning' },
    critical: { color: 'bg-danger', label: 'Critical' },
    unknown: { color: 'bg-neutral', label: 'Unknown' },
  }

  const config = statusConfig[status]

  return (
    <div className="flex items-center gap-2">
      <span
        className={cn('w-2 h-2 rounded-full', config.color, pulse && 'animate-pulse')}
      />
      {label && <span className="text-sm text-text-secondary">{label || config.label}</span>}
    </div>
  )
}

interface EmptyStateProps {
  icon?: React.ReactNode
  title: string
  description?: string
  action?: React.ReactNode
  className?: string
}

export function EmptyState({ icon, title, description, action, className }: EmptyStateProps) {
  return (
    <div className={cn('flex flex-col items-center justify-center py-12 text-center', className)}>
      {icon && <div className="mb-4 text-text-secondary">{icon}</div>}
      <h3 className="text-sm font-medium text-foreground mb-1">{title}</h3>
      {description && <p className="text-xs text-text-secondary mb-4 max-w-xs">{description}</p>}
      {action}
    </div>
  )
}

interface PageHeaderProps {
  title: string
  description?: string
  breadcrumbs?: { label: string; href?: string }[]
  actions?: React.ReactNode
  className?: string
}

export function PageHeader({ title, description, breadcrumbs, actions, className }: PageHeaderProps) {
  return (
    <div className={cn('mb-6', className)}>
      {breadcrumbs && breadcrumbs.length > 0 && (
        <div className="flex items-center gap-2 text-xs text-text-secondary mb-2">
          {breadcrumbs.map((crumb, i) => (
            <span key={i} className="flex items-center gap-2">
              {i > 0 && <span>/</span>}
              {crumb.href ? (
                <a href={crumb.href} className="hover:text-foreground transition-colors">
                  {crumb.label}
                </a>
              ) : (
                <span>{crumb.label}</span>
              )}
            </span>
          ))}
        </div>
      )}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-lg font-semibold text-foreground">{title}</h1>
          {description && <p className="text-sm text-text-secondary mt-0.5">{description}</p>}
        </div>
        {actions && <div className="flex items-center gap-2">{actions}</div>}
      </div>
    </div>
  )
}