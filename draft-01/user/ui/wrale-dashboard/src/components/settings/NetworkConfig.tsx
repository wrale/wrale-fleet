'use client'

interface NetworkConfig {
  id: string
  type: 'default' | 'custom'
  name: string
  subnet: string
  gateway: string
  dns: string[]
  dhcp: boolean
}

export function NetworkConfig() {
  return (
    <div className="space-y-6">
      <section>
        <h2 className="text-lg font-medium text-gray-900 mb-4">Network Settings</h2>
        <div className="bg-white shadow rounded-lg p-6">
          <div className="grid grid-cols-1 gap-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Network Mode
              </label>
              <select 
                className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-wrale-primary focus:border-wrale-primary sm:text-sm rounded-md"
                defaultValue="dhcp"
              >
                <option value="dhcp">DHCP</option>
                <option value="static">Static IP</option>
                <option value="mixed">Mixed Mode</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                IP Range Configuration
              </label>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-gray-500 mb-1">Subnet</label>
                  <input
                    type="text"
                    className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-wrale-primary focus:border-wrale-primary sm:text-sm"
                    placeholder="192.168.1.0/24"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-500 mb-1">Gateway</label>
                  <input
                    type="text"
                    className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-wrale-primary focus:border-wrale-primary sm:text-sm"
                    placeholder="192.168.1.1"
                  />
                </div>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                DNS Configuration
              </label>
              <div className="space-y-2">
                <div>
                  <label className="block text-sm text-gray-500 mb-1">Primary DNS</label>
                  <input
                    type="text"
                    className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-wrale-primary focus:border-wrale-primary sm:text-sm"
                    placeholder="8.8.8.8"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-500 mb-1">Secondary DNS</label>
                  <input
                    type="text"
                    className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-wrale-primary focus:border-wrale-primary sm:text-sm"
                    placeholder="8.8.4.4"
                  />
                </div>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Advanced Settings
              </label>
              <div className="space-y-4">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    className="form-checkbox h-4 w-4 text-wrale-primary"
                    defaultChecked
                  />
                  <span className="ml-2 text-sm text-gray-600">
                    Enable automatic IP conflict detection
                  </span>
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    className="form-checkbox h-4 w-4 text-wrale-primary"
                    defaultChecked
                  />
                  <span className="ml-2 text-sm text-gray-600">
                    Enable network monitoring
                  </span>
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    className="form-checkbox h-4 w-4 text-wrale-primary"
                  />
                  <span className="ml-2 text-sm text-gray-600">
                    Enable advanced network metrics
                  </span>
                </label>
              </div>
            </div>

            <div className="pt-4 border-t border-gray-200">
              <div className="flex justify-end space-x-4">
                <button className="px-4 py-2 text-sm text-wrale-primary border border-wrale-primary rounded-lg hover:bg-wrale-primary/5">
                  Test Configuration
                </button>
                <button className="px-4 py-2 text-sm text-white bg-wrale-primary rounded-lg hover:bg-wrale-primary/90">
                  Save Changes
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}