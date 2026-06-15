import { useState, useEffect, useRef } from 'react'
import { Search, X } from 'lucide-react'
import { cn } from '../../lib/utils'

interface SearchInputProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  className?: string
  debounceMs?: number
}

export function SearchInput({
  value,
  onChange,
  placeholder = 'Search...',
  className,
  debounceMs = 300,
}: SearchInputProps) {
  const [localValue, setLocalValue] = useState(value)
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [onChange])

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value
    setLocalValue(newValue)

    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }

    timeoutRef.current = setTimeout(() => {
      onChange(newValue)
    }, debounceMs)
  }

  const handleClear = () => {
    setLocalValue('')
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }
    onChange('')
  }

  return (
    <div className={cn('relative', className)}>
      <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-secondary" />
      <input
        type="text"
        value={localValue}
        onChange={handleChange}
        placeholder={placeholder}
        className="w-full h-9 pl-9 pr-8 bg-surface border border-border rounded text-sm text-foreground placeholder:text-text-secondary focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-colors"
      />
      {localValue && (
        <button
          onClick={handleClear}
          className="absolute right-3 top-1/2 -translate-y-1/2 p-0.5 text-text-secondary hover:text-foreground"
        >
          <X className="w-3.5 h-3.5" />
        </button>
      )}
    </div>
  )
}