// Fleet operation types
export interface FleetCommand {
    operation: string
    devices: string[]
    params?: Record<string, any>
    timeout?: number
}

export interface FleetMetrics {
    totalDevices: number
    activeDevices: number
    totalPower: number
    avgTemp: number
    avgCpu: number
    avgMemory: number
    resourceUsage?: {
        cpu: number
        memory: number
        power: number
    }
    healthyDevices?: number
    alertCount?: number
    recommendationCount?: number
}

export interface DeviceGroup {
    devices: string[]
    config?: Record<string, any>
    validFrom?: string
    validTo?: string
}

export interface Alert {
    id: string
    deviceId?: string
    level: string
    message: string
    time: string
}

export interface Recommendation {
    id: string
    priority: number
    action: string
    reason: string
    deviceIds: string[]
    time: string
}
