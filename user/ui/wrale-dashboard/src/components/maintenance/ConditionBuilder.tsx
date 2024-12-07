'use client'

import { useState, useEffect } from 'react'
import { FormInput, FormSelect } from '@/components/ui/Form'
import type { MaintenanceCondition, PhysicalMetric, ComparisonOperator } from '@/types/maintenance'
import { getMetricBounds, getMinDuration } from '@/lib/maintenanceValidation'

interface ConditionBuilderProps {
  condition: Partial<MaintenanceCondition>
  onChange: (condition: Partial<MaintenanceCondition>) => void
  onValid: (isValid: boolean) => void
  errors?: { [key: string]: string }
}

const METRICS: { value: PhysicalMetric; label: string }[] = [
  { value: 'temperature', label: 'Temperature' },
  { value: 'voltage', label: 'Voltage' },
  { value: 'current', label: 'Current Draw' },
  { value: 'power', label: 'Power Consumption' },
  { value: 'humidity', label: 'Humidity' },
  { value: 'fanSpeed', label: 'Fan Speed' },
  { value: 'storageHealth', label: 'Storage Health' },
  { value: 'networkLatency', label: 'Network Latency' },
  { value: 'cpuUsage', label: 'CPU Usage' },
  { value: 'memoryUsage', label: 'Memory Usage' }
]

const OPERATORS: { value: ComparisonOperator; label: string }[] = [
  { value: '>', label: 'Greater than' },
  { value: '<', label: 'Less than' },
  { value: '>=', label: 'Greater than or equal to' },
  { value: '<=', label: 'Less than or equal to' },
  { value: '=', label: 'Equal to' },
  { value: '!=', label: 'Not equal to' }
]

export function ConditionBuilder({ condition, onChange, onValid, errors = {} }: ConditionBuilderProps) {
  const [localErrors, setLocalErrors] = useState<{ [key: string]: string }>({})
  const [bounds, setBounds] = useState({ min: 0, max: 100, unit: '' })
  const [minDuration, setMinDuration] = useState(0)

  useEffect(() => {
    if (condition.metric) {
      const newBounds = getMetricBounds(condition.metric)
      const newMinDuration = getMinDuration(condition.metric)
      setBounds(newBounds)
      setMinDuration(newMinDuration)
      
      // Validate threshold when bounds change
      validateThreshold(condition.threshold, newBounds)
      validateDuration(condition.duration, newMinDuration)
    }
  }, [condition.metric])

  const validateThreshold = (value: number | undefined, currentBounds = bounds) => {
    const errors: { [key: string]: string } = { ...localErrors }
    delete errors.threshold

    if (value === undefined) {
      errors.threshold = 'Threshold is required'
    } else if (value < currentBounds.min || value > currentBounds.max) {
      errors.threshold = `Must be between ${currentBounds.min} and ${currentBounds.max} ${currentBounds.unit}`
    }

    setLocalErrors(errors)
    onValid(Object.keys(errors).length === 0)
    return errors
  }

  const validateDuration = (value: number | undefined, currentMinDuration = minDuration) => {
    const errors: { [key: string]: string } = { ...localErrors }
    delete errors.duration

    if (value === undefined) {
      errors.duration = 'Duration is required'
    } else if (value < currentMinDuration) {
      errors.duration = `Must be at least ${currentMinDuration} seconds`
    }

    setLocalErrors(errors)
    onValid(Object.keys(errors).length === 0)
    return errors
  }

  return (
    <div className="space-y-4 p-4 bg-gray-50 rounded-lg">
      <div className="text-sm text-gray-500 mb-4">
        Configure condition parameters based on physical device specifications
      </div>

      <FormSelect
        label="Metric"
        value={condition.metric || ''}
        onChange={(e) => {
          const metric = e.target.value as PhysicalMetric
          onChange({
            ...condition,
            metric,
            unit: getMetricBounds(metric).unit
          })
        }}
        options={METRICS}
        error={errors.metric || localErrors.metric}
        required
      />

      <FormSelect
        label="Operator"
        value={condition.operator || '>'}
        onChange={(e) => onChange({
          ...condition,
          operator: e.target.value as ComparisonOperator
        })}
        options={OPERATORS}
        error={errors.operator || localErrors.operator}
        required
      />

      <FormInput
        label={`Threshold (${bounds.unit})`}
        type="number"
        min={bounds.min}
        max={bounds.max}
        step="0.1"
        value={condition.threshold || ''}
        onChange={(e) => {
          const value = parseFloat(e.target.value)
          onChange({
            ...condition,
            threshold: value
          })
          validateThreshold(value)
        }}
        error={errors.threshold || localErrors.threshold}
        required
      />

      <FormInput
        label="Duration (seconds)"
        type="number"
        min={minDuration}
        value={condition.duration || ''}
        onChange={(e) => {
          const value = parseInt(e.target.value)
          onChange({
            ...condition,
            duration: value
          })
          validateDuration(value)
        }}
        error={errors.duration || localErrors.duration}
        required
      />

      {condition.metric && (
        <div className="mt-4 border-t border-gray-200 pt-4">
          <div className="text-sm text-gray-500 space-y-1">
            <p>Valid range: {bounds.min} to {bounds.max} {bounds.unit}</p>
            <p>Minimum duration: {minDuration} seconds</p>
            <p>Recommended for: {METRICS.find(m => m.value === condition.metric)?.label}</p>
          </div>
        </div>
      )}
    </div>
  )
}