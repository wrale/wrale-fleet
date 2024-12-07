export type DeviceStatus = 'online' | 'offline' | 'warning'

export interface Device {
  id: string
  name: string
  status: DeviceStatus
  location: string
  lastSeen: string
  temperature: number
  cpuLoad: number
  memoryUsage: number
  model?: string
  serialNumber?: string
  networkAddress?: string
  physicalPosition?: {
    rack: string
    unit: number
    coordinates?: {
      x: number
      y: number
      z: number
    }
  }
  environmentalData?: {
    humidity?: number
    ambientLight?: number
    airQuality?: number
    vibration?: number
  }
  powerMetrics?: {
    voltage: number
    current: number
    powerDraw: number
    efficiency: number
  }
}

export interface DeviceGroup {
  id: string
  name: string
  description?: string
  devices: Device[]
  location?: string
  tags?: string[]
}

export interface Location {
  id: string
  name: string
  type: 'room' | 'rack' | 'shelf' | 'zone'
  parent?: string // Parent location ID
  coordinates?: {
    x: number
    y: number
    z: number
  }
  dimensions?: {
    width: number
    height: number
    depth: number
  }
  environmentalData?: {
    temperature: number
    humidity: number
    airQuality?: number
  }
}