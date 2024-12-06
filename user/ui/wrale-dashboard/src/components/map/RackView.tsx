import Link from 'next/link'
import type { Device } from '@/types/device'

interface RackViewProps {
  id: string
  name: string
}

export function RackView({ id, name }: RackViewProps) {
  // TODO: Replace with real API call
  const devices: Device[] = [
    {
      id: '1',
      name: 'pi-cluster-01',
      status: 'online',
      location: 'Unit 3',
      temperature: 45,
      cpuLoad: 32,
      memoryUsage: 67,
      lastSeen: '2m ago'
    },
    {
      id: '2',
      name: 'pi-cluster-02',
      status: 'warning',
      location: 'Unit 4',
      temperature: 52,
      cpuLoad: 89,
      memoryUsage: 78,
      lastSeen: '1m ago'
    }
  ]

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
    <div>
      <h3 className="text-lg font-medium mb-2">{name}</h3>
      <div className="space-y-2">
        {devices.map((device) => (
          <Link
            key={device.id}
            href={`/devices/${device.id}`}
            className="block bg-gray-50 rounded-lg p-3 hover:bg-gray-100 transition-colors"
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center">
                <div className={`w-2 h-2 rounded-full ${getStatusColor(device.status)} mr-2`} />
                <span className="font-medium">{device.name}</span>
              </div>
              <span className="text-sm text-gray-500">{device.lastSeen}</span>
            </div>
            <div className="mt-2 grid grid-cols-3 gap-2 text-sm text-gray-600">
              <div>
                <span className="block text-gray-500">Temp</span>
                <span>{device.temperature}°C</span>
              </div>
              <div>
                <span className="block text-gray-500">CPU</span>
                <span>{device.cpuLoad}%</span>
              </div>
              <div>
                <span className="block text-gray-500">RAM</span>
                <span>{device.memoryUsage}%</span>
              </div>
            </div>
          </Link>
        ))}
      </div>
    </div>
  )
}