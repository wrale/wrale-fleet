'use client'

import { useState, useEffect } from 'react'
import { FormInput, FormSelect } from '@/components/ui/Form'
import { validateNumber, validateRequired } from '@/lib/validation'
import { useLoading } from '@/components/ui/LoadingProvider'

interface RackConfig {
  id: string
  name: string
  location: string
  units: number
  maxPower: number
  coolingType: string
}

interface FormErrors {
  [key: string]: string | undefined
}

export function PhysicalConfig() {
  const { setIsLoading } = useLoading()
  const [racks, setRacks] = useState<RackConfig[]>([])
  const [errors, setErrors] = useState<FormErrors>({})
  const [environmentalSettings, setEnvironmentalSettings] = useState({
    tempWarning: '45',
    tempCritical: '55',
    humidityMin: '30',
    humidityMax: '70',
    powerWarning: '2500',
    powerCritical: '2800'
  })

  useEffect(() => {
    async function fetchRacks() {
      try {
        setIsLoading(true)
        // TODO: Replace with actual API call
        const sampleRacks = [
          {
            id: 'rack1',
            name: 'Rack 1',
            location: 'Room A',
            units: 42,
            maxPower: 3000,
            coolingType: 'Active Air'
          }
        ]
        setRacks(sampleRacks)
      } catch (error) {
        console.error('Failed to fetch racks:', error)
      } finally {
        setIsLoading(false)
      }
    }

    fetchRacks()
  }, [setIsLoading])

  const validateEnvironmentalSettings = () => {
    const newErrors: FormErrors = {}

    // Temperature validation
    newErrors.tempWarning = validateNumber(environmentalSettings.tempWarning, 0, 100)
    newErrors.tempCritical = validateNumber(environmentalSettings.tempCritical, 0, 100)
    if (!newErrors.tempWarning && !newErrors.tempCritical) {
      if (Number(environmentalSettings.tempWarning) >= Number(environmentalSettings.tempCritical)) {
        newErrors.tempWarning = 'Warning temperature must be lower than critical'
      }
    }

    // Humidity validation
    newErrors.humidityMin = validateNumber(environmentalSettings.humidityMin, 0, 100)
    newErrors.humidityMax = validateNumber(environmentalSettings.humidityMax, 0, 100)
    if (!newErrors.humidityMin && !newErrors.humidityMax) {
      if (Number(environmentalSettings.humidityMin) >= Number(environmentalSettings.humidityMax)) {
        newErrors.humidityMin = 'Minimum humidity must be lower than maximum'
      }
    }

    // Power validation
    newErrors.powerWarning = validateNumber(environmentalSettings.powerWarning, 0)
    newErrors.powerCritical = validateNumber(environmentalSettings.powerCritical, 0)
    if (!newErrors.powerWarning && !newErrors.powerCritical) {
      if (Number(environmentalSettings.powerWarning) >= Number(environmentalSettings.powerCritical)) {
        newErrors.powerWarning = 'Warning power must be lower than critical'
      }
    }

    setErrors(newErrors)
    return Object.values(newErrors).every(error => !error)
  }

  const handleUpdateEnvironmental = (key: keyof typeof environmentalSettings, value: string) => {
    setEnvironmentalSettings(prev => ({
      ...prev,
      [key]: value
    }))
    // Clear error when user starts typing
    if (errors[key]) {
      setErrors(prev => ({ ...prev, [key]: undefined }))
    }
  }

  const handleSave = async () => {
    if (!validateEnvironmentalSettings()) {
      return
    }

    try {
      setIsLoading(true)
      // TODO: Add API call to save settings
      await new Promise(resolve => setTimeout(resolve, 1000)) // Simulated API call
    } catch (error) {
      console.error('Failed to save settings:', error)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <section>
        <h2 className="text-lg font-medium text-gray-900 mb-4">Rack Configuration</h2>
        <div className="bg-white shadow rounded-lg divide-y divide-gray-200">
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
                <FormInput
                  label="Location"
                  value={rack.location}
                  disabled
                />
                <FormInput
                  label="Units"
                  value={`${rack.units}U`}
                  disabled
                />
                <FormInput
                  label="Max Power"
                  value={`${rack.maxPower}W`}
                  disabled
                />
                <FormInput
                  label="Cooling"
                  value={rack.coolingType}
                  disabled
                />
              </div>
            </div>
          ))}
        </div>
      </section>

      <section>
        <h2 className="text-lg font-medium text-gray-900 mb-4">Environmental Settings</h2>
        <div className="bg-white shadow rounded-lg p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <div className="space-y-4">
              <FormInput
                label="Warning Temperature"
                type="number"
                value={environmentalSettings.tempWarning}
                onChange={(e) => handleUpdateEnvironmental('tempWarning', e.target.value)}
                error={errors.tempWarning}
                required
                min="0"
                max="100"
                suffix="°C"
              />
              <FormInput
                label="Critical Temperature"
                type="number"
                value={environmentalSettings.tempCritical}
                onChange={(e) => handleUpdateEnvironmental('tempCritical', e.target.value)}
                error={errors.tempCritical}
                required
                min="0"
                max="100"
                suffix="°C"
              />
            </div>

            <div className="space-y-4">
              <FormInput
                label="Minimum Humidity"
                type="number"
                value={environmentalSettings.humidityMin}
                onChange={(e) => handleUpdateEnvironmental('humidityMin', e.target.value)}
                error={errors.humidityMin}
                required
                min="0"
                max="100"
                suffix="%"
              />
              <FormInput
                label="Maximum Humidity"
                type="number"
                value={environmentalSettings.humidityMax}
                onChange={(e) => handleUpdateEnvironmental('humidityMax', e.target.value)}
                error={errors.humidityMax}
                required
                min="0"
                max="100"
                suffix="%"
              />
            </div>

            <div className="space-y-4">
              <FormInput
                label="Power Warning Threshold"
                type="number"
                value={environmentalSettings.powerWarning}
                onChange={(e) => handleUpdateEnvironmental('powerWarning', e.target.value)}
                error={errors.powerWarning}
                required
                min="0"
                suffix="W"
              />
              <FormInput
                label="Power Critical Threshold"
                type="number"
                value={environmentalSettings.powerCritical}
                onChange={(e) => handleUpdateEnvironmental('powerCritical', e.target.value)}
                error={errors.powerCritical}
                required
                min="0"
                suffix="W"
              />
            </div>
          </div>

          <div className="mt-6 flex justify-end">
            <button
              onClick={handleSave}
              className="px-4 py-2 bg-wrale-primary text-white rounded-lg hover:bg-wrale-primary/90 transition-colors"
            >
              Save Changes
            </button>
          </div>
        </div>
      </section>
    </div>
  )
}