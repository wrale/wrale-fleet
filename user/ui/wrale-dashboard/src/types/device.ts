// Device types
export interface Device {
    id: string
    status: string
    location: Location
    metrics: DeviceMetrics
    config: DeviceConfig
    lastUpdate: string
}

export interface Location {
    rack: string
    position: number
    zone: string
}

export interface DeviceMetrics {
    temperature: number
    powerUsage: number
    cpuLoad: number
    memoryUsage: number
}

export interface DeviceConfig {
    [key: string]: any
}

export interface DeviceCommand {
    operation: string
    params?: Record<string, any>
    timeout?: number
}

export interface CommandResponse {
    id: string
    status: string
    startTime: string
    endTime?: string
    result?: any
    error?: string
}

// Update request types
export interface DeviceCreateRequest {
    id: string
    location: Location
    config?: DeviceConfig
}

export interface DeviceUpdateRequest {
    status?: string
    location?: Location
    config?: DeviceConfig
}
