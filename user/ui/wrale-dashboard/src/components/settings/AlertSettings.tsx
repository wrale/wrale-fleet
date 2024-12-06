'use client'

import { useState, useEffect } from 'react'
import { FormInput, FormSelect, FormCheckbox } from '@/components/ui/Form'
import { validateRequired } from '@/lib/validation'
import { useLoading } from '@/components/ui/LoadingProvider'
import { alertsApi } from '@/services/api'

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

interface FormErrors {
  [key: string]: string | undefined
}

export function AlertSettings() {
  const { setIsLoading } = useLoading()
  const [channels, setChannels] = useState<AlertChannel[]>([])
  const [rules, setRules] = useState<AlertRule[]>([])
  const [errors, setErrors] = useState<FormErrors>({})
  const [isEditing, setIsEditing] = useState<'channel' | 'rule' | null>(null)
  const [editingItem, setEditingItem] = useState<any>(null)

  useEffect(() => {
    async function fetchData() {
      try {
        setIsLoading(true)
        const [channelsData, rulesData] = await Promise.all([
          alertsApi.getChannels(),
          alertsApi.getRules()
        ])
        setChannels(channelsData)
        setRules(rulesData)
      } catch (error) {
        console.error('Failed to fetch alert settings:', error)
      } finally {
        setIsLoading(false)
      }
    }

    fetchData()
  }, [setIsLoading])

  const validateChannel = (channel: Partial<AlertChannel>): boolean => {
    const newErrors: FormErrors = {}

    newErrors.name = validateRequired(channel.name || '')

    if (channel.type === 'email') {
      if (!channel.config?.recipients?.length) {
        newErrors.recipients = 'At least one recipient is required'
      }
    } else if (channel.type === 'slack') {
      newErrors.channelName = validateRequired(channel.config?.channel || '')
    } else if (channel.type === 'webhook') {
      newErrors.webhookUrl = validateRequired(channel.config?.webhook_url || '')
    }

    setErrors(newErrors)
    return Object.values(newErrors).every(e => !e)
  }

  const validateRule = (rule: Partial<AlertRule>): boolean => {
    const newErrors: FormErrors = {}

    newErrors.name = validateRequired(rule.name || '')
    newErrors.condition = validateRequired(rule.condition || '')
    
    if (!rule.channels?.length) {
      newErrors.channels = 'At least one channel must be selected'
    }

    setErrors(newErrors)
    return Object.values(newErrors).every(e => !e)
  }

  const handleEditChannel = (channel?: AlertChannel) => {
    setIsEditing('channel')
    setEditingItem(channel || {
      type: 'email',
      config: {},
      enabled: true
    })
    setErrors({})
  }

  const handleEditRule = (rule?: AlertRule) => {
    setIsEditing('rule')
    setEditingItem(rule || {
      severity: 'medium',
      channels: [],
      enabled: true
    })
    setErrors({})
  }

  const handleSaveChannel = async () => {
    if (!validateChannel(editingItem)) return

    try {
      setIsLoading(true)
      if (editingItem.id) {
        await alertsApi.updateChannel(editingItem.id, editingItem)
      } else {
        await alertsApi.createChannel(editingItem)
      }
      
      const updatedChannels = await alertsApi.getChannels()
      setChannels(updatedChannels)
      setIsEditing(null)
      setEditingItem(null)
    } catch (error) {
      console.error('Failed to save channel:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleSaveRule = async () => {
    if (!validateRule(editingItem)) return

    try {
      setIsLoading(true)
      if (editingItem.id) {
        await alertsApi.updateRule(editingItem.id, editingItem)
      } else {
        await alertsApi.createRule(editingItem)
      }
      
      const updatedRules = await alertsApi.getRules()
      setRules(updatedRules)
      setIsEditing(null)
      setEditingItem(null)
    } catch (error) {
      console.error('Failed to save rule:', error)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <section>
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-medium text-gray-900">Alert Channels</h2>
          <button
            onClick={() => handleEditChannel()}
            className="bg-wrale-primary text-white px-4 py-2 rounded-lg hover:bg-wrale-primary/90"
          >
            Add Channel
          </button>
        </div>

        <div className="bg-white shadow rounded-lg divide-y divide-gray-200">
          {channels.map(channel => (
            <div key={channel.id} className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-medium">{channel.name}</h3>
                  <p className="text-sm text-gray-500">{channel.type}</p>
                </div>
                <div className="space-x-3">
                  <button
                    onClick={() => handleEditChannel(channel)}
                    className="text-wrale-primary hover:text-wrale-primary/80 text-sm font-medium"
                  >
                    Edit
                  </button>
                  <FormCheckbox
                    label="Enabled"
                    checked={channel.enabled}
                    onChange={async (e) => {
                      try {
                        setIsLoading(true)
                        await alertsApi.updateChannel(channel.id, {
                          ...channel,
                          enabled: e.target.checked
                        })
                        const updatedChannels = await alertsApi.getChannels()
                        setChannels(updatedChannels)
                      } catch (error) {
                        console.error('Failed to update channel:', error)
                      } finally {
                        setIsLoading(false)
                      }
                    }}
                  />
                </div>
              </div>
            </div>
          ))}
        </div>
      </section>

      {isEditing === 'channel' && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-lg w-full">
            <h3 className="text-lg font-medium mb-4">
              {editingItem.id ? 'Edit Channel' : 'New Channel'}
            </h3>

            <div className="space-y-4">
              <FormInput
                label="Channel Name"
                value={editingItem.name || ''}
                onChange={(e) => setEditingItem({ ...editingItem, name: e.target.value })}
                error={errors.name}
                required
              />

              <FormSelect
                label="Channel Type"
                value={editingItem.type}
                onChange={(e) => setEditingItem({
                  ...editingItem,
                  type: e.target.value,
                  config: {}
                })}
                options={[
                  { value: 'email', label: 'Email' },
                  { value: 'slack', label: 'Slack' },
                  { value: 'webhook', label: 'Webhook' }
                ]}
                required
              />

              {editingItem.type === 'email' && (
                <FormInput
                  label="Recipients (comma-separated)"
                  value={editingItem.config?.recipients?.join(', ') || ''}
                  onChange={(e) => setEditingItem({
                    ...editingItem,
                    config: {
                      ...editingItem.config,
                      recipients: e.target.value.split(',').map(s => s.trim())
                    }
                  })}
                  error={errors.recipients}
                  required
                />
              )}

              {editingItem.type === 'slack' && (
                <FormInput
                  label="Slack Channel"
                  value={editingItem.config?.channel || ''}
                  onChange={(e) => setEditingItem({
                    ...editingItem,
                    config: {
                      ...editingItem.config,
                      channel: e.target.value
                    }
                  })}
                  error={errors.channelName}
                  required
                />
              )}

              {editingItem.type === 'webhook' && (
                <FormInput
                  label="Webhook URL"
                  value={editingItem.config?.webhook_url || ''}
                  onChange={(e) => setEditingItem({
                    ...editingItem,
                    config: {
                      ...editingItem.config,
                      webhook_url: e.target.value
                    }
                  })}
                  error={errors.webhookUrl}
                  required
                />
              )}
            </div>

            <div className="mt-6 flex justify-end space-x-3">
              <button
                onClick={() => {
                  setIsEditing(null)
                  setEditingItem(null)
                }}
                className="px-4 py-2 text-gray-700 hover:text-gray-900"
              >
                Cancel
              </button>
              <button
                onClick={handleSaveChannel}
                className="px-4 py-2 bg-wrale-primary text-white rounded-lg hover:bg-wrale-primary/90"
              >
                Save
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Similar modal for rules */}
    </div>
  )
}