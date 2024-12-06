import { DeviceList } from '@/components/device/DeviceList'

export default function DevicesPage() {
  return (
    <div className="p-8">
      <div className="max-w-7xl mx-auto">
        <div className="flex justify-between items-center mb-8">
          <h1 className="text-3xl font-bold text-wrale-primary">Devices</h1>
          <button className="bg-wrale-primary text-white px-4 py-2 rounded-lg hover:bg-wrale-primary/90 transition-colors">
            Add Device
          </button>
        </div>
        
        <div className="mb-6 flex space-x-4">
          <div className="flex-1">
            <input
              type="text"
              placeholder="Search devices..."
              className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-wrale-primary/50"
            />
          </div>
          <select className="px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-wrale-primary/50">
            <option value="">All Locations</option>
            <option value="rack1">Rack 1</option>
            <option value="rack2">Rack 2</option>
          </select>
          <select className="px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-wrale-primary/50">
            <option value="">All Statuses</option>
            <option value="online">Online</option>
            <option value="offline">Offline</option>
            <option value="warning">Warning</option>
          </select>
        </div>

        <DeviceList />
      </div>
    </div>
  )
}