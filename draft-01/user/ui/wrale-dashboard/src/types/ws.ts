import type { Device, DeviceMetrics } from './device'
import type { Alert } from './fleet'

// WebSocket message types
export interface WSMessage {
    type: string
    payload: WSPayload
}

export type WSPayload = 
    | WSStateUpdate
    | WSMetricsUpdate
    | WSAlertMessage
    | WSError

export interface WSStateUpdate {
    deviceId: string
    state: Device
    time: string
}

export interface WSMetricsUpdate {
    deviceId: string
    metrics: DeviceMetrics
    time: string
}

export interface WSAlertMessage {
    deviceId?: string
    level: string
    message: string
    time: string
}

export interface WSError {
    code: string
    message: string
    details?: string
}

// WebSocket connection types
export interface WSConnection {
    subscribe: (callback: (msg: WSMessage) => void) => () => void
    close: () => void
}
