import { ClockIcon, ExclamationTriangleIcon, WrenchIcon } from '@heroicons/react/24/outline'

interface MaintenanceAlert {
  id: string
  deviceName: string
  issue: string
  priority: 'high' | 'medium' | 'low'
  predictedTime: string
  indicators: string[]
}

export function MaintenancePredictor() {
  // TODO: Replace with real API call
  const alerts: MaintenanceAlert[] = [
    {
      id: '1',
      deviceName: 'pi-cluster-01',
      issue: 'Potential Power Supply Failure',
      priority: 'high',
      predictedTime: '48 hours',
      indicators: [
        'Voltage fluctuations',
        'Increased temperature',
        'Power draw anomalies'
      ]
    },
    {
      id: '2',
      deviceName: 'pi-cluster-02',
      issue: 'SD Card Degradation',
      priority: 'medium',
      predictedTime: '5 days',
      indicators: [
        'Increased I/O errors',
        'Slower write speeds',
        'Filesystem warnings'
      ]
    }
  ]

  const getPriorityColor = (priority: MaintenanceAlert['priority']) => {
    switch (priority) {
      case 'high':
        return 'text-wrale-danger'
      case 'medium':
        return 'text-wrale-warning'
      case 'low':
        return 'text-wrale-success'
    }
  }

  const getPriorityBg = (priority: MaintenanceAlert['priority']) => {
    switch (priority) {
      case 'high':
        return 'bg-wrale-danger/10'
      case 'medium':
        return 'bg-wrale-warning/10'
      case 'low':
        return 'bg-wrale-success/10'
    }
  }

  return (
    <div>
      <div className="space-y-4">
        {alerts.map(alert => (
          <div 
            key={alert.id} 
            className={`rounded-lg p-4 ${getPriorityBg(alert.priority)}`}
          >
            <div className="flex items-start justify-between">
              <div>
                <h3 className="font-medium">{alert.deviceName}</h3>
                <p className={`text-sm mt-1 ${getPriorityColor(alert.priority)}`}>
                  {alert.issue}
                </p>
              </div>
              <ExclamationTriangleIcon 
                className={`w-5 h-5 ${getPriorityColor(alert.priority)}`} 
              />
            </div>

            <div className="mt-3 flex items-center text-sm text-gray-600">
              <ClockIcon className="w-4 h-4 mr-1" />
              <span>Predicted in: {alert.predictedTime}</span>
            </div>

            <div className="mt-3">
              <h4 className="text-sm font-medium mb-2">Indicators:</h4>
              <ul className="text-sm space-y-1">
                {alert.indicators.map((indicator, index) => (
                  <li key={index} className="flex items-center text-gray-600">
                    <span className="w-1.5 h-1.5 rounded-full bg-gray-400 mr-2" />
                    {indicator}
                  </li>
                ))}
              </ul>
            </div>

            <div className="mt-4">
              <button className="flex items-center text-wrale-primary hover:text-wrale-primary/80 text-sm font-medium">
                <WrenchIcon className="w-4 h-4 mr-1" />
                Schedule Maintenance
              </button>
            </div>
          </div>
        ))}
      </div>

      <div className="mt-6">
        <h3 className="text-sm font-medium text-gray-500 mb-2">System Health Score</h3>
        <div className="flex items-center">
          <div className="flex-1 bg-gray-200 rounded-full h-2 mr-3">
            <div 
              className="bg-wrale-warning rounded-full h-2" 
              style={{ width: '75%' }}
            />
          </div>
          <span className="text-sm font-medium">75%</span>
        </div>
        <p className="mt-2 text-sm text-gray-500">
          Based on current maintenance predictions and system performance
        </p>
      </div>
    </div>
  )
}