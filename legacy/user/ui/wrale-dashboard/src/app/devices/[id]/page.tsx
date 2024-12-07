"use client"

import { useState, useEffect } from 'react'
import { useParams } from 'next/navigation'
import { deviceApi, connectWebSocket } from '@/services/api'
import { Device, DeviceCommand } from '@/types/device'
import { WSMessage } from '@/types/ws'
import { useAuth } from '@/components/auth/AuthProvider'
import { DeviceDetails } from '@/components/device/DeviceDetails'
import { DeviceMetrics } from '@/components/device/DeviceMetrics'
import { DeviceEnvironment } from '@/components/device/DeviceEnvironment'
import { ErrorBoundary } from '@/components/error/ErrorBoundary'

export default function DeviceDetailPage() {
    const { id } = useParams()
    const [device, setDevice] = useState<Device | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const { isAuthenticated } = useAuth()

    useEffect(() => {
        if (!isAuthenticated || !id) return

        // Load initial device data
        loadDevice()

        // Set up WebSocket connection
        const setupWebSocket = async () => {
            try {
                const ws = await connectWebSocket([id as string])
                ws.subscribe(handleWebSocketMessage)
                return ws.close // Return cleanup function
            } catch (err) {
                console.error('WebSocket connection failed:', err)
            }
        }

        const cleanup = setupWebSocket()
        return () => {
            cleanup?.()
        }
    }, [isAuthenticated, id])

    const loadDevice = async () => {
        if (!id) return

        try {
            setLoading(true)
            const data = await deviceApi.get(id as string)
            setDevice(data)
            setError(null)
        } catch (err) {
            setError('Failed to load device')
            console.error(err)
        } finally {
            setLoading(false)
        }
    }

    const handleWebSocketMessage = (msg: WSMessage) => {
        if (!device) return

        switch (msg.type) {
            case 'state_update':
                const stateUpdate = msg.payload
                if (stateUpdate.deviceId === device.id) {
                    setDevice(current => ({ ...current!, ...stateUpdate.state }))
                }
                break

            case 'metrics_update':
                const metricsUpdate = msg.payload
                if (metricsUpdate.deviceId === device.id) {
                    setDevice(current => ({
                        ...current!,
                        metrics: metricsUpdate.metrics
                    }))
                }
                break

            case 'device_removed':
                const { device_id } = msg.payload as { device_id: string }
                if (device_id === device.id) {
                    setError('Device has been removed')
                }
                break
        }
    }

    const executeCommand = async (command: DeviceCommand) => {
        if (!device) return

        try {
            await deviceApi.executeCommand(device.id, command)
            // Device state will be updated via WebSocket
        } catch (err) {
            console.error('Command execution failed:', err)
            throw err
        }
    }

    const updateConfig = async (config: Record<string, any>) => {
        if (!device) return

        try {
            await deviceApi.update(device.id, { config })
            // Device config will be updated via WebSocket
        } catch (err) {
            console.error('Config update failed:', err)
            throw err
        }
    }

    if (error) {
        return (
            <div className="p-4">
                <div className="bg-red-50 border border-red-400 text-red-700 px-4 py-3 rounded">
                    <strong>Error:</strong> {error}
                    <button 
                        onClick={loadDevice}
                        className="ml-4 text-sm underline"
                    >
                        Try Again
                    </button>
                </div>
            </div>
        )
    }

    if (loading) {
        return (
            <div className="flex justify-center items-center h-64">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
            </div>
        )
    }

    if (!device) {
        return (
            <div className="p-4">
                <div className="text-center text-gray-500">Device not found</div>
            </div>
        )
    }

    return (
        <ErrorBoundary>
            <div className="p-4 space-y-6">
                <DeviceDetails
                    device={device}
                    onCommand={executeCommand}
                    onConfigUpdate={updateConfig}
                />

                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                    <div className="bg-white rounded-lg shadow p-4">
                        <h2 className="text-lg font-semibold mb-4">Metrics</h2>
                        <DeviceMetrics metrics={device.metrics} />
                    </div>

                    <div className="bg-white rounded-lg shadow p-4">
                        <h2 className="text-lg font-semibold mb-4">Environment</h2>
                        <DeviceEnvironment
                            metrics={device.metrics}
                            location={device.location}
                        />
                    </div>
                </div>
            </div>
        </ErrorBoundary>
    )
}
