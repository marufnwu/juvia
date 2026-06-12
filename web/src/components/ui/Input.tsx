import { forwardRef, InputHTMLAttributes, TextareaHTMLAttributes } from 'react'
import { cn } from '../../lib/utils'

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  error?: boolean
  icon?: React.ReactNode
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, error, icon, ...props }, ref) => {
    return (
      <div className="relative">
        {icon && (
          <div className="absolute left-3 top-1/2 -translate-y-1/2 text-text-secondary">
            {icon}
          </div>
        )}
        <input
          ref={ref}
          className={cn(
            'w-full h-9 px-3 bg-surface border border-border rounded text-sm text-foreground placeholder:text-text-secondary focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-colors duration-150',
            error && 'border-danger focus:ring-danger/50 focus:border-danger',
            icon && 'pl-9',
            className
          )}
          {...props}
        />
      </div>
    )
  }
)

Input.displayName = 'Input'

export interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  error?: boolean
}

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, error, ...props }, ref) => {
    return (
      <textarea
        ref={ref}
        className={cn(
          'w-full px-3 py-2 bg-surface border border-border rounded text-sm text-foreground placeholder:text-text-secondary focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-colors duration-150 resize-none font-mono',
          error && 'border-danger focus:ring-danger/50 focus:border-danger',
          className
        )}
        {...props}
      />
    )
  }
)

Textarea.displayName = 'Textarea'

export interface SelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  error?: boolean
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  ({ className, error, children, ...props }, ref) => {
    return (
      <select
        ref={ref}
        className={cn(
          'w-full h-9 px-3 bg-surface border border-border rounded text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-colors duration-150 cursor-pointer',
          error && 'border-danger focus:ring-danger/50 focus:border-danger',
          className
        )}
        {...props}
      >
        {children}
      </select>
    )
  }
)

Select.displayName = 'Select'

export interface CheckboxProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
  label?: string
}

export const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(
  ({ className, label, ...props }, ref) => {
    return (
      <label className="inline-flex items-center gap-2 cursor-pointer">
        <input
          ref={ref}
          type="checkbox"
          className={cn(
            'w-4 h-4 rounded border border-border bg-surface text-primary focus:ring-2 focus:ring-primary/50 cursor-pointer',
            className
          )}
          {...props}
        />
        {label && <span className="text-sm text-foreground">{label}</span>}
      </label>
    )
  }
)

Checkbox.displayName = 'Checkbox'

export interface SwitchProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
  label?: string
}

export const Switch = forwardRef<HTMLInputElement, SwitchProps>(
  ({ className, label, checked, onChange, ...props }, ref) => {
    return (
      <label className="inline-flex items-center gap-3 cursor-pointer">
        <div className="relative">
          <input
            ref={ref}
            type="checkbox"
            checked={checked}
            onChange={onChange}
            className="sr-only peer"
            {...props}
          />
          <div className="w-9 h-5 bg-border rounded-full peer-checked:bg-primary transition-colors" />
          <div className="absolute left-0.5 top-0.5 w-4 h-4 bg-white rounded-full peer-checked:translate-x-4 transition-transform" />
        </div>
        {label && <span className="text-sm text-foreground">{label}</span>}
      </label>
    )
  }
)

Switch.displayName = 'Switch'

export function Label({ children, className, htmlFor }: { children: React.ReactNode; className?: string; htmlFor?: string }) {
  return (
    <label
      htmlFor={htmlFor}
      className={cn('text-sm font-medium text-foreground mb-1.5 block', className)}
    >
      {children}
    </label>
  )
}

export function FormGroup({ children, className }: { children: React.ReactNode; className?: string }) {
  return <div className={cn('space-y-1', className)}>{children}</div>
}

export function FormError({ children }: { children: React.ReactNode }) {
  return <p className="text-xs text-danger mt-1">{children}</p>
}