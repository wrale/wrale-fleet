"use client"

import { useEffect, useState } from 'react'
import { Device } from '@/types/device'

interface TemperatureHeatmapProps {
    devices: Device[]
}

interface GridCell {
    deviceId: string | null
    temperature: number | null
    status: string | null
}

const GRID_SIZE = 10 // 10x10 grid for demonstration
const TEMPERATURE_RANGE = {
    min: 20,
    max: 80
}

export function TemperatureHeatmap({ devices }: TemperatureHeatmapProps) {
    const [grid, setGrid] = useState<GridCell[][]>([])

    useEffect(() => {
        // Create empty grid
        const emptyGrid = Array(GRID_SIZE).fill(null).map(() =>
            Array(GRID_SIZE).fill(null).map(() => ({
                deviceId: null,
                temperature: null,
                status: null
            }))
        )

        // Place devices on grid based on their position
        devices.forEach(device => {
            const row = Math.floor(device.location.position / GRID_SIZE)
            const col = device.location.position % GRID_SIZE
            if (row < GRID_SIZE && col < GRID_SIZE) {
                emptyGrid[row][col] = {
                    deviceId: device.id,
                    temperature: device.metrics.temperature,
                    status: device.status
                }
            }
        })

        setGrid(emptyGrid)
    }, [devices])

    const getTemperatureColor = (temp: number | null) => {
        if (temp === null) return '#f3f4f6' // Gray for empty cells

        // Calculate color based on temperature range
        const percentage = Math.min(
            Math.max(
                (temp - TEMPERATURE_RANGE.min) /
                (TEMPERATURE_RANGE.max - TEMPERATURE_RANGE.min),
                0
            ),
            1
        )

        // Color gradient from blue (cool) to red (hot)
        const blue = Math.round(255 * (1 - percentage))
        const red = Math.round(255 * percentage)
        return `rgb(${red}, 0, ${blue})`
    }

    return (
        <div className="space-y-4">
            {/* Temperature grid */}
            <div className="grid gap-1" 
                style={{ 
                    gridTemplateColumns: `repeat(${GRID_SIZE}, minmax(0, 1fr))`
                }}>
                {grid.flat().map((cell, index) => (
                    <div
                        key={index}
                        className="aspect-square relative rounded-sm"
                        style={{ 
                            backgroundColor: getTemperatureColor(cell.temperature)
                        }}
                    >
                        {cell.temperature !== null && (
                            <div className="absolute inset-0 flex items-center justify-center text-xs font-medium text-white">
                                {cell.temperature.toFixed(1)}°C
                            </div>
                        )}
                    </div>
                ))}
            </div>

            {/* Legend */}
            <div className="flex items-center justify-center space-x-2">
                <div className="text-sm">Cool</div>
                <div className="h-2 w-32 bg-gradient-to-r from-blue-500 to-red-500 rounded" />
                <div className="text-sm">Hot</div>
            </div>

            {/* Temperature range */}
            <div className="flex justify-between text-sm text-gray-500">
                <span>{TEMPERATURE_RANGE.min}°C</span>
                <span>{TEMPERATURE_RANGE.max}°C</span>
            </div>
        </div>
    )
}
