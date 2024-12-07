'use client'

import { FormInput, FormSelect, FormCheckbox } from '@/components/ui/Form'
import type { AlertRule, AlertChannel } from '@/types/alerts'

interface AlertRuleModalProps {
  rule: Partial<AlertRule>
  channels: AlertChannel[]
  errors: { [key: string]: string | undefined }
  onClose: () => void
  onSave: () => void
  onChange: (rule: Partial<AlertRule>) => void
}

export function AlertRuleModal({ 
  rule, 
  channels, 
  errors, 
  onClose, 
  onSave, 
  onChange 
}: AlertRuleModalProps) {
  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg p-6 max-w-lg w-full">
        <h3 className="text-lg font-medium mb-4">
          {rule.id ? 'Edit Rule' : 'New Rule'}
        </h3>

        <div className="space-y-4">
          <FormInput
            label="Rule Name"
            value={rule.name || ''}
            onChange={(e) => onChange({ ...rule, name: e.target.value })}
            error={errors.name}
            required
          />

          <FormInput
            label="Condition"
            value={rule.condition || ''}
            onChange={(e) => onChange({ ...rule, condition: e.target.value })}
            error={errors.condition}
            placeholder="e.g., temperature > 50 for 5m"
            required
          />

          <FormSelect
            label="Severity"
            value={rule.severity || 'medium'}
            onChange={(e) => onChange({ ...rule, severity: e.target.value as AlertRule['severity'] })}
            options={[
              { value: 'low', label: 'Low' },
              { value: 'medium', label: 'Medium' },
              { value: 'high', label: 'High' }
            ]}
            required
          />

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Alert Channels
              {errors.channels && (
                <span className="text-wrale-danger text-xs ml-2">{errors.channels}</span>
              )}
            </label>
            <div className="space-y-2">
              {channels.map(channel => (
                <FormCheckbox
                  key={channel.id}
                  label={channel.name}
                  checked={rule.channels?.includes(channel.id)}
                  onChange={(e) => {
                    const newChannels = e.target.checked
                      ? [...(rule.channels || []), channel.id]
                      : (rule.channels || []).filter(id => id !== channel.id)
                    onChange({ ...rule, channels: newChannels })
                  }}
                />
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Actions
            </label>
            <div className="space-y-2">
              <FormCheckbox
                label="Auto-restart device if offline"
                checked={rule.actions?.includes('restart')}
                onChange={(e) => {
                  const newActions = e.target.checked
                    ? [...(rule.actions || []), 'restart']
                    : (rule.actions || []).filter(action => action !== 'restart')
                  onChange({ ...rule, actions: newActions })
                }}
              />
              <FormCheckbox
                label="Create maintenance ticket"
                checked={rule.actions?.includes('ticket')}
                onChange={(e) => {
                  const newActions = e.target.checked
                    ? [...(rule.actions || []), 'ticket']
                    : (rule.actions || []).filter(action => action !== 'ticket')
                  onChange({ ...rule, actions: newActions })
                }}
              />
            </div>
          </div>
        </div>

        <div className="mt-6 flex justify-end space-x-3">
          <button
            onClick={onClose}
            className="px-4 py-2 text-gray-700 hover:text-gray-900"
          >
            Cancel
          </button>
          <button
            onClick={onSave}
            className="px-4 py-2 bg-wrale-primary text-white rounded-lg hover:bg-wrale-primary/90"
          >
            Save
          </button>
        </div>
      </div>
    </div>
  )
}