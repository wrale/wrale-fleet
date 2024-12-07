"use client"

import { useEffect, useState } from 'react'
import {
    BarChart,
    Bar,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer,
    Cell
} from 'recharts'
import { Device } from '@/types/device'

interface PowerUsageChartProps {
    powerUsage: number
    devices: Device[]
}

interface PowerDataPoint {
    name: string
    value: number
    status: string
}

const STATUS_COLORS = {
    active: '#10B981',    // Green
    standby: '#F59E0B',   // Yellow
    error: '#EF4444',     // Red
    default: '#6B7280'    // Gray
}

export function PowerUsageChart({
    powerUsage,
    devices
}: PowerUsageChartProps) {
    const [data, setData] = useState<PowerDataPoint[]>([])

    useEffect(() => {
        // Transform devices into power usage data
        const powerData = devices
            .map(device => ({
                name: device.id,
                value: device.metrics.powerUsage,
                status: device.status
            }))
            .sort((a, b) => b.value - a.value) // Sort by power usage descending

        setData(powerData)
    }, [devices])

    const getStatusColor = (status: string) => {
        return STATUS_COLORS[status as keyof typeof STATUS_COLORS] || STATUS_COLORS.default
    }

    return (
        <div className="space-y-4">
            {/* Total power usage */}
            <div className="text-center">
                <div className="text-sm font-medium text-gray-500">Total Power Usage</div>
                <div className="mt-1 text-3xl font-semibold text-blue-600">
                    {powerUsage.toFixed(1)} W
                </div>
            </div>

            {/* Power usage by device */}
            <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={data}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis
                            dataKey="name"
                            angle={-45}
                            textAnchor="end"
                            height={60}
                            interval={0}
                            tick={{ fontSize: 12 }}
                        />
                        <YAxis
                            label={{ 
                                value: 'Power Usage (W)',
                                angle: -90,
                                position: 'insideLeft'
                            }}
                        />
                        <Tooltip
                            formatter={(value: number) => [`${value.toFixed(1)}W`]}
                            labelFormatter={(label: string) => `Device: ${label}`}
                        />
                        <Bar dataKey="value">
                            {data.map((entry, index) => (
                                <Cell 
                                    key={`cell-${index}`}
                                    fill={getStatusColor(entry.status)}
                                />
                            ))}
                        </Bar>
                    </BarChart>
                </ResponsiveContainer>
            </div>

            {/* Legend */}
            <div className="flex justify-center space-x-4">
                {Object.entries(STATUS_COLORS).map(([status, color]) => (
                    <div key={status} className="flex items-center">
                        <div 
                            className="w-3 h-3 rounded-full mr-1"
                            style={{ backgroundColor: color }}
                        />
                        <span className="text-sm capitalize">{status}</span>
                    </div>
                ))}
            </div>
        </div>
    )
}
