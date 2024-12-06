import { DeviceDetails } from '@/components/device/DeviceDetails'
import { DeviceEnvironment } from '@/components/device/DeviceEnvironment'
import { DeviceMetrics } from '@/components/device/DeviceMetrics'

export default function DevicePage({ params }: { params: { id: string } }) {
  return (
    <div className="p-8">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <h1 className="text-3xl font-bold text-wrale-primary">Device Details</h1>
            <div className="space-x-4">
              <button className="px-4 py-2 rounded-lg border border-wrale-primary text-wrale-primary hover:bg-wrale-primary/5 transition-colors">
                Edit
              </button>
              <button className="px-4 py-2 rounded-lg bg-wrale-danger text-white hover:bg-wrale-danger/90 transition-colors">
                Shutdown
              </button>
              <button className="px-4 py-2 rounded-lg bg-wrale-primary text-white hover:bg-wrale-primary/90 transition-colors">
                Reboot
              </button>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2">
            <DeviceDetails id={params.id} />
            <div className="mt-6">
              <DeviceMetrics id={params.id} />
            </div>
          </div>
          <div>
            <DeviceEnvironment id={params.id} />
          </div>
        </div>
      </div>
    </div>
  )
}