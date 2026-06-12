import { useEffect, useCallback } from 'react'
import { X } from 'lucide-react'
import { cn } from '../../lib/utils'

interface ModalProps {
  open: boolean
  onClose: () => void
  title?: string
  description?: string
  children: React.ReactNode
  size?: 'sm' | 'md' | 'lg' | 'xl'
  className?: string
}

export function Modal({ open, onClose, title, description, children, size = 'md', className }: ModalProps) {
  const handleEscape = useCallback((e: KeyboardEvent) => {
    if (e.key === 'Escape') onClose()
  }, [onClose])

  useEffect(() => {
    if (open) {
      document.addEventListener('keydown', handleEscape)
      document.body.style.overflow = 'hidden'
    }
    return () => {
      document.removeEventListener('keydown', handleEscape)
      document.body.style.overflow = ''
    }
  }, [open, handleEscape])

  if (!open) return null

  const sizeClasses = {
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-lg',
    xl: 'max-w-xl',
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm modal-enter"
        onClick={onClose}
      />
      <div
        className={cn(
          'relative w-full mx-4 bg-surface border border-border rounded-card shadow-xl modal-enter',
          sizeClasses[size],
          className
        )}
      >
        {title && (
          <div className="flex items-start justify-between p-4 border-b border-border">
            <div>
              <h2 className="text-base font-semibold text-foreground">{title}</h2>
              {description && (
                <p className="text-xs text-text-secondary mt-0.5">{description}</p>
              )}
            </div>
            <button
              onClick={onClose}
              className="p-1 text-text-secondary hover:text-foreground transition-colors"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        )}
        <div className="p-4">{children}</div>
      </div>
    </div>
  )
}

export function Drawer({ open, onClose, title, children, side = 'right' }: {
  open: boolean
  onClose: () => void
  title?: string
  children: React.ReactNode
  side?: 'left' | 'right'
}) {
  const handleEscape = useCallback((e: KeyboardEvent) => {
    if (e.key === 'Escape') onClose()
  }, [onClose])

  useEffect(() => {
    if (open) {
      document.addEventListener('keydown', handleEscape)
      document.body.style.overflow = 'hidden'
    }
    return () => {
      document.removeEventListener('keydown', handleEscape)
      document.body.style.overflow = ''
    }
  }, [open, handleEscape])

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex">
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm modal-enter"
        onClick={onClose}
      />
      <div
        className={cn(
          'relative bg-surface border-border shadow-xl modal-enter flex flex-col',
          side === 'right' ? 'ml-auto w-full max-w-md border-l' : 'mr-auto w-full max-w-md border-r'
        )}
      >
        {title && (
          <div className="flex items-center justify-between p-4 border-b border-border">
            <h2 className="text-base font-semibold text-foreground">{title}</h2>
            <button
              onClick={onClose}
              className="p-1 text-text-secondary hover:text-foreground transition-colors"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        )}
        <div className="flex-1 overflow-auto p-4">{children}</div>
      </div>
    </div>
  )
}

interface ConfirmModalProps {
  open: boolean
  onClose: () => void
  onConfirm: () => void
  title: string
  description?: string
  confirmLabel?: string
  variant?: 'danger' | 'primary'
  loading?: boolean
}

export function ConfirmModal({
  open, onClose, onConfirm, title, description, confirmLabel = 'Confirm', variant = 'primary', loading
}: ConfirmModalProps) {
  return (
    <Modal open={open} onClose={onClose} title={title} description={description} size="sm">
      <div className="flex justify-end gap-3 mt-4">
        <button
          onClick={onClose}
          className="px-4 py-2 text-sm border border-border rounded hover:bg-accent transition-colors"
        >
          Cancel
        </button>
        <button
          onClick={onConfirm}
          disabled={loading}
          className={cn(
            'px-4 py-2 text-sm font-medium rounded transition-colors disabled:opacity-50',
            variant === 'danger'
              ? 'bg-danger text-white hover:bg-danger/90'
              : 'bg-primary text-white hover:bg-primary/90'
          )}
        >
          {loading ? 'Loading...' : confirmLabel}
        </button>
      </div>
    </Modal>
  )
}