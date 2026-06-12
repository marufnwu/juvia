import { cn } from '../../lib/utils'

interface SkeletonProps {
  className?: string
}

export function Skeleton({ className }: SkeletonProps) {
  return <div className={cn('skeleton rounded', className)} />
}

export function TableSkeleton({ rows = 5, columns = 4 }: { rows?: number; columns?: number }) {
  return (
    <div className="border border-border rounded-card overflow-hidden">
      <table className="w-full">
        <thead className="bg-accent/50">
          <tr>
            {[...Array(columns)].map((_, i) => (
              <th key={i} className="text-left p-3">
                <Skeleton className="h-3 w-20" />
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {[...Array(rows)].map((_, i) => (
            <tr key={i} className="border-t border-border">
              {[...Array(columns)].map((_, j) => (
                <td key={j} className="p-3">
                  <Skeleton className="h-4 w-full" />
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export function CardSkeleton() {
  return (
    <div className="bg-surface border border-border rounded-card p-4 space-y-3">
      <div className="flex items-center gap-3">
        <Skeleton className="w-10 h-10 rounded" />
        <div className="space-y-2 flex-1">
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-3 w-24" />
        </div>
      </div>
      <Skeleton className="h-3 w-full" />
      <Skeleton className="h-3 w-3/4" />
    </div>
  )
}

export function MetricSkeleton() {
  return (
    <div className="bg-surface border border-border rounded-card p-4">
      <Skeleton className="h-3 w-20 mb-2" />
      <Skeleton className="h-6 w-24 mb-1" />
      <Skeleton className="h-3 w-16" />
    </div>
  )
}

export function ListSkeleton({ items = 5 }: { items?: number }) {
  return (
    <div className="space-y-2">
      {[...Array(items)].map((_, i) => (
        <div key={i} className="flex items-center gap-3 p-3 border border-border rounded-card">
          <Skeleton className="w-8 h-8 rounded" />
          <div className="space-y-2 flex-1">
            <Skeleton className="h-4 w-40" />
            <Skeleton className="h-3 w-24" />
          </div>
        </div>
      ))}
    </div>
  )
}