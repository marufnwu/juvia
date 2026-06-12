import { useState, useEffect } from 'react'
import { X, CheckCircle, AlertCircle, AlertTriangle, Info } from 'lucide-react'
import { cn } from '../../lib/utils'

type ToastVariant = 'success' | 'error' | 'warning' | 'info'

interface Toast {
  id: string
  variant: ToastVariant
  title: string
  message?: string
  duration?: number
}

const variantConfig: Record<ToastVariant, { icon: React.ReactNode; className: string }> = {
  success: { icon: <CheckCircle className="w-4 h-4 text-success" />, className: 'border-success/30' },
  error: { icon: <AlertCircle className="w-4 h-4 text-danger" />, className: 'border-danger/30' },
  warning: { icon: <AlertTriangle className="w-4 h-4 text-warning" />, className: 'border-warning/30' },
  info: { icon: <Info className="w-4 h-4 text-primary" />, className: 'border-primary/30' },
}

interface ToastProps extends Toast {
  onDismiss: (id: string) => void
}

function ToastItem({ id, variant, title, message, duration = 5000, onDismiss }: ToastProps) {
  useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(() => onDismiss(id), duration)
      return () => clearTimeout(timer)
    }
  }, [id, duration, onDismiss])

  const { icon, className } = variantConfig[variant]

  return (
    <div
      className={cn(
        'flex items-start gap-3 p-4 bg-surface border rounded-card shadow-lg toast-enter min-w-[320px] max-w-[400px]',
        className
      )}
    >
      {icon}
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-foreground">{title}</p>
        {message && <p className="text-xs text-text-secondary mt-0.5">{message}</p>}
      </div>
      <button
        onClick={() => onDismiss(id)}
        className="p-0.5 text-text-secondary hover:text-foreground transition-colors"
      >
        <X className="w-3.5 h-3.5" />
      </button>
    </div>
  )
}

// Toast store
let toasts: Toast[] = []
let listeners: ((toasts: Toast[]) => void)[] = []

function notify(listeners: ((toasts: Toast[]) => void)[], toasts: Toast[]) {
  listeners.forEach(l => l(toasts))
}

export const toastStore = {
  add: (toast: Omit<Toast, 'id'>) => {
    const id = Math.random().toString(36).slice(2)
    toasts = [...toasts, { ...toast, id }]
    notify(listeners, toasts)
  },
  remove: (id: string) => {
    toasts = toasts.filter(t => t.id !== id)
    notify(listeners, toasts)
  },
  subscribe: (listener: (toasts: Toast[]) => void) => {
    listeners = [...listeners, listener]
    return () => {
      listeners = listeners.filter(l => l !== listener)
    }
  },
}

export function toast(options: Omit<Toast, 'id'>) {
  toastStore.add(options)
}

export function ToastContainer() {
  const [localToasts, setLocalToasts] = useState<Toast[]>([])

  useEffect(() => {
    return toastStore.subscribe(setLocalToasts)
  }, [])

  if (localToasts.length === 0) return null

  return (
    <div className="fixed bottom-4 right-4 z-[100] flex flex-col gap-2">
      {localToasts.map(t => (
        <ToastItem key={t.id} {...t} onDismiss={toastStore.remove} />
      ))}
    </div>
  )
}

export function useToast() {
  return { toast }
}