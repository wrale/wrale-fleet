'use client'

import { useState, useEffect } from 'react'
import { FormInput, FormSelect, FormCheckbox } from '@/components/ui/Form'
import { validateRequired } from '@/lib/validation'
import { useLoading } from '@/components/ui/LoadingProvider'
import { alertsApi } from '@/services/api'
import { AlertRuleModal } from './AlertRuleModal'
import type { AlertChannel, AlertRule } from '@/types/alerts'

interface FormErrors {
  [key: string]: string | undefined
}

export function AlertSettings() {
  const { setIsLoading } = useLoading()
  const [channels, setChannels] = useState<AlertChannel[]>([])
  const [rules, setRules] = useState<AlertRule[]>([])
  const [errors, setErrors] = useState<FormErrors>({})
  const [isEditing, setIsEditing] = useState<'channel' | 'rule' | null>(null)
  const [editingChannel, setEditingChannel] = useState<Partial<AlertChannel> | null>(null)
  const [editingRule, setEditingRule] = useState<Partial<AlertRule> | null>(null)

  useEffect(() => {
    loadData()
  }, [])

  async function loadData() {
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

  const handleSaveChannel = async () => {
    if (!editingChannel || !validateChannel(editingChannel)) return

    try {
      setIsLoading(true)
      if (editingChannel.id) {
        await alertsApi.updateChannel(editingChannel.id, editingChannel)
      } else {
        await alertsApi.createChannel(editingChannel)
      }
      await loadData()
      setIsEditing(null)
      setEditingChannel(null)
      setErrors({})
    } catch (error) {
      console.error('Failed to save channel:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleSaveRule = async () => {
    if (!editingRule || !validateRule(editingRule)) return

    try {
      setIsLoading(true)
      if (editingRule.id) {
        await alertsApi.updateRule(editingRule.id, editingRule)
      } else {
        await alertsApi.createRule(editingRule)
      }
      await loadData()
      setIsEditing(null)
      setEditingRule(null)
      setErrors({})
    } catch (error) {
      console.error('Failed to save rule:', error)
    } finally {
      setIsLoading(false)
    }
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

  return (
    <div className="space-y-6">
      <section>
        {/* Channels Section */}
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-medium text-gray-900">Alert Channels</h2>
          <button
            onClick={() => {
              setIsEditing('channel')
              setEditingChannel({
                type: 'email',
                config: {},
                enabled: true
              })
            }}
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
                  <p className="text-sm text-gray-500 mt-1">
                    {channel.type === 'email' && `Recipients: ${channel.config.recipients?.join(', ')}`}
                    {channel.type === 'slack' && `Channel: ${channel.config.channel}`}
                    {channel.type === 'webhook' && 'Webhook configured'}
                  </p>
                </div>
                <div className="space-x-3">
                  <button
                    onClick={() => {
                      setIsEditing('channel')
                      setEditingChannel(channel)
                    }}
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
                        await loadData()
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

      <section>
        {/* Rules Section */}
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-medium text-gray-900">Alert Rules</h2>
          <button
            onClick={() => {
              setIsEditing('rule')
              setEditingRule({
                severity: 'medium',
                channels: [],
                enabled: true
              })
            }}
            className="bg-wrale-primary text-white px-4 py-2 rounded-lg hover:bg-wrale-primary/90"
          >
            Add Rule
          </button>
        </div>

        <div className="bg-white shadow rounded-lg divide-y divide-gray-200">
          {rules.map(rule => (
            <div key={rule.id} className="p-6">
              <div className="flex items-center justify-between mb-2">
                <div>
                  <h3 className="text-lg font-medium">{rule.name}</h3>
                  <p className="text-sm text-gray-500 mt-1">{rule.condition}</p>
                </div>
                <div className="space-x-3">
                  <button
                    onClick={() => {
                      setIsEditing('rule')
                      setEditingRule(rule)
                    }}
                    className="text-wrale-primary hover:text-wrale-primary/80 text-sm font-medium"
                  >
                    Edit
                  </button>
                  <FormCheckbox
                    label="Enabled"
                    checked={rule.enabled}
                    onChange={async (e) => {
                      try {
                        setIsLoading(true)
                        await alertsApi.updateRule(rule.id, {
                          ...rule,
                          enabled: e.target.checked
                        })
                        await loadData()
                      } catch (error) {
                        console.error('Failed to update rule:', error)
                      } finally {
                        setIsLoading(false)
                      }
                    }}
                  />
                </div>
              </div>
              <div className="text-sm text-gray-500">
                Channels: {rule.channels.map(id => 
                  channels.find(c => c.id === id)?.name
                ).filter(Boolean).join(', ')}
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* Channel Modal */}
      {isEditing === 'channel' && editingChannel && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-lg w-full">
            <h3 className="text-lg font-medium mb-4">
              {editingChannel.id ? 'Edit Channel' : 'New Channel'}
            </h3>
            
            <div className="space-y-4">
              <FormInput
                label="Channel Name"
                value={editingChannel.name || ''}
                onChange={(e) => setEditingChannel({
                  ...editingChannel,
                  name: e.target.value
                })}
                error={errors.name}
                required
              />

              <FormSelect
                label="Channel Type"
                value={editingChannel.type}
                onChange={(e) => setEditingChannel({
                  ...editingChannel,
                  type: e.target.value as AlertChannel['type'],
                  config: {}
                })}
                options={[
                  { value: 'email', label: 'Email' },
                  { value: 'slack', label: 'Slack' },
                  { value: 'webhook', label: 'Webhook' }
                ]}
                required
              />

              {editingChannel.type === 'email' && (
                <FormInput
                  label="Recipients (comma-separated)"
                  value={editingChannel.config?.recipients?.join(', ') || ''}
                  onChange={(e) => setEditingChannel({
                    ...editingChannel,
                    config: {
                      ...editingChannel.config,
                      recipients: e.target.value.split(',').map(s => s.trim())
                    }
                  })}
                  error={errors.recipients}
                  required
                />
              )}

              {editingChannel.type === 'slack' && (
                <FormInput
                  label="Slack Channel"
                  value={editingChannel.config?.channel || ''}
                  onChange={(e) => setEditingChannel({
                    ...editingChannel,
                    config: {
                      ...editingChannel.config,
                      channel: e.target.value
                    }
                  })}
                  error={errors.channelName}
                  required
                />
              )}

              {editingChannel.type === 'webhook' && (
                <FormInput
                  label="Webhook URL"
                  value={editingChannel.config?.webhook_url || ''}
                  onChange={(e) => setEditingChannel({
                    ...editingChannel,
                    config: {
                      ...editingChannel.config,
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
                  setEditingChannel(null)
                  setErrors({})
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

      {/* Rule Modal */}
      {isEditing === 'rule' && editingRule && (
        <AlertRuleModal
          rule={editingRule}
          channels={channels}
          errors={errors}
          onClose={() => {
            setIsEditing(null)
            setEditingRule(null)
            setErrors({})
          }}
          onSave={handleSaveRule}
          onChange={setEditingRule}
        />
      )}
    </div>
  )
}