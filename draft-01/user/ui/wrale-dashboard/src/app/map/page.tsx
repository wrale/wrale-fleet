import { PhysicalMap } from '@/components/map/PhysicalMap'
import { RackView } from '@/components/map/RackView'

export default function MapPage() {
  return (
    <div className="p-8">
      <div className="max-w-7xl mx-auto">
        <div className="flex justify-between items-center mb-8">
          <h1 className="text-3xl font-bold text-wrale-primary">Physical Layout</h1>
          <button className="bg-wrale-primary text-white px-4 py-2 rounded-lg hover:bg-wrale-primary/90 transition-colors">
            Edit Layout
          </button>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          <div className="lg:col-span-3">
            <div className="bg-white rounded-lg shadow p-6">
              <PhysicalMap />
            </div>
          </div>
          
          <div>
            <div className="bg-white rounded-lg shadow p-6 mb-6">
              <h2 className="text-lg font-semibold mb-4">Environmental Overview</h2>
              <div className="space-y-4">
                <div>
                  <div className="flex justify-between text-sm mb-1">
                    <span>Room Temperature</span>
                    <span className="font-medium">24°C</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div className="bg-wrale-success rounded-full h-2" style={{ width: '40%' }}></div>
                  </div>
                </div>
                <div>
                  <div className="flex justify-between text-sm mb-1">
                    <span>Humidity</span>
                    <span className="font-medium">45%</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div className="bg-wrale-success rounded-full h-2" style={{ width: '45%' }}></div>
                  </div>
                </div>
                <div>
                  <div className="flex justify-between text-sm mb-1">
                    <span>Power Usage</span>
                    <span className="font-medium">2.4kW</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div className="bg-wrale-warning rounded-full h-2" style={{ width: '75%' }}></div>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-lg font-semibold mb-4">Rack Status</h2>
              <div className="space-y-6">
                <RackView id="rack1" name="Rack 1" />
                <RackView id="rack2" name="Rack 2" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}