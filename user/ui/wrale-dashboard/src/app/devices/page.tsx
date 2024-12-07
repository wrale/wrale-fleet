"use client"

import { useState, useEffect } from 'react'
import { deviceApi, connectWebSocket } from '@/services/api'
import { Device, DeviceCommand } from '@/types/device'
import { WSMessage } from '@/types/ws'
import { useAuth } from '@/components/auth/AuthProvider'
import { DeviceList } from '@/components/device/DeviceList'
import { DeviceStatusGrid } from '@/components/device/DeviceStatusGrid'
import { ErrorBoundary } from '@/components/error/ErrorBoundary'

export default function DevicesPage() {
    const [devices, setDevices] = useState<Device[]>([])
    const [view, setView] = useState<'list' | 'grid'>('grid')
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const { isAuthenticated } = useAuth()

    useEffect(() => {
        if (!isAuthenticated) return

        // Load initial devices
        loadDevices()

        // Set up WebSocket connection
        const setupWebSocket = async () => {
            try {
                const ws = await connectWebSocket()
                ws.subscribe(handleWebSocketMessage)
            } catch (err) {
                console.error('WebSocket connection failed:', err)
            }
        }

        setupWebSocket()
    }, [isAuthenticated])

    const loadDevices = async () => {
        try {
            setLoading(true)
            const data = await deviceApi.list()
            setDevices(data)
            setError(null)
        } catch (err) {
            setError('Failed to load devices')
            console.error(err)
        } finally {
            setLoading(false)
        }
    }

    const handleWebSocketMessage = (msg: WSMessage) => {
        switch (msg.type) {
            case 'state_update':
                const update = msg.payload
                setDevices(current => 
                    current.map(device => 
                        device.id === update.deviceId
                            ? { ...device, ...update.state }
                            : device
                    )
                )
                break

            case 'device_removed':
                const { device_id } = msg.payload as { device_id: string }
                setDevices(current =>
                    current.filter(d => d.id !== device_id)
                )
                break
        }
    }

    const executeCommand = async (deviceId: string, command: DeviceCommand) => {
        try {
            await deviceApi.executeCommand(deviceId, command)
            // Device state will be updated via WebSocket
        } catch (err) {
            console.error('Command execution failed:', err)
            throw err
        }
    }

    if (error) {
        return (
            <div className="p-4">
                <div className="bg-red-50 border border-red-400 text-red-700 px-4 py-3 rounded">
                    <strong>Error:</strong> {error}
                    <button 
                        onClick={loadDevices}
                        className="ml-4 text-sm underline"
                    >
                        Try Again
                    </button>
                </div>
            </div>
        )
    }

    return (
        <ErrorBoundary>
            <div className="p-4 space-y-4">
                <div className="flex justify-between items-center">
                    <h1 className="text-2xl font-bold">Devices</h1>
                    <div className="flex space-x-2">
                        <button
                            onClick={() => setView('list')}
                            className={`px-3 py-1 rounded ${
                                view === 'list'
                                    ? 'bg-blue-500 text-white'
                                    : 'bg-gray-200'
                            }`}
                        >
                            List
                        </button>
                        <button
                            onClick={() => setView('grid')}
                            className={`px-3 py-1 rounded ${
                                view === 'grid'
                                    ? 'bg-blue-500 text-white'
                                    : 'bg-gray-200'
                            }`}
                        >
                            Grid
                        </button>
                    </div>
                </div>

                {loading ? (
                    <div className="flex justify-center items-center h-64">
                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
                    </div>
                ) : view === 'list' ? (
                    <DeviceList
                        devices={devices}
                        onCommand={executeCommand}
                    />
                ) : (
                    <DeviceStatusGrid
                        devices={devices}
                        onCommand={executeCommand}
                    />
                )}
            </div>
        </ErrorBoundary>
    )
}
