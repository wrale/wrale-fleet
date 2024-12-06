'use client'

import { useState } from 'react'

interface RackConfig {
  id: string
  name: string
  location: string
  units: number
  maxPower: number
  coolingType: string
}

export function PhysicalConfig() {
  const [racks] = useState<RackConfig[]>([
    {
      id: 'rack1',
      name: 'Rack 1',
      location: 'Room A',
      units: 42,
      maxPower: 3000,
      coolingType: 'Active Air'
    },
    {
      id: 'rack2',
      name: 'Rack 2',
      location: 'Room A',
      units: 42,
      maxPower: 3000,
      coolingType: 'Active Air'
    }
  ])

  return (
    <div className="space-y-6">
      <section>
        <h2 className="text-lg font-medium text-gray-900 mb-4">Rack Configuration</h2>
        <div className="bg-white shadow rounded-lg">
          <div className="divide-y divide-gray-200">
            {racks.map((rack) => (
              <div key={rack.id} className="p-6">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-medium">{rack.name}</h3>
                  <div className="space-x-3">
                    <button className="text-wrale-primary hover:text-wrale-primary/80 text-sm font-medium">
                      Edit
                    </button>
                    <button className="text-wrale-danger hover:text-wrale-danger/80 text-sm font-medium">
                      Delete
                    </button>
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-500">Location</label>
                    <p className="mt-1">{rack.location}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-500">Units</label>
                    <p className="mt-1">{rack.units}U</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-500">Max Power</label>
                    <p className="mt-1">{rack.maxPower}W</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-500">Cooling</label>
                    <p className="mt-1">{rack.coolingType}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        <button className="mt-4 flex items-center text-sm text-wrale-primary hover:text-wrale-primary/80">
          <svg className="w-5 h-5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Add Rack
        </button>
      </section>

      <section>
        <h2 className="text-lg font-medium text-gray-900 mb-4">Environmental Settings</h2>
        <div className="bg-white shadow rounded-lg p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Temperature Limits
              </label>
              <div className="space-y-2">
                <div className="flex items-center">
                  <span className="text-sm text-gray-500 w-20">Warning</span>
                  <input
                    type="number"
                    className="block w-20 rounded-md border-gray-300 shadow-sm focus:border-wrale-primary focus:ring-wrale-primary sm:text-sm"
                    defaultValue={45}
                  />
                  <span className="ml-2 text-sm text-gray-500">°C</span>
                </div>
                <div className="flex items-center">
                  <span className="text-sm text-gray-500 w-20">Critical</span>
                  <input
                    type="number"
                    className="block w-20 rounded-md border-gray-300 shadow-sm focus:border-wrale-primary focus:ring-wrale-primary sm:text-sm"
                    defaultValue={55}
                  />
                  <span className="ml-2 text-sm text-gray-500">°C</span>
                </div>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Humidity Range
              </label>
              <div className="space-y-2">
                <div className="flex items-center">
                  <span className="text-sm text-gray-500 w-20">Min</span>
                  <input
                    type="number"
                    className="block w-20 rounded-md border-gray-300 shadow-sm focus:border-wrale-primary focus:ring-wrale-primary sm:text-sm"
                    defaultValue={30}
                  />
                  <span className="ml-2 text-sm text-gray-500">%</span>
                </div>
                <div className="flex items-center">
                  <span className="text-sm text-gray-500 w-20">Max</span>
                  <input
                    type="number"
                    className="block w-20 rounded-md border-gray-300 shadow-sm focus:border-wrale-primary focus:ring-wrale-primary sm:text-sm"
                    defaultValue={70}
                  />
                  <span className="ml-2 text-sm text-gray-500">%</span>
                </div>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Power Settings
              </label>
              <div className="space-y-2">
                <div className="flex items-center">
                  <span className="text-sm text-gray-500 w-20">Warning</span>
                  <input
                    type="number"
                    className="block w-20 rounded-md border-gray-300 shadow-sm focus:border-wrale-primary focus:ring-wrale-primary sm:text-sm"
                    defaultValue={2500}
                  />
                  <span className="ml-2 text-sm text-gray-500">W</span>
                </div>
                <div className="flex items-center">
                  <span className="text-sm text-gray-500 w-20">Critical</span>
                  <input
                    type="number"
                    className="block w-20 rounded-md border-gray-300 shadow-sm focus:border-wrale-primary focus:ring-wrale-primary sm:text-sm"
                    defaultValue={2800}
                  />
                  <span className="ml-2 text-sm text-gray-500">W</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}