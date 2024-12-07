import { useState } from 'react'
import type { Device } from '@/types/device'

interface DeviceEnvironmentProps {
  id: string
}

export function DeviceEnvironment({ id }: DeviceEnvironmentProps) {
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
    environmentalData: {
      humidity: 45,
      ambientLight: 500,
      airQuality: 95,
      vibration: 0.2
    },
    powerMetrics: {
      voltage: 5.1,
      current: 1.2,
      powerDraw: 6.12,
      efficiency: 92
    }
  })

  const getMetricColor = (value: number, thresholds: { warning: number; danger: number }) => {
    if (value >= thresholds.danger) return 'text-wrale-danger'
    if (value >= thresholds.warning) return 'text-wrale-warning'
    return 'text-wrale-success'
  }

  return (
    <div className="bg-white rounded-lg shadow overflow-hidden">
      <div className="px-6 py-4 border-b border-gray-200">
        <h2 className="text-xl font-semibold">Environmental Data</h2>
      </div>
      <div className="p-6">
        <div className="space-y-6">
          <div>
            <h3 className="text-sm font-medium text-gray-500 mb-4">Environmental Metrics</h3>
            <dl className="space-y-3">
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Temperature</dt>
                <dd className={`text-sm font-medium ${getMetricColor(device.temperature, { warning: 50, danger: 70 })}`}>
                  {device.temperature}°C
                </dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Humidity</dt>
                <dd className={`text-sm font-medium ${getMetricColor(device.environmentalData?.humidity || 0, { warning: 70, danger: 85 })}`}>
                  {device.environmentalData?.humidity}%
                </dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Air Quality</dt>
                <dd className={`text-sm font-medium ${getMetricColor(device.environmentalData?.airQuality || 0, { warning: 50, danger: 30 })}`}>
                  {device.environmentalData?.airQuality}/100
                </dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Ambient Light</dt>
                <dd className="text-sm font-medium">
                  {device.environmentalData?.ambientLight} lux
                </dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Vibration</dt>
                <dd className={`text-sm font-medium ${getMetricColor(device.environmentalData?.vibration || 0, { warning: 0.5, danger: 1.0 })}`}>
                  {device.environmentalData?.vibration} g
                </dd>
              </div>
            </dl>
          </div>

          <div>
            <h3 className="text-sm font-medium text-gray-500 mb-4">Power Metrics</h3>
            <dl className="space-y-3">
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Voltage</dt>
                <dd className={`text-sm font-medium ${getMetricColor(Math.abs(5 - (device.powerMetrics?.voltage || 5)), { warning: 0.3, danger: 0.5 })}`}>
                  {device.powerMetrics?.voltage}V
                </dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Current</dt>
                <dd className={`text-sm font-medium ${getMetricColor(device.powerMetrics?.current || 0, { warning: 2.0, danger: 2.5 })}`}>
                  {device.powerMetrics?.current}A
                </dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Power Draw</dt>
                <dd className="text-sm font-medium">
                  {device.powerMetrics?.powerDraw}W
                </dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-sm text-gray-600">Efficiency</dt>
                <dd className={`text-sm font-medium ${getMetricColor(100 - (device.powerMetrics?.efficiency || 100), { warning: 15, danger: 25 })}`}>
                  {device.powerMetrics?.efficiency}%
                </dd>
              </div>
            </dl>
          </div>
        </div>
      </div>
    </div>
  )
}