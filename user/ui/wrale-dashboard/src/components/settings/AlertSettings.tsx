'use client'

import { useState } from 'react'

interface AlertChannel {
  id: string
  type: 'email' | 'slack' | 'webhook'
  name: string
  config: {
    recipients?: string[]
    webhook_url?: string
    channel?: string
  }
  enabled: boolean
}

interface AlertRule {
  id: string
  name: string
  condition: string
  severity: 'low' | 'medium' | 'high'
  channels: string[]
  enabled: boolean
}

export function AlertSettings() {
  const [channels] = useState<AlertChannel[]>([
    {
      id: '1',
      type: 'email',
      name: 'Operations Team',
      config: {
        recipients: ['ops@example.com', 'alerts@example.com']
      },
      enabled: true
    },
    {
      id: '2',
      type: 'slack',
      name: 'Slack #alerts',
      config: {
        channel: '#alerts'
      },
      enabled: true
    }
  ])

  const [rules] = useState<AlertRule[]>([
    {
      id: '1',
      name: 'High Temperature Alert',
      condition: 'temperature > 50°C for 5 minutes',
      severity: 'high',
      channels: ['1', '2'],
      enabled: true
    },
    {
      id: '2',
      name: 'Power Usage Warning',
      condition: 'power_usage > 2500W for 10 minutes',
      severity: 'medium',
      channels: ['1'],
      enabled: true
    }
  ])

  const getSeverityColor = (severity: AlertRule['severity']) => {
    switch (severity) {
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
        <h2 className="text-lg font-medium text-gray-900 mb-4">Alert Channels</h2>
        <div className="bg-white shadow rounded-lg">
          <div className="divide-y divide-gray-200">
            {channels.map((channel) => (
              <div key={channel.id} className="p-6">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center">
                    <h3 className="text-lg font-medium">{channel.name}</h3>
                    <span className={`ml-3 text-sm ${channel.enabled ? 'text-wrale-success' : 'text-gray-500'}`}>
                      {channel.enabled ? 'Active' : 'Inactive'}
                    </span>
                  </div>
                  <div className="space-x-3">
                    <button className="text-wrale-primary hover:text-wrale-primary/80 text-sm font-medium">
                      Edit
                    </button>
                    <button className="text-wrale-danger hover:text-wrale-danger/80 text-sm font-medium">
                      Delete
                    </button>
                  </div>
                </div>

                <div className="text-sm text-gray-500">
                  {channel.type === 'email' && (
                    <div>Recipients: {channel.config.recipients?.join(', ')}</div>
                  )}
                  {channel.type === 'slack' && (
                    <div>Channel: {channel.config.channel}</div>
                  )}
                  {channel.type === 'webhook' && (
                    <div>Webhook URL: {channel.config.webhook_url}</div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>

        <button className="mt-4 flex items-center text-sm text-wrale-primary hover:text-wrale-primary/80">
          <svg className="w-5 h-5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Add Channel
        </button>
      </section>

      <section>
        <h2 className="text-lg font-medium text-gray-900 mb-4">Alert Rules</h2>
        <div className="bg-white shadow rounded-lg">
          <div className="divide-y divide-gray-200">
            {rules.map((rule) => (
              <div key={rule.id} className="p-6">
                <div className="flex items-center justify-between mb-4">
                  <div>
                    <h3 className="text-lg font-medium">{rule.name}</h3>
                    <span className={`text-sm ${getSeverityColor(rule.severity)}`}>
                      {rule.severity.charAt(0).toUpperCase() + rule.severity.slice(1)} Priority
                    </span>
                  </div>
                  <div className="space-x-3">
                    <button className="text-wrale-primary hover:text-wrale-primary/80 text-sm font-medium">
                      Edit
                    </button>
                    <button className="text-wrale-danger hover:text-wrale-danger/80 text-sm font-medium">
                      Delete
                    </button>
                  </div>
                </div>

                <div className="text-sm text-gray-500 space-y-2">
                  <div>Condition: {rule.condition}</div>
                  <div>
                    Channels: {rule.channels.map(id => 
                      channels.find(c => c.id === id)?.name
                    ).join(', ')}
                  </div>
                </div>

                <div className="mt-4">
                  <label className="flex items-center">
                    <input
                      type="checkbox"
                      className="form-checkbox h-4 w-4 text-wrale-primary"
                      checked={rule.enabled}
                      onChange={() => {}}
                    />
                    <span className="ml-2 text-sm text-gray-600">
                      Rule enabled
                    </span>
                  </label>
                </div>
              </div>
            ))}
          </div>
        </div>

        <button className="mt-4 flex items-center text-sm text-wrale-primary hover:text-wrale-primary/80">
          <svg className="w-5 h-5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Add Rule
        </button>
      </section>
    </div>
  )
}