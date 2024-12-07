import type { PhysicalDevice } from '@/types/physical'

const METAL_API_BASE = process.env.NEXT_PUBLIC_METAL_API_URL || 'http://localhost:3001'

async function fetchFromMetal(endpoint: string, options: RequestInit = {}) {
  const response = await fetch(`${METAL_API_BASE}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  })

  if (!response.ok) {
    throw new Error(`Metal API error: ${response.statusText}`)
  }

  return response.json()
}

export const metalApi = {
  // Hardware-level interactions
  async getPhysicalDevices(): Promise<PhysicalDevice[]> {
    return fetchFromMetal('/devices')
  },

  async getPhysicalDevice(id: string): Promise<PhysicalDevice> {
    return fetchFromMetal(`/devices/${id}`)
  },

  async getSubsystemStatus(deviceId: string, subsystem: string) {
    return fetchFromMetal(`/devices/${deviceId}/subsystems/${subsystem}`)
  },

  // Environmental monitoring
  async getRackEnvironment(rackId: string) {
    return fetchFromMetal(`/racks/${rackId}/environment`)
  },

  async getLocationTemperatureMap() {
    return fetchFromMetal('/environment/temperature-map')
  },

  // Power management
  async getPowerMetrics(deviceId: string) {
    return fetchFromMetal(`/devices/${deviceId}/power`)
  },

  async setDevicePowerState(deviceId: string, state: 'on' | 'off' | 'restart') {
    return fetchFromMetal(`/devices/${deviceId}/power`, {
      method: 'POST',
      body: JSON.stringify({ state }),
    })
  },

  // Thermal management
  async getThermalMetrics(deviceId: string) {
    return fetchFromMetal(`/devices/${deviceId}/thermal`)
  },

  async setFanSpeed(deviceId: string, speed: number) {
    return fetchFromMetal(`/devices/${deviceId}/thermal/fan`, {
      method: 'POST',
      body: JSON.stringify({ speed }),
    })
  },

  // Storage health
  async getStorageHealth(deviceId: string) {
    return fetchFromMetal(`/devices/${deviceId}/storage/health`)
  },

  // Diagnostics
  async runDiagnostics(deviceId: string, tests: string[]) {
    return fetchFromMetal(`/devices/${deviceId}/diagnostics`, {
      method: 'POST',
      body: JSON.stringify({ tests }),
    })
  },

  async getMaintenanceHistory(deviceId: string) {
    return fetchFromMetal(`/devices/${deviceId}/maintenance`)
  },

  // Physical operations that require confirmation
  async validatePhysicalOperation(deviceId: string, operation: string) {
    return fetchFromMetal(`/devices/${deviceId}/validate-operation`, {
      method: 'POST',
      body: JSON.stringify({ operation }),
    })
  },

  async confirmPhysicalOperation(deviceId: string, operation: string, confirmationToken: string) {
    return fetchFromMetal(`/devices/${deviceId}/confirm-operation`, {
      method: 'POST',
      body: JSON.stringify({ operation, confirmationToken }),
    })
  }
}