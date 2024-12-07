"use client"

import { useState, useEffect } from 'react'
import { fleetApi, connectWebSocket } from '@/services/api'
import { FleetMetrics } from '@/types/fleet'
import { WSMessage } from '@/types/ws'
import { useAuth } from '@/components/auth/AuthProvider'
import { DeviceStatusGrid } from '@/components/device/DeviceStatusGrid'
import { PerformanceMetrics } from '@/components/analytics/PerformanceMetrics'
import { PowerUsageChart } from '@/components/analytics/PowerUsageChart'
import { TemperatureHeatmap } from '@/components/analytics/TemperatureHeatmap'
import { ErrorBoundary } from '@/components/error/ErrorBoundary'

export default function DashboardPage() {
    const [metrics, setMetrics] = useState<FleetMetrics | null>(null)
    const [devices, setDevices] = useState<Device[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const { isAuthenticated } = useAuth()

    useEffect(() => {
        if (!isAuthenticated) return

        // Load initial metrics
        loadMetrics()
        const metricsInterval = setInterval(loadMetrics, 30000) // Update every 30s

        // Set up WebSocket connection
        const setupWebSocket = async () => {
            try {
                const ws = await connectWebSocket()
                ws.subscribe(handleWebSocketMessage)
                return ws.close
            } catch (err) {
                console.error('WebSocket connection failed:', err)
            }
        }

        const cleanup = setupWebSocket()
        return () => {
            clearInterval(metricsInterval)
            cleanup?.()
        }
    }, [isAuthenticated])

    const loadMetrics = async () => {
        try {
            const data = await fleetApi.getMetrics()
            setMetrics(data)
            setError(null)
        } catch (err) {
            setError('Failed to load fleet metrics')
            console.error(err)
        } finally {
            setLoading(false)
        }
    }

    const handleWebSocketMessage = (msg: WSMessage) => {
        switch (msg.type) {
            case 'metrics_update':
                // Update individual device metrics
                const update = msg.payload
                setDevices(current => 
                    current.map(device => 
                        device.id === update.deviceId
                            ? { ...device, metrics: update.metrics }
                            : device
                    )
                )
                break

            case 'state_update':
                // Update device states
                const stateUpdate = msg.payload
                setDevices(current =>
                    current.map(device =>
                        device.id === stateUpdate.deviceId
                            ? { ...device, ...stateUpdate.state }
                            : device
                    )
                )
                break
        }
    }

    if (error) {
        return (
            <div className="p-4">
                <div className="bg-red-50 border border-red-400 text-red-700 px-4 py-3 rounded">
                    <strong>Error:</strong> {error}
                    <button 
                        onClick={loadMetrics}
                        className="ml-4 text-sm underline"
                    >
                        Try Again
                    </button>
                </div>
            </div>
        )
    }

    if (loading || !metrics) {
        return (
            <div className="flex justify-center items-center h-64">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
            </div>
        )
    }

    return (
        <ErrorBoundary>
            <div className="p-4 space-y-6">
                {/* Overview stats */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                    <div className="bg-white rounded-lg shadow p-4">
                        <h3 className="text-sm font-medium text-gray-500">Total Devices</h3>
                        <p className="mt-1 text-3xl font-semibold text-gray-900">
                            {metrics.totalDevices}
                        </p>
                    </div>
                    <div className="bg-white rounded-lg shadow p-4">
                        <h3 className="text-sm font-medium text-gray-500">Active Devices</h3>
                        <p className="mt-1 text-3xl font-semibold text-green-600">
                            {metrics.activeDevices}
                        </p>
                    </div>
                    <div className="bg-white rounded-lg shadow p-4">
                        <h3 className="text-sm font-medium text-gray-500">Total Power</h3>
                        <p className="mt-1 text-3xl font-semibold text-blue-600">
                            {metrics.totalPower.toFixed(1)} W
                        </p>
                    </div>
                    <div className="bg-white rounded-lg shadow p-4">
                        <h3 className="text-sm font-medium text-gray-500">Avg Temperature</h3>
                        <p className="mt-1 text-3xl font-semibold text-orange-600">
                            {metrics.avgTemp.toFixed(1)}°C
                        </p>
                    </div>
                </div>

                {/* Performance metrics */}
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                    <div className="bg-white rounded-lg shadow p-4">
                        <h2 className="text-lg font-semibold mb-4">Performance</h2>
                        <PerformanceMetrics
                            cpuLoad={metrics.avgCpu}
                            memoryUsage={metrics.avgMemory}
                            resourceUsage={metrics.resourceUsage}
                        />
                    </div>

                    <div className="bg-white rounded-lg shadow p-4">
                        <h2 className="text-lg font-semibold mb-4">Power Usage</h2>
                        <PowerUsageChart
                            powerUsage={metrics.totalPower}
                            devices={devices}
                        />
                    </div>
                </div>

                {/* Temperature heatmap */}
                <div className="bg-white rounded-lg shadow p-4">
                    <h2 className="text-lg font-semibold mb-4">Temperature Distribution</h2>
                    <TemperatureHeatmap devices={devices} />
                </div>

                {/* Device status grid */}
                <div className="bg-white rounded-lg shadow p-4">
                    <h2 className="text-lg font-semibold mb-4">Device Status</h2>
                    <DeviceStatusGrid devices={devices} />
                </div>
            </div>
        </ErrorBoundary>
    )
}
