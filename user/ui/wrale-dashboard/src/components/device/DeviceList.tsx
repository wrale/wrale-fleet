import { useState } from 'react'
import Link from 'next/link'
import type { Device } from '@/types/device'
import { ChevronRightIcon } from '@heroicons/react/24/outline'

export function DeviceList() {
  // TODO: Replace with real API call
  const [devices] = useState<Device[]>([
    {
      id: '1',
      name: 'pi-cluster-01',
      status: 'online',
      location: 'Rack 1, Unit 3',
      lastSeen: '2 minutes ago',
      temperature: 45,
      cpuLoad: 32,
      memoryUsage: 67,
      model: 'Raspberry Pi 4 Model B',
      serialNumber: 'RP4-123456',
      networkAddress: '192.168.1.101'
    },
    {
      id: '2',
      name: 'pi-cluster-02',
      status: 'warning',
      location: 'Rack 1, Unit 4',
      lastSeen: '1 minute ago',
      temperature: 52,
      cpuLoad: 89,
      memoryUsage: 78,
      model: 'Raspberry Pi 4 Model B',
      serialNumber: 'RP4-123457',
      networkAddress: '192.168.1.102'
    },
    {
      id: '3',
      name: 'pi-cluster-03',
      status: 'offline',
      location: 'Rack 2, Unit 1',
      lastSeen: '15 minutes ago',
      temperature: 0,
      cpuLoad: 0,
      memoryUsage: 0,
      model: 'Raspberry Pi 4 Model B',
      serialNumber: 'RP4-123458',
      networkAddress: '192.168.1.103'
    },
  ])

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

  return (
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
          {devices.map((device) => (
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
          ))}
        </div>
      </div>
    </div>
  )
}