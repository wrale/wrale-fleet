'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import type { Device } from '@/types/device'
import { ChevronRightIcon } from '@heroicons/react/24/outline'
import { deviceApi } from '@/services/api'
import { useLoading } from '@/components/ui/LoadingProvider'
import { TableRowSkeleton } from '@/components/ui/Skeleton'
import { ErrorBoundary } from '@/components/error/ErrorBoundary'

export function DeviceList() {
  const [devices, setDevices] = useState<Device[]>([])
  const [error, setError] = useState<string>()
  const { setIsLoading } = useLoading()

  useEffect(() => {
    async function fetchDevices() {
      try {
        setIsLoading(true)
        const data = await deviceApi.getAll()
        setDevices(data)
        setError(undefined)
      } catch (err) {
        setError('Failed to load devices')
        console.error('Error loading devices:', err)
      } finally {
        setIsLoading(false)
      }
    }

    fetchDevices()
  }, [setIsLoading])

  const getStatusColor = (status: Device['status']) => {
    switch (status) {
      case 'online':
        return 'bg-wrale-success'
      case 'warning':
        return 'bg-wrale-warning'
      case 'offline':
        return 'bg-wrale-danger'
    }
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4">
        <p className="text-red-800">{error}</p>
        <button 
          onClick={() => window.location.reload()}
          className="mt-2 text-red-600 hover:text-red-800 text-sm font-medium"
        >
          Retry
        </button>
      </div>
    )
  }

  return (
    <ErrorBoundary>
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="min-w-full divide-y divide-gray-200">
          <div className="bg-gray-50 px-6 py-3">
            <div className="grid grid-cols-12 gap-4">
              <div className="col-span-3">Name</div>
              <div className="col-span-2">Status</div>
              <div className="col-span-2">Location</div>
              <div className="col-span-2">Model</div>
              <div className="col-span-2">Last Seen</div>
              <div className="col-span-1"></div>
            </div>
          </div>

          <div className="divide-y divide-gray-200 bg-white">
            {devices.length === 0 ? (
              <>
                <TableRowSkeleton />
                <TableRowSkeleton />
                <TableRowSkeleton />
              </>
            ) : (
              devices.map((device) => (
                <Link
                  key={device.id}
                  href={`/devices/${device.id}`}
                  className="block hover:bg-gray-50 transition-colors"
                >
                  <div className="px-6 py-4">
                    <div className="grid grid-cols-12 gap-4 items-center">
                      <div className="col-span-3 font-medium text-gray-900">
                        {device.name}
                      </div>
                      <div className="col-span-2">
                        <span className="inline-flex items-center">
                          <span className={`w-2.5 h-2.5 rounded-full mr-2 ${getStatusColor(device.status)}`} />
                          {device.status.charAt(0).toUpperCase() + device.status.slice(1)}
                        </span>
                      </div>
                      <div className="col-span-2 text-gray-500">
                        {device.location}
                      </div>
                      <div className="col-span-2 text-gray-500">
                        {device.model}
                      </div>
                      <div className="col-span-2 text-gray-500">
                        {device.lastSeen}
                      </div>
                      <div className="col-span-1 text-right">
                        <ChevronRightIcon className="w-5 h-5 text-gray-400 inline-block" />
                      </div>
                    </div>
                  </div>
                </Link>
              ))
            )}
          </div>
        </div>
      </div>
    </ErrorBoundary>
  )
}