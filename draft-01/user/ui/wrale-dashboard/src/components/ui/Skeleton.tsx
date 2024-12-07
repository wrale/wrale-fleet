interface SkeletonProps {
  className?: string
}

export function Skeleton({ className = '' }: SkeletonProps) {
  return (
    <div
      className={`animate-pulse bg-gray-200 rounded ${className}`}
    />
  )
}

export function DeviceCardSkeleton() {
  return (
    <div className="border rounded-lg p-4">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center">
          <Skeleton className="w-3 h-3 rounded-full mr-2" />
          <Skeleton className="h-6 w-32" />
        </div>
        <Skeleton className="w-20 h-4" />
      </div>
      <div className="space-y-2">
        <div className="flex justify-between">
          <Skeleton className="w-24 h-4" />
          <Skeleton className="w-24 h-4" />
        </div>
        <div className="flex justify-between">
          <Skeleton className="w-24 h-4" />
          <Skeleton className="w-24 h-4" />
        </div>
        <div className="flex justify-between">
          <Skeleton className="w-24 h-4" />
          <Skeleton className="w-24 h-4" />
        </div>
      </div>
    </div>
  )
}

export function MetricsChartSkeleton() {
  return (
    <div className="bg-white rounded-lg shadow overflow-hidden">
      <div className="px-6 py-4 border-b border-gray-200">
        <Skeleton className="h-8 w-48" />
      </div>
      <div className="p-6">
        <Skeleton className="h-80 w-full" />
      </div>
    </div>
  )
}

export function TableRowSkeleton() {
  return (
    <div className="px-6 py-4">
      <div className="grid grid-cols-12 gap-4 items-center">
        <div className="col-span-3">
          <Skeleton className="h-6 w-full" />
        </div>
        <div className="col-span-2">
          <Skeleton className="h-6 w-full" />
        </div>
        <div className="col-span-2">
          <Skeleton className="h-6 w-full" />
        </div>
        <div className="col-span-2">
          <Skeleton className="h-6 w-full" />
        </div>
        <div className="col-span-2">
          <Skeleton className="h-6 w-full" />
        </div>
        <div className="col-span-1">
          <Skeleton className="h-6 w-6 ml-auto" />
        </div>
      </div>
    </div>
  )
}