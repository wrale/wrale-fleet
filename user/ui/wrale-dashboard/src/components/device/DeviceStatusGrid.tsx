import { useState } from 'react'

interface Device {
  id: string
  name: string
  status: 'online' | 'offline' | 'warning'
  location: string
  lastSeen: string
  temperature: number
  cpuLoad: number
  memoryUsage: number
}

export function DeviceStatusGrid() {
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
      memoryUsage: 67
    },
    {
      id: '2',
      name: 'pi-cluster-02',
      status: 'warning',
      location: 'Rack 1, Unit 4',
      lastSeen: '1 minute ago',
      temperature: 52,
      cpuLoad: 89,
      memoryUsage: 78
    },
    {
      id: '3',
      name: 'pi-cluster-03',
      status: 'offline',
      location: 'Rack 2, Unit 1',
      lastSeen: '15 minutes ago',
      temperature: 0,
      cpuLoad: 0,
      memoryUsage: 0
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

  const getMetricColor = (value: number) => {
    if (value >= 80) return 'text-wrale-danger'
    if (value >= 60) return 'text-wrale-warning'
    return 'text-wrale-success'
  }

  return (
    <div className="bg-white rounded-lg shadow overflow-hidden">
      <div className="px-6 py-4 border-b border-gray-200">
        <h2 className="text-xl font-semibold">Device Status</h2>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 p-6">
        {devices.map((device) => (
          <div key={device.id} className="border rounded-lg p-4">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center">
                <div className={`w-3 h-3 rounded-full mr-2 ${getStatusColor(device.status)}`} />
                <h3 className="font-medium">{device.name}</h3>
              </div>
              <span className="text-sm text-gray-500">{device.lastSeen}</span>
            </div>
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span>Location:</span>
                <span>{device.location}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Temperature:</span>
                <span className={getMetricColor(device.temperature)}>
                  {device.temperature}°C
                </span>
              </div>
              <div className="flex justify-between text-sm">
                <span>CPU Load:</span>
                <span className={getMetricColor(device.cpuLoad)}>
                  {device.cpuLoad}%
                </span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Memory:</span>
                <span className={getMetricColor(device.memoryUsage)}>
                  {device.memoryUsage}%
                </span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}