"use client"

import { useEffect, useState } from 'react'
import {
    LineChart,
    Line,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer
} from 'recharts'

interface PerformanceMetricsProps {
    cpuLoad: number
    memoryUsage: number
    resourceUsage?: {
        cpu: number
        memory: number
        power: number
    }
}

interface MetricsPoint {
    timestamp: number
    cpu: number
    memory: number
}

const MAX_POINTS = 20

export function PerformanceMetrics({
    cpuLoad,
    memoryUsage,
    resourceUsage
}: PerformanceMetricsProps) {
    const [history, setHistory] = useState<MetricsPoint[]>([])

    useEffect(() => {
        // Add new data point
        const timestamp = Date.now()
        setHistory(current => {
            const updated = [
                ...current,
                { timestamp, cpu: cpuLoad, memory: memoryUsage }
            ]
            // Keep only last N points
            return updated.slice(-MAX_POINTS)
        })
    }, [cpuLoad, memoryUsage])

    return (
        <div className="space-y-4">
            {/* Current values */}
            <div className="grid grid-cols-2 gap-4">
                <div>
                    <h3 className="text-sm font-medium text-gray-500">CPU Load</h3>
                    <div className="mt-1">
                        <div className="relative pt-1">
                            <div className="flex mb-2 items-center justify-between">
                                <div>
                                    <span className="text-xs font-semibold inline-block py-1 px-2 uppercase rounded-full text-blue-600 bg-blue-200">
                                        {cpuLoad.toFixed(1)}%
                                    </span>
                                </div>
                            </div>
                            <div className="overflow-hidden h-2 mb-4 text-xs flex rounded bg-blue-200">
                                <div
                                    style={{ width: `${cpuLoad}%` }}
                                    className="shadow-none flex flex-col text-center whitespace-nowrap text-white justify-center bg-blue-500"
                                />
                            </div>
                        </div>
                    </div>
                </div>

                <div>
                    <h3 className="text-sm font-medium text-gray-500">Memory Usage</h3>
                    <div className="mt-1">
                        <div className="relative pt-1">
                            <div className="flex mb-2 items-center justify-between">
                                <div>
                                    <span className="text-xs font-semibold inline-block py-1 px-2 uppercase rounded-full text-green-600 bg-green-200">
                                        {memoryUsage.toFixed(1)}%
                                    </span>
                                </div>
                            </div>
                            <div className="overflow-hidden h-2 mb-4 text-xs flex rounded bg-green-200">
                                <div
                                    style={{ width: `${memoryUsage}%` }}
                                    className="shadow-none flex flex-col text-center whitespace-nowrap text-white justify-center bg-green-500"
                                />
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* Historical chart */}
            <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={history}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis
                            dataKey="timestamp"
                            type="number"
                            domain={['auto', 'auto']}
                            tickFormatter={(ts) => new Date(ts).toLocaleTimeString()}
                        />
                        <YAxis domain={[0, 100]} />
                        <Tooltip
                            labelFormatter={(ts) => new Date(ts).toLocaleTimeString()}
                            formatter={(value: number) => [`${value.toFixed(1)}%`]}
                        />
                        <Line
                            type="monotone"
                            dataKey="cpu"
                            stroke="#3B82F6"
                            name="CPU"
                            dot={false}
                        />
                        <Line
                            type="monotone"
                            dataKey="memory"
                            stroke="#10B981"
                            name="Memory"
                            dot={false}
                        />
                    </LineChart>
                </ResponsiveContainer>
            </div>

            {/* Resource usage comparison */}
            {resourceUsage && (
                <div className="grid grid-cols-3 gap-4 mt-4">
                    <div className="text-center">
                        <div className="text-sm font-medium text-gray-500">CPU Efficiency</div>
                        <div className="mt-1 text-lg font-semibold">
                            {(resourceUsage.cpu * 100).toFixed(1)}%
                        </div>
                    </div>
                    <div className="text-center">
                        <div className="text-sm font-medium text-gray-500">Memory Efficiency</div>
                        <div className="mt-1 text-lg font-semibold">
                            {(resourceUsage.memory * 100).toFixed(1)}%
                        </div>
                    </div>
                    <div className="text-center">
                        <div className="text-sm font-medium text-gray-500">Power Efficiency</div>
                        <div className="mt-1 text-lg font-semibold">
                            {(resourceUsage.power * 100).toFixed(1)}%
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}
