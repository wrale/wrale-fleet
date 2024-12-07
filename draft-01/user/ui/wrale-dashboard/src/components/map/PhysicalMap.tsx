import { useState } from 'react'
import Link from 'next/link'
import type { Device } from '@/types/device'

const GRID_SIZE = 20
const CELL_SIZE = 30

interface GridCell {
  x: number
  y: number
  type: 'empty' | 'rack' | 'device'
  id?: string
  name?: string
  status?: Device['status']
}

export function PhysicalMap() {
  const [grid] = useState<GridCell[][]>(
    Array.from({ length: GRID_SIZE }, (_, y) =>
      Array.from({ length: GRID_SIZE }, (_, x) => ({
        x,
        y,
        type: 'empty'
      }))
    )
  )

  // Add example racks and devices
  grid[5][5] = { x: 5, y: 5, type: 'rack', id: 'rack1', name: 'Rack 1' }
  grid[5][6] = { x: 5, y: 6, type: 'device', id: 'device1', name: 'pi-cluster-01', status: 'online' }
  grid[5][7] = { x: 5, y: 7, type: 'device', id: 'device2', name: 'pi-cluster-02', status: 'warning' }
  
  grid[10][5] = { x: 10, y: 5, type: 'rack', id: 'rack2', name: 'Rack 2' }
  grid[10][6] = { x: 10, y: 6, type: 'device', id: 'device3', name: 'pi-cluster-03', status: 'offline' }

  const getCellColor = (cell: GridCell) => {
    if (cell.type === 'rack') return 'bg-gray-600'
    if (cell.type === 'device') {
      switch (cell.status) {
        case 'online':
          return 'bg-wrale-success'
        case 'warning':
          return 'bg-wrale-warning'
        case 'offline':
          return 'bg-wrale-danger'
        default:
          return 'bg-gray-400'
      }
    }
    return 'bg-gray-100'
  }

  return (
    <div className="overflow-auto">
      <div 
        className="grid gap-px bg-gray-200 p-1"
        style={{
          gridTemplateColumns: `repeat(${GRID_SIZE}, ${CELL_SIZE}px)`,
          width: `${GRID_SIZE * CELL_SIZE}px`
        }}
      >
        {grid.flat().map((cell) => (
          <div
            key={`${cell.x}-${cell.y}`}
            className={`relative ${getCellColor(cell)} transition-colors`}
            style={{ height: CELL_SIZE }}
          >
            {cell.type !== 'empty' && (
              <div className="absolute inset-0 flex items-center justify-center">
                <Link 
                  href={cell.type === 'device' ? `/devices/${cell.id}` : '#'}
                  className="w-full h-full flex items-center justify-center hover:bg-black/10"
                >
                  <span className="sr-only">{cell.name}</span>
                </Link>
              </div>
            )}
          </div>
        ))}
      </div>

      <div className="mt-4">
        <div className="text-sm text-gray-500">Legend:</div>
        <div className="flex space-x-4 mt-2">
          <div className="flex items-center">
            <div className="w-4 h-4 bg-gray-600 mr-2"></div>
            <span className="text-sm">Rack</span>
          </div>
          <div className="flex items-center">
            <div className="w-4 h-4 bg-wrale-success mr-2"></div>
            <span className="text-sm">Online Device</span>
          </div>
          <div className="flex items-center">
            <div className="w-4 h-4 bg-wrale-warning mr-2"></div>
            <span className="text-sm">Warning</span>
          </div>
          <div className="flex items-center">
            <div className="w-4 h-4 bg-wrale-danger mr-2"></div>
            <span className="text-sm">Offline Device</span>
          </div>
        </div>
      </div>
    </div>
  )
}