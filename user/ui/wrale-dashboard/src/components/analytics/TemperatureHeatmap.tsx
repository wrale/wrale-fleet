interface RackTemperature {
  id: string
  unit: number
  temperature: number
}

export function TemperatureHeatmap() {
  const rack1Data: RackTemperature[] = [
    { id: '1', unit: 1, temperature: 35 },
    { id: '2', unit: 2, temperature: 38 },
    { id: '3', unit: 3, temperature: 42 },
    { id: '4', unit: 4, temperature: 45 },
    { id: '5', unit: 5, temperature: 40 },
    { id: '6', unit: 6, temperature: 37 },
  ]

  const rack2Data: RackTemperature[] = [
    { id: '7', unit: 1, temperature: 36 },
    { id: '8', unit: 2, temperature: 39 },
    { id: '9', unit: 3, temperature: 41 },
    { id: '10', unit: 4, temperature: 43 },
    { id: '11', unit: 5, temperature: 38 },
    { id: '12', unit: 6, temperature: 35 },
  ]

  const getTemperatureColor = (temp: number) => {
    if (temp >= 45) return 'bg-red-500'
    if (temp >= 40) return 'bg-orange-400'
    if (temp >= 35) return 'bg-yellow-300'
    return 'bg-green-300'
  }

  const getTemperatureOpacity = (temp: number) => {
    const percentage = ((temp - 30) / 20) * 100
    return `${Math.min(100, Math.max(20, percentage))}%`
  }

  const RackColumn = ({ data, name }: { data: RackTemperature[], name: string }) => (
    <div>
      <h3 className="text-sm font-medium mb-2">{name}</h3>
      <div className="space-y-1">
        {[...data].reverse().map((item) => (
          <div
            key={item.id}
            className={`h-8 rounded ${getTemperatureColor(item.temperature)}`}
            style={{ opacity: getTemperatureOpacity(item.temperature) }}
          >
            <div className="flex items-center justify-between px-2 py-1">
              <span className="text-xs font-medium">U{item.unit}</span>
              <span className="text-xs font-medium">{item.temperature}°C</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )

  return (
    <div>
      <div className="grid grid-cols-2 gap-8">
        <RackColumn data={rack1Data} name="Rack 1" />
        <RackColumn data={rack2Data} name="Rack 2" />
      </div>

      <div className="mt-6">
        <div className="text-sm text-gray-500 mb-2">Temperature Scale:</div>
        <div className="flex items-center space-x-4">
          <div className="flex items-center">
            <div className="w-4 h-4 bg-green-300 rounded mr-1"></div>
            <span className="text-sm">&lt; 35°C</span>
          </div>
          <div className="flex items-center">
            <div className="w-4 h-4 bg-yellow-300 rounded mr-1"></div>
            <span className="text-sm">35-40°C</span>
          </div>
          <div className="flex items-center">
            <div className="w-4 h-4 bg-orange-400 rounded mr-1"></div>
            <span className="text-sm">40-45°C</span>
          </div>
          <div className="flex items-center">
            <div className="w-4 h-4 bg-red-500 rounded mr-1"></div>
            <span className="text-sm">&gt; 45°C</span>
          </div>
        </div>
      </div>
    </div>
  )
}