import { useState } from 'react'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'

interface DeviceMetricsProps {
  id: string
}

const data = [
  { time: '00:00', cpu: 65, memory: 60, temperature: 45 },
  { time: '01:00', cpu: 68, memory: 62, temperature: 46 },
  { time: '02:00', cpu: 75, memory: 65, temperature: 48 },
  { time: '03:00', cpu: 70, memory: 63, temperature: 47 },
  { time: '04:00', cpu: 72, memory: 64, temperature: 47 },
  { time: '05:00', cpu: 80, memory: 70, temperature: 50 },
  { time: '06:00', cpu: 85, memory: 75, temperature: 52 },
  { time: '07:00', cpu: 78, memory: 68, temperature: 49 },
]

export function DeviceMetrics({ id }: DeviceMetricsProps) {
  const [timeRange, setTimeRange] = useState('24h')

  return (
    <div className="bg-white rounded-lg shadow overflow-hidden">
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold">Performance Metrics</h2>
          <select 
            value={timeRange}
            onChange={(e) => setTimeRange(e.target.value)}
            className="px-3 py-1 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-wrale-primary/50"
          >
            <option value="1h">Last Hour</option>
            <option value="6h">Last 6 Hours</option>
            <option value="24h">Last 24 Hours</option>
            <option value="7d">Last 7 Days</option>
          </select>
        </div>
      </div>
      <div className="p-6">
        <div className="h-96">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={data}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line 
                type="monotone" 
                dataKey="cpu" 
                stroke="#f56565" 
                name="CPU Usage (%)" 
                dot={false}
              />
              <Line 
                type="monotone" 
                dataKey="memory" 
                stroke="#4fd1c5" 
                name="Memory Usage (%)" 
                dot={false}
              />
              <Line 
                type="monotone" 
                dataKey="temperature" 
                stroke="#ed8936" 
                name="Temperature (°C)" 
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  )
}