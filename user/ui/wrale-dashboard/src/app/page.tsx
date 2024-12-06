import Link from 'next/link'
import { DeviceStatusGrid } from '@/components/device/DeviceStatusGrid'

export default function Home() {
  return (
    <main className="min-h-screen p-8">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-3xl font-bold text-wrale-primary mb-8">
          Wrale Fleet Dashboard
        </h1>
        
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
          <div className="bg-white p-6 rounded-lg shadow">
            <h2 className="text-xl font-semibold mb-4">Fleet Overview</h2>
            <div className="grid grid-cols-2 gap-4">
              <div className="text-center">
                <p className="text-2xl font-bold text-wrale-success">24</p>
                <p className="text-sm text-gray-600">Online</p>
              </div>
              <div className="text-center">
                <p className="text-2xl font-bold text-wrale-danger">2</p>
                <p className="text-sm text-gray-600">Offline</p>
              </div>
            </div>
          </div>
          
          <div className="bg-white p-6 rounded-lg shadow">
            <h2 className="text-xl font-semibold mb-4">System Health</h2>
            <div className="space-y-2">
              <div className="flex justify-between items-center">
                <span>CPU Load</span>
                <span className="text-wrale-warning">72%</span>
              </div>
              <div className="flex justify-between items-center">
                <span>Memory</span>
                <span className="text-wrale-success">45%</span>
              </div>
              <div className="flex justify-between items-center">
                <span>Storage</span>
                <span className="text-wrale-success">38%</span>
              </div>
            </div>
          </div>
          
          <div className="bg-white p-6 rounded-lg shadow">
            <h2 className="text-xl font-semibold mb-4">Environmental</h2>
            <div className="space-y-2">
              <div className="flex justify-between items-center">
                <span>Temperature</span>
                <span className="text-wrale-success">24°C</span>
              </div>
              <div className="flex justify-between items-center">
                <span>Humidity</span>
                <span className="text-wrale-success">45%</span>
              </div>
              <div className="flex justify-between items-center">
                <span>Power</span>
                <span className="text-wrale-success">Stable</span>
              </div>
            </div>
          </div>
        </div>

        <DeviceStatusGrid />
      </div>
    </main>
  )
}