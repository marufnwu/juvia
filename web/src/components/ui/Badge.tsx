import { cn } from '../../lib/utils'

type BadgeVariant = 'success' | 'warning' | 'danger' | 'neutral' | 'pending' | 'info'

interface BadgeProps {
  variant?: BadgeVariant
  children: React.ReactNode
  className?: string
  dot?: boolean
  pulse?: boolean
}

const variantClasses: Record<BadgeVariant, string> = {
  success: 'bg-success/10 text-success border border-success/20',
  warning: 'bg-warning/10 text-warning border border-warning/20',
  danger: 'bg-danger/10 text-danger border border-danger/20',
  neutral: 'bg-neutral/10 text-neutral border border-neutral/20',
  pending: 'bg-primary/10 text-primary border border-primary/20',
  info: 'bg-blue-500/10 text-blue-400 border border-blue-500/20',
}

const dotClasses: Record<BadgeVariant, string> = {
  success: 'bg-success',
  warning: 'bg-warning',
  danger: 'bg-danger',
  neutral: 'bg-neutral',
  pending: 'bg-primary',
  info: 'bg-blue-400',
}

export function Badge({ variant = 'neutral', children, className, dot, pulse }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 px-2 py-0.5 text-xs font-medium rounded-tag',
        variantClasses[variant],
        className
      )}
    >
      {dot && (
        <span
          className={cn('w-1.5 h-1.5 rounded-full', dotClasses[variant], pulse && 'animate-pulse')}
        />
      )}
      {children}
    </span>
  )
}

export function StatusBadge({ status }: { status: string }) {
  const config: Record<string, { variant: BadgeVariant; label: string; pulse?: boolean }> = {
    active: { variant: 'success', label: 'Active', pulse: false },
    online: { variant: 'success', label: 'Online', pulse: true },
    running: { variant: 'success', label: 'Running', pulse: true },
    suspended: { variant: 'warning', label: 'Suspended' },
    pending: { variant: 'pending', label: 'Pending', pulse: true },
    deploying: { variant: 'pending', label: 'Deploying', pulse: true },
    stopped: { variant: 'neutral', label: 'Stopped' },
    offline: { variant: 'danger', label: 'Offline' },
    error: { variant: 'danger', label: 'Error' },
    failed: { variant: 'danger', label: 'Failed' },
  }

  const { variant, label, pulse } = config[status.toLowerCase()] || {
    variant: 'neutral' as BadgeVariant,
    label: status,
  }

  return (
    <Badge variant={variant} dot pulse={pulse}>
      {label}
    </Badge>
  )
}