'use client'

import { useState } from 'react'
import { FormInput, FormSelect, FormCheckbox } from '@/components/ui/Form'
import { ConditionBuilder } from './ConditionBuilder'
import type { MaintenanceRule, MaintenanceCondition, MaintenanceAction } from '@/types/maintenance'
import { validateMaintenanceRule } from '@/lib/maintenanceValidation'

interface MaintenanceRuleModalProps {
  rule: Partial<MaintenanceRule>
  onSave: (rule: MaintenanceRule) => Promise<void>
  onClose: () => void
}

const PRIORITY_OPTIONS = [
  { value: 'low', label: 'Low' },
  { value: 'medium', label: 'Medium' },
  { value: 'high', label: 'High' }
]

const ACTION_OPTIONS = [
  { value: 'inspect', label: 'Physical Inspection Required' },
  { value: 'replace', label: 'Component Replacement Needed' },
  { value: 'clean', label: 'Cleaning/Maintenance Required' },
  { value: 'calibrate', label: 'Calibration Needed' },
  { value: 'update', label: 'Software Update Required' },
  { value: 'reboot', label: 'Device Reboot Required' }
]

export function MaintenanceRuleModal({ rule, onSave, onClose }: MaintenanceRuleModalProps) {
  const [editingRule, setEditingRule] = useState<Partial<MaintenanceRule>>(rule)
  const [errors, setErrors] = useState<{ [key: string]: string }>({})
  const [conditionValidStates, setConditionValidStates] = useState<boolean[]>([])

  const handleAddCondition = () => {
    setEditingRule(prev => ({
      ...prev,
      conditions: [
        ...(prev.conditions || []),
        { metric: 'temperature', operator: '>', threshold: 0, duration: 30 }
      ]
    }))
    setConditionValidStates(prev => [...prev, false])
  }

  const handleRemoveCondition = (index: number) => {
    setEditingRule(prev => ({
      ...prev,
      conditions: prev.conditions?.filter((_, i) => i !== index)
    }))
    setConditionValidStates(prev => prev.filter((_, i) => i !== index))
  }

  const handleConditionChange = (index: number, condition: Partial<MaintenanceCondition>) => {
    setEditingRule(prev => ({
      ...prev,
      conditions: prev.conditions?.map((c, i) => i === index ? { ...c, ...condition } : c)
    }))
  }

  const handleConditionValid = (index: number, isValid: boolean) => {
    setConditionValidStates(prev => {
      const newStates = [...prev]
      newStates[index] = isValid
      return newStates
    })
  }

  const handleActionToggle = (action: MaintenanceAction) => {
    setEditingRule(prev => ({
      ...prev,
      recommendedActions: prev.recommendedActions?.includes(action)
        ? prev.recommendedActions.filter(a => a !== action)
        : [...(prev.recommendedActions || []), action]
    }))
  }

  const handleSave = async () => {
    const validationErrors = validateMaintenanceRule(editingRule)
    
    if (Object.keys(validationErrors).length > 0 || !conditionValidStates.every(Boolean)) {
      setErrors(validationErrors)
      return
    }

    try {
      await onSave(editingRule as MaintenanceRule)
      onClose()
    } catch (error) {
      console.error('Failed to save maintenance rule:', error)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg w-full max-w-4xl max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          <h2 className="text-lg font-medium mb-4">
            {rule.id ? 'Edit Maintenance Rule' : 'New Maintenance Rule'}
          </h2>

          <div className="space-y-6">
            <FormInput
              label="Rule Name"
              value={editingRule.name || ''}
              onChange={(e) => setEditingRule({ ...editingRule, name: e.target.value })}
              error={errors.name}
              required
            />

            <FormInput
              label="Description"
              value={editingRule.description || ''}
              onChange={(e) => setEditingRule({ ...editingRule, description: e.target.value })}
              error={errors.description}
              required
            />

            <div>
              <h3 className="text-sm font-medium text-gray-700 mb-2">Conditions</h3>
              {errors.conditions && (
                <p className="text-sm text-wrale-danger mb-2">{errors.conditions}</p>
              )}
              
              <div className="space-y-4">
                {editingRule.conditions?.map((condition, index) => (
                  <div key={index} className="relative">
                    <button
                      onClick={() => handleRemoveCondition(index)}
                      className="absolute -right-2 -top-2 p-1 bg-wrale-danger text-white rounded-full hover:bg-wrale-danger/90"
                    >
                      ×
                    </button>
                    <ConditionBuilder
                      condition={condition}
                      onChange={(updated) => handleConditionChange(index, updated)}
                      onValid={(isValid) => handleConditionValid(index, isValid)}
                      errors={errors}
                    />
                  </div>
                ))}
              </div>

              <button
                onClick={handleAddCondition}
                className="mt-4 text-wrale-primary hover:text-wrale-primary/80 text-sm"
              >
                + Add Condition
              </button>
            </div>

            <FormSelect
              label="Priority"
              value={editingRule.priority || 'medium'}
              onChange={(e) => setEditingRule({
                ...editingRule,
                priority: e.target.value as MaintenanceRule['priority']
              })}
              options={PRIORITY_OPTIONS}
              error={errors.priority}
              required
            />

            <div>
              <h3 className="text-sm font-medium text-gray-700 mb-2">Recommended Actions</h3>
              {errors.actions && (
                <p className="text-sm text-wrale-danger mb-2">{errors.actions}</p>
              )}
              
              <div className="space-y-2">
                {ACTION_OPTIONS.map(action => (
                  <FormCheckbox
                    key={action.value}
                    label={action.label}
                    checked={editingRule.recommendedActions?.includes(action.value as MaintenanceAction)}
                    onChange={() => handleActionToggle(action.value as MaintenanceAction)}
                  />
                ))}
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
              onClick={handleSave}
              className="px-4 py-2 bg-wrale-primary text-white rounded-lg hover:bg-wrale-primary/90"
            >
              Save Rule
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}