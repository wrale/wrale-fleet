import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'

const data = [
  { time: '00:00', totalPower: 2100, rackOne: 1200, rackTwo: 900 },
  { time: '04:00', totalPower: 2300, rackOne: 1300, rackTwo: 1000 },
  { time: '08:00', totalPower: 2600, rackOne: 1500, rackTwo: 1100 },
  { time: '12:00', totalPower: 2800, rackOne: 1600, rackTwo: 1200 },
  { time: '16:00', totalPower: 2500, rackOne: 1400, rackTwo: 1100 },
  { time: '20:00', totalPower: 2200, rackOne: 1300, rackTwo: 900 },
]

export function PowerUsageChart() {
  return (
    <div className="h-80">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="time" />
          <YAxis />
          <Tooltip />
          <Legend />
          <Line 
            type="monotone" 
            dataKey="totalPower" 
            name="Total Power (W)"
            stroke="#4fd1c5" 
            strokeWidth={2}
          />
          <Line 
            type="monotone" 
            dataKey="rackOne" 
            name="Rack 1 (W)"
            stroke="#9f7aea" 
            strokeDasharray="5 5"
          />
          <Line 
            type="monotone" 
            dataKey="rackTwo" 
            name="Rack 2 (W)"
            stroke="#ed64a6" 
            strokeDasharray="5 5"
          />
        </LineChart>
      </ResponsiveContainer>

      <div className="mt-4 grid grid-cols-2 gap-4">
        <div>
          <h3 className="text-sm font-medium text-gray-500">Daily Average</h3>
          <p className="mt-1 text-2xl font-semibold">2.4 kW</p>
          <p className="text-sm text-wrale-success">↓ 12% vs last week</p>
        </div>
        <div>
          <h3 className="text-sm font-medium text-gray-500">Peak Usage</h3>
          <p className="mt-1 text-2xl font-semibold">2.8 kW</p>
          <p className="text-sm text-wrale-warning">↑ 5% vs last week</p>
        </div>
      </div>
    </div>
  )
}