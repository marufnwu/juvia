import { cn } from '../../lib/utils'
import { ChevronUp, ChevronDown, ChevronsUpDown } from 'lucide-react'

interface Column<T> {
  key: string
  header: string
  width?: string
  sortable?: boolean
  render?: (row: T) => React.ReactNode
  className?: string
}

interface TableProps<T> {
  columns: Column<T>[]
  data: T[]
  keyField: keyof T
  onRowClick?: (row: T) => void
  loading?: boolean
  emptyMessage?: string
  className?: string
}

export function Table<T extends Record<string, any>>({
  columns,
  data,
  keyField,
  onRowClick,
  loading,
  emptyMessage = 'No data available',
  className,
}: TableProps<T>) {
  const [sortKey, setSortKey] = React.useState<string | null>(null)
  const [sortOrder, setSortOrder] = React.useState<'asc' | 'desc'>('asc')

  const handleSort = (key: string) => {
    if (sortKey === key) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortOrder('asc')
    }
  }

  const sortedData = React.useMemo(() => {
    if (!sortKey) return data
    return [...data].sort((a, b) => {
      const aVal = a[sortKey]
      const bVal = b[sortKey]
      if (aVal === bVal) return 0
      const cmp = aVal < bVal ? -1 : 1
      return sortOrder === 'asc' ? cmp : -cmp
    })
  }, [data, sortKey, sortOrder])

  if (loading) {
    return (
      <div className="border border-border rounded-card overflow-hidden">
        <table className="w-full">
          <thead className="bg-accent/50">
            <tr>
              {columns.map((col) => (
                <th key={col.key} className="text-left p-3 text-xs font-medium text-text-secondary uppercase tracking-wider">
                  {col.header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {[...Array(5)].map((_, i) => (
              <tr key={i} className="border-t border-border">
                {columns.map((col) => (
                  <td key={col.key} className="p-3">
                    <div className="h-4 w-20 skeleton rounded" />
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    )
  }

  return (
    <div className={cn('border border-border rounded-card overflow-hidden', className)}>
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead className="bg-accent/50 border-b border-border">
            <tr>
              {columns.map((col) => (
                <th
                  key={col.key}
                  className={cn(
                    'text-left p-3 text-xs font-medium text-text-secondary uppercase tracking-wider',
                    col.sortable && 'cursor-pointer hover:text-foreground transition-colors',
                    col.className
                  )}
                  style={{ width: col.width }}
                  onClick={() => col.sortable && handleSort(col.key)}
                >
                  <div className="flex items-center gap-1">
                    {col.header}
                    {col.sortable && (
                      <span className="ml-1">
                        {sortKey === col.key ? (
                          sortOrder === 'asc' ? (
                            <ChevronUp className="w-3 h-3" />
                          ) : (
                            <ChevronDown className="w-3 h-3" />
                          )
                        ) : (
                          <ChevronsUpDown className="w-3 h-3 opacity-50" />
                        )}
                      </span>
                    )}
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {sortedData.length === 0 ? (
              <tr>
                <td colSpan={columns.length} className="p-8 text-center text-text-secondary text-sm">
                  {emptyMessage}
                </td>
              </tr>
            ) : (
              sortedData.map((row) => (
                <tr
                  key={String(row[keyField])}
                  className={cn(
                    'border-t border-border hover:bg-accent/30 transition-colors',
                    onRowClick && 'cursor-pointer'
                  )}
                  onClick={() => onRowClick?.(row)}
                >
                  {columns.map((col) => (
                    <td key={col.key} className={cn('p-3 text-sm', col.className)}>
                      {col.render ? col.render(row) : row[col.key]}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}

import React from 'react'