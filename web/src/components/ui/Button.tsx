import { forwardRef, ButtonHTMLAttributes } from 'react'
import { Loader2 } from 'lucide-react'
import { cn } from '../../lib/utils'

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'warning' | 'outline'
  size?: 'sm' | 'md' | 'lg' | 'icon'
  loading?: boolean
}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'primary', size = 'md', loading, disabled, children, ...props }, ref) => {
    const base = 'inline-flex items-center justify-center font-medium transition-colors duration-150 focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50 disabled:cursor-not-allowed'

    const variants = {
      primary: 'bg-primary text-white hover:bg-primary/90',
      secondary: 'bg-surface text-foreground hover:bg-accent',
      ghost: 'hover:bg-accent text-foreground',
      danger: 'bg-danger text-white hover:bg-danger/90',
      warning: 'bg-warning text-white hover:bg-warning/90',
      outline: 'border border-border bg-transparent hover:bg-accent text-foreground',
    }

    const sizes = {
      sm: 'h-8 px-3 text-xs rounded',
      md: 'h-9 px-4 text-sm rounded-md',
      lg: 'h-11 px-6 text-base rounded-md',
      icon: 'h-9 w-9 text-sm rounded-md',
    }

    return (
      <button
        ref={ref}
        className={cn(base, variants[variant], sizes[size], className)}
        disabled={disabled || loading}
        {...props}
      >
        {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
        {children}
      </button>
    )
  }
)

Button.displayName = 'Button'
export { Button }