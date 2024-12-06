import type { Device, DeviceGroup, Location } from '@/types/device'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

async function fetchWithAuth(endpoint: string, options: RequestInit = {}) {
  // TODO: Add real auth token handling
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new Error(error.message || 'An error occurred')
  }

  return response.json()
}

export const deviceApi = {
  getAll: () => fetchWithAuth('/devices'),
  getById: (id: string) => fetchWithAuth(`/devices/${id}`),
  getMetrics: (id: string) => fetchWithAuth(`/devices/${id}/metrics`),
  updateStatus: (id: string, status: Partial<Device>) =>
    fetchWithAuth(`/devices/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify(status),
    }),
}

export const locationApi = {
  getAll: () => fetchWithAuth('/locations'),
  getById: (id: string) => fetchWithAuth(`/locations/${id}`),
  getEnvironmentalData: (id: string) => 
    fetchWithAuth(`/locations/${id}/environmental`),
}

export const alertsApi = {
  getChannels: () => fetchWithAuth('/alerts/channels'),
  getRules: () => fetchWithAuth('/alerts/rules'),
  createChannel: (channel: any) =>
    fetchWithAuth('/alerts/channels', {
      method: 'POST',
      body: JSON.stringify(channel),
    }),
  createRule: (rule: any) =>
    fetchWithAuth('/alerts/rules', {
      method: 'POST',
      body: JSON.stringify(rule),
    }),
}

export const maintenanceApi = {
  getRules: () => fetchWithAuth('/maintenance/rules'),
  getPredictions: () => fetchWithAuth('/maintenance/predictions'),
  createRule: (rule: any) =>
    fetchWithAuth('/maintenance/rules', {
      method: 'POST',
      body: JSON.stringify(rule),
    }),
}

export const networkApi = {
  getConfig: () => fetchWithAuth('/network/config'),
  updateConfig: (config: any) =>
    fetchWithAuth('/network/config', {
      method: 'PUT',
      body: JSON.stringify(config),
    }),
  testConfig: (config: any) =>
    fetchWithAuth('/network/test', {
      method: 'POST',
      body: JSON.stringify(config),
    }),
}