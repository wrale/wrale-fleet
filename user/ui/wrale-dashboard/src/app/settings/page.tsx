import { TabGroup } from '@/components/settings/TabGroup'
import { PhysicalConfig } from '@/components/settings/PhysicalConfig'
import { AlertSettings } from '@/components/settings/AlertSettings'
import { MaintenanceRules } from '@/components/settings/MaintenanceRules'
import { NetworkConfig } from '@/components/settings/NetworkConfig'

export default function SettingsPage() {
  const tabs = [
    {
      id: 'physical',
      label: 'Physical Configuration',
      content: <PhysicalConfig />
    },
    {
      id: 'alerts',
      label: 'Alert Settings',
      content: <AlertSettings />
    },
    {
      id: 'maintenance',
      label: 'Maintenance Rules',
      content: <MaintenanceRules />
    },
    {
      id: 'network',
      label: 'Network Settings',
      content: <NetworkConfig />
    }
  ]

  return (
    <div className="p-8">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-wrale-primary">Settings</h1>
          <p className="mt-2 text-gray-600">
            Configure your fleet management settings and physical infrastructure
          </p>
        </div>

        <TabGroup tabs={tabs} />
      </div>
    </div>
  )
}