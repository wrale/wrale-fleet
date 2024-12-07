import type { Device, DeviceCommand, DeviceConfig, Location } from '@/types/device'
import type { FleetCommand, FleetMetrics } from '@/types/fleet'
import type { WSMessage, WSConnection } from '@/types/ws'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
const API_WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'

// Auth token management
let authToken: string | null = null

export function setAuthToken(token: string) {
    authToken = token
    localStorage.setItem('auth_token', token)
}

export function getAuthToken(): string | null {
    if (!authToken) {
        authToken = localStorage.getItem('auth_token')
    }
    return authToken
}

export function clearAuthToken() {
    authToken = null
    localStorage.removeItem('auth_token')
}

// API request helper
async function fetchAPI<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const token = getAuthToken()
    const headers = {
        'Content-Type': 'application/json',
        ...(token ? { 'Authorization': token } : {}),
        ...options.headers,
    }

    const response = await fetch(`${API_BASE_URL}/api/v1${endpoint}`, {
        ...options,
        headers,
    })

    if (!response.ok) {
        // Handle authentication errors
        if (response.status === 401) {
            clearAuthToken()
            window.location.href = '/login'
            throw new Error('Authentication required')
        }

        const error = await response.json().catch(() => ({}))
        throw new Error(error.message || 'An error occurred')
    }

    const data = await response.json()
    return data.data as T // Unwrap API response
}

// WebSocket connection management
let wsConnection: WebSocket | null = null
const wsSubscribers = new Set<(msg: WSMessage) => void>()

export function connectWebSocket(deviceIds?: string[]): Promise<WSConnection> {
    return new Promise((resolve, reject) => {
        const token = getAuthToken()
        if (!token) {
            reject(new Error('Authentication required'))
            return
        }

        const url = new URL(API_WS_URL + '/api/v1/ws')
        if (deviceIds && deviceIds.length > 0) {
            deviceIds.forEach(id => url.searchParams.append('device', id))
        }

        wsConnection = new WebSocket(url.toString())
        
        wsConnection.onopen = () => {
            wsConnection!.send(JSON.stringify({ type: 'auth', token }))
            resolve({
                subscribe: (callback: (msg: WSMessage) => void) => {
                    wsSubscribers.add(callback)
                    return () => wsSubscribers.delete(callback)
                },
                close: () => wsConnection?.close(),
            })
        }

        wsConnection.onmessage = (event) => {
            const msg = JSON.parse(event.data) as WSMessage
            wsSubscribers.forEach(callback => callback(msg))
        }

        wsConnection.onerror = (error) => {
            reject(error)
        }

        wsConnection.onclose = () => {
            // Auto-reconnect after delay
            setTimeout(() => {
                if (wsConnection?.readyState === WebSocket.CLOSED) {
                    connectWebSocket(deviceIds).catch(console.error)
                }
            }, 5000)
        }
    })
}

// Device API
export const deviceApi = {
    list: () => 
        fetchAPI<Device[]>('/devices'),
    
    get: (id: string) => 
        fetchAPI<Device>(`/devices/${id}`),
    
    create: (device: Partial<Device>) =>
        fetchAPI<Device>('/devices', {
            method: 'POST',
            body: JSON.stringify(device),
        }),
    
    update: (id: string, updates: Partial<Device>) =>
        fetchAPI<Device>(`/devices/${id}`, {
            method: 'PUT',
            body: JSON.stringify(updates),
        }),
    
    delete: (id: string) =>
        fetchAPI<void>(`/devices/${id}`, {
            method: 'DELETE',
        }),
    
    executeCommand: (id: string, command: DeviceCommand) =>
        fetchAPI<any>(`/devices/${id}/command`, {
            method: 'POST',
            body: JSON.stringify(command),
        }),
}

// Fleet API
export const fleetApi = {
    executeCommand: (command: FleetCommand) =>
        fetchAPI<any>('/fleet/command', {
            method: 'POST',
            body: JSON.stringify(command),
        }),
    
    getMetrics: () =>
        fetchAPI<FleetMetrics>('/fleet/metrics'),
    
    updateConfig: (config: DeviceConfig, deviceIds?: string[]) =>
        fetchAPI<void>('/fleet/config', {
            method: 'PUT',
            body: JSON.stringify({
                config,
                devices: deviceIds,
            }),
        }),
    
    getConfig: (deviceIds?: string[]) =>
        fetchAPI<Record<string, DeviceConfig>>('/fleet/config', {
            method: 'GET',
            ...(deviceIds ? {
                headers: {
                    'X-Device-IDs': deviceIds.join(','),
                },
            } : {}),
        }),
}

// Auth API
export const authApi = {
    login: (username: string, password: string) =>
        fetchAPI<{ token: string }>('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ username, password }),
        }).then(data => {
            setAuthToken(data.token)
            return data
        }),
    
    logout: () => {
        clearAuthToken()
        if (wsConnection) {
            wsConnection.close()
            wsConnection = null
        }
    },
}

// Health check
export const healthApi = {
    check: () =>
        fetchAPI<{ status: string }>('/health'),
}
