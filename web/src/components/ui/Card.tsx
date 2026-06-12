import { cn } from '../../lib/utils'

interface CardProps {
  className?: string
  children: React.ReactNode
  hover?: boolean
  padding?: 'none' | 'sm' | 'md' | 'lg'
  onClick?: () => void
}

export function Card({ className, children, hover = false, padding = 'md', onClick }: CardProps) {
  const paddingClasses = {
    none: '',
    sm: 'p-3',
    md: 'p-4',
    lg: 'p-6',
  }

  return (
    <div
      onClick={onClick}
      className={cn(
        'bg-surface border border-border rounded-card',
        hover && 'transition-colors duration-150 hover:border-primary/50 cursor-pointer',
        paddingClasses[padding],
        className
      )}
    >
      {children}
    </div>
  )
}

export function CardHeader({ className, children }: { className?: string; children: React.ReactNode }) {
  return <div className={cn('font-semibold text-base mb-3', className)}>{children}</div>
}

export function CardTitle({ className, children }: { className?: string; children: React.ReactNode }) {
  return <h3 className={cn('font-semibold text-sm', className)}>{children}</h3>
}

export function CardDescription({ className, children }: { className?: string; children: React.ReactNode }) {
  return <p className={cn('text-text-secondary text-xs mt-1', className)}>{children}</p>
}

export function CardContent({ className, children }: { className?: string; children: React.ReactNode }) {
  return <div className={className}>{children}</div>
}

export function CardFooter({ className, children }: { className?: string; children: React.ReactNode }) {
  return <div className={cn('mt-4 pt-4 border-t border-border flex items-center gap-2', className)}>{children}</div>
}