import { useState } from 'react'
import type { Device } from '@/types/device'

interface DeviceDetailsProps {
  id: string
}

export function DeviceDetails({ id }: DeviceDetailsProps) {
  // TODO: Replace with real API call
  const [device] = useState<Device>({
    id,
    name: 'pi-cluster-01',
    status: 'online',
    location: 'Rack 1, Unit 3',
    lastSeen: '2 minutes ago',
    temperature: 45,
    cpuLoad: 32,
    memoryUsage: 67,
    model: 'Raspberry Pi 4 Model B',
    serialNumber: 'RP4-123456',
    networkAddress: '192.168.1.101',
    physicalPosition: {
      rack: 'R1',
      unit: 3,
      coordinates: {
        x: 10,
        y: 20,
        z: 30
      }
    }
  })

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
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex items-center">
          <span className={`w-3 h-3 rounded-full mr-2 ${getStatusColor(device.status)}`} />
          <h2 className="text-xl font-semibold">{device.name}</h2>
        </div>
      </div>
      <div className="p-6">
        <div className="grid grid-cols-2 gap-6">
          <div>
            <h3 className="text-sm font-medium text-gray-500 mb-4">Device Information</h3>
            <dl className="space-y-3">
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Model</dt>
                <dd className="text-sm font-medium">{device.model}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Serial Number</dt>
                <dd className="text-sm font-medium">{device.serialNumber}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Network Address</dt>
                <dd className="text-sm font-medium">{device.networkAddress}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Last Seen</dt>
                <dd className="text-sm font-medium">{device.lastSeen}</dd>
              </div>
            </dl>
          </div>
          <div>
            <h3 className="text-sm font-medium text-gray-500 mb-4">Physical Location</h3>
            <dl className="space-y-3">
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Location</dt>
                <dd className="text-sm font-medium">{device.location}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Rack</dt>
                <dd className="text-sm font-medium">{device.physicalPosition?.rack}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Unit</dt>
                <dd className="text-sm font-medium">{device.physicalPosition?.unit}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Coordinates</dt>
                <dd className="text-sm font-medium">
                  {device.physicalPosition?.coordinates ? (
                    `(${device.physicalPosition.coordinates.x}, ${device.physicalPosition.coordinates.y}, ${device.physicalPosition.coordinates.z})`
                  ) : 'N/A'}
                </dd>
              </div>
            </dl>
          </div>
        </div>
      </div>
    </div>
  )
}