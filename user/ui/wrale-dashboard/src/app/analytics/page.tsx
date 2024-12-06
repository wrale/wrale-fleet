import { PowerUsageChart } from '@/components/analytics/PowerUsageChart'
import { TemperatureHeatmap } from '@/components/analytics/TemperatureHeatmap'
import { PerformanceMetrics } from '@/components/analytics/PerformanceMetrics'
import { MaintenancePredictor } from '@/components/analytics/MaintenancePredictor'

export default function AnalyticsPage() {
  return (
    <div className="p-8">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-wrale-primary">Fleet Analytics</h1>
          <p className="mt-2 text-gray-600">
            Physical-first analysis of fleet performance and environmental impact
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200">
              <h2 className="text-xl font-semibold">Power Consumption</h2>
            </div>
            <div className="p-6">
              <PowerUsageChart />
            </div>
          </div>

          <div className="bg-white rounded-lg shadow overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200">
              <h2 className="text-xl font-semibold">Temperature Distribution</h2>
            </div>
            <div className="p-6">
              <TemperatureHeatmap />
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2">
            <div className="bg-white rounded-lg shadow overflow-hidden">
              <div className="px-6 py-4 border-b border-gray-200">
                <h2 className="text-xl font-semibold">Performance Trends</h2>
              </div>
              <div className="p-6">
                <PerformanceMetrics />
              </div>
            </div>
          </div>

          <div>
            <div className="bg-white rounded-lg shadow overflow-hidden">
              <div className="px-6 py-4 border-b border-gray-200">
                <h2 className="text-xl font-semibold">Maintenance Analysis</h2>
              </div>
              <div className="p-6">
                <MaintenancePredictor />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}