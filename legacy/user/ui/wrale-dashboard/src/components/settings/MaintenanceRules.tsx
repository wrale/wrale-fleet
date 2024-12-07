'use client'

import { useState } from 'react'

interface MaintenanceRule {
  id: string
  name: string
  description: string
  conditions: {
    metric: string
    operator: string
    value: number
    unit: string
  }[]
  priority: 'low' | 'medium' | 'high'
  recommendedAction: string
  enabled: boolean
}

export function MaintenanceRules() {
  const [rules] = useState<MaintenanceRule[]>([
    {
      id: '1',
      name: 'Power Supply Health Check',
      description: 'Monitor power supply metrics for potential failures',
      conditions: [
        {
          metric: 'voltage_fluctuation',
          operator: '>',
          value: 0.5,
          unit: 'V'
        },
        {
          metric: 'power_efficiency',
          operator: '<',
          value: 85,
          unit: '%'
        }
      ],
      priority: 'high',
      recommendedAction: 'Inspect and potentially replace power supply unit',
      enabled: true
    },
    {
      id: '2',
      name: 'Storage Performance Check',
      description: 'Monitor storage read/write speeds and errors',
      conditions: [
        {
          metric: 'write_speed',
          operator: '<',
          value: 20,
          unit: 'MB/s'
        },
        {
          metric: 'io_errors',
          operator: '>',
          value: 10,
          unit: 'errors/hour'
        }
      ],
      priority: 'medium',
      recommendedAction: 'Check SD card health and consider replacement',
      enabled: true
    }
  ])

  const getPriorityColor = (priority: MaintenanceRule['priority']) => {
    switch (priority) {
      case 'high':
        return 'text-wrale-danger'
      case 'medium':
        return 'text-wrale-warning'
      case 'low':
        return 'text-wrale-success'
    }
  }

  return (
    <div className="space-y-6">
      <section>
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-medium text-gray-900">Maintenance Rules</h2>
          <button className="bg-wrale-primary text-white px-4 py-2 rounded-lg hover:bg-wrale-primary/90 text-sm">
            Add Rule
          </button>
        </div>

        <div className="space-y-4">
          {rules.map((rule) => (
            <div key={rule.id} className="bg-white shadow rounded-lg p-6">
              <div className="flex items-start justify-between mb-4">
                <div>
                  <h3 className="text-lg font-medium">{rule.name}</h3>
                  <p className="text-sm text-gray-500 mt-1">{rule.description}</p>
                </div>
                <div className="flex items-center space-x-2">
                  <span className={`text-sm font-medium ${getPriorityColor(rule.priority)}`}>
                    {rule.priority.charAt(0).toUpperCase() + rule.priority.slice(1)} Priority
                  </span>
                  <label className="flex items-center">
                    <input
                      type="checkbox"
                      className="form-checkbox h-4 w-4 text-wrale-primary"
                      checked={rule.enabled}
                      onChange={() => {}}
                    />
                  </label>
                </div>
              </div>

              <div className="mt-4">
                <h4 className="text-sm font-medium text-gray-700 mb-2">Conditions:</h4>
                <ul className="space-y-2">
                  {rule.conditions.map((condition, index) => (
                    <li key={index} className="flex items-center text-sm text-gray-600">
                      <span className="w-2 h-2 rounded-full bg-gray-400 mr-2"></span>
                      {condition.metric.replace('_', ' ')} {condition.operator} {condition.value}
                      {condition.unit}
                    </li>
                  ))}
                </ul>
              </div>

              <div className="mt-4">
                <h4 className="text-sm font-medium text-gray-700 mb-2">Recommended Action:</h4>
                <p className="text-sm text-gray-600">{rule.recommendedAction}</p>
              </div>

              <div className="mt-4 pt-4 border-t border-gray-200 flex justify-end space-x-3">
                <button className="text-wrale-primary hover:text-wrale-primary/80 text-sm font-medium">
                  Edit
                </button>
                <button className="text-wrale-danger hover:text-wrale-danger/80 text-sm font-medium">
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  )
}