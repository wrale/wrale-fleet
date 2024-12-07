import { useState } from 'react'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'

const data = [
  {
    date: '2024-01-01',
    avgCpuLoad: 65,
    avgMemoryUsage: 72,
    avgTemperature: 42,
    networkLatency: 15
  },
  {
    date: '2024-01-02',
    avgCpuLoad: 68,
    avgMemoryUsage: 75,
    avgTemperature: 43,
    networkLatency: 18
  },
  {
    date: '2024-01-03',
    avgCpuLoad: 72,
    avgMemoryUsage: 78,
    avgTemperature: 44,
    networkLatency: 20
  },
  {
    date: '2024-01-04',
    avgCpuLoad: 70,
    avgMemoryUsage: 76,
    avgTemperature: 43,
    networkLatency: 16
  },
  {
    date: '2024-01-05',
    avgCpuLoad: 75,
    avgMemoryUsage: 80,
    avgTemperature: 45,
    networkLatency: 22
  }
]

type MetricType = 'cpu' | 'memory' | 'temperature' | 'network'

export function PerformanceMetrics() {
  const [selectedMetrics, setSelectedMetrics] = useState<MetricType[]>(['cpu', 'memory'])

  const metrics = [
    { id: 'cpu', name: 'CPU Load', color: '#f56565', dataKey: 'avgCpuLoad' },
    { id: 'memory', name: 'Memory Usage', color: '#4fd1c5', dataKey: 'avgMemoryUsage' },
    { id: 'temperature', name: 'Temperature', color: '#ed8936', dataKey: 'avgTemperature' },
    { id: 'network', name: 'Network Latency', color: '#9f7aea', dataKey: 'networkLatency' }
  ]

  const toggleMetric = (metricId: MetricType) => {
    if (selectedMetrics.includes(metricId)) {
      setSelectedMetrics(selectedMetrics.filter(id => id !== metricId))
    } else {
      setSelectedMetrics([...selectedMetrics, metricId])
    }
  }

  const getMetricTrend = (metricId: MetricType) => {
    const metric = metrics.find(m => m.id === metricId)
    if (!metric) return { value: 0, trend: 0 }

    const values = data.map(d => d[metric.dataKey as keyof typeof data[0]] as number)
    const current = values[values.length - 1]
    const previous = values[values.length - 2]
    const trend = ((current - previous) / previous) * 100

    return {
      value: current,
      trend: parseFloat(trend.toFixed(1))
    }
  }

  return (
    <div>
      <div className="mb-4 flex flex-wrap gap-2">
        {metrics.map(metric => (
          <button
            key={metric.id}
            onClick={() => toggleMetric(metric.id as MetricType)}
            className={`px-3 py-1 rounded-full text-sm font-medium transition-colors ${
              selectedMetrics.includes(metric.id as MetricType)
                ? 'bg-gray-800 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            }`}
          >
            {metric.name}
          </button>
        ))}
      </div>

      <div className="h-80">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={data}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="date" />
            <YAxis />
            <Tooltip />
            <Legend />
            {metrics
              .filter(metric => selectedMetrics.includes(metric.id as MetricType))
              .map(metric => (
                <Line
                  key={metric.id}
                  type="monotone"
                  dataKey={metric.dataKey}
                  name={metric.name}
                  stroke={metric.color}
                  strokeWidth={2}
                />
              ))}
          </LineChart>
        </ResponsiveContainer>
      </div>

      <div className="mt-6 grid grid-cols-2 lg:grid-cols-4 gap-4">
        {metrics.map(metric => {
          const { value, trend } = getMetricTrend(metric.id as MetricType)
          return (
            <div key={metric.id} className="bg-gray-50 rounded-lg p-4">
              <h3 className="text-sm font-medium text-gray-500">{metric.name}</h3>
              <p className="mt-1 text-2xl font-semibold">{value}</p>
              <p className={`text-sm ${trend > 0 ? 'text-wrale-warning' : 'text-wrale-success'}`}>
                {trend > 0 ? '↑' : '↓'} {Math.abs(trend)}% vs previous
              </p>
            </div>
          )
        })}
      </div>
    </div>
  )
}