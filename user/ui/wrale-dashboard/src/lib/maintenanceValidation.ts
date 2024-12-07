import type { MaintenanceRule, MaintenanceCondition, PhysicalMetric } from '@/types/maintenance'

export function validateCondition(condition: MaintenanceCondition): string | undefined {
  // Validate metric exists and is valid for physical device
  if (!isValidPhysicalMetric(condition.metric)) {
    return `Invalid metric: ${condition.metric}`
  }

  // Validate threshold values are within physical bounds
  const metricBounds = getMetricBounds(condition.metric)
  if (condition.threshold < metricBounds.min || condition.threshold > metricBounds.max) {
    return `Threshold must be between ${metricBounds.min} and ${metricBounds.max} ${metricBounds.unit}`
  }

  // Validate duration makes sense for the metric
  if (condition.duration < getMinDuration(condition.metric)) {
    return `Duration must be at least ${getMinDuration(condition.metric)}s for ${condition.metric}`
  }

  return undefined
}

export function validateMaintenanceRule(rule: Partial<MaintenanceRule>): { [key: string]: string } {
  const errors: { [key: string]: string } = {}

  // Basic field validation
  if (!rule.name?.trim()) {
    errors.name = 'Name is required'
  }

  if (!rule.description?.trim()) {
    errors.description = 'Description is required'
  }

  // Validate conditions exist and make sense
  if (!rule.conditions?.length) {
    errors.conditions = 'At least one condition is required'
  } else {
    // Validate each condition
    const conditionErrors = rule.conditions.map(validateCondition)
    if (conditionErrors.some(error => error !== undefined)) {
      errors.conditions = 'One or more conditions are invalid'
    }

    // Check for conflicting conditions
    if (hasConflictingConditions(rule.conditions)) {
      errors.conditions = 'Conditions contain conflicts'
    }
  }

  // Validate actions
  if (!rule.recommendedActions?.length) {
    errors.actions = 'At least one recommended action is required'
  }

  return errors
}

// Check if a metric is valid for physical device monitoring
function isValidPhysicalMetric(metric: string): metric is PhysicalMetric {
  const validMetrics: PhysicalMetric[] = [
    'temperature',
    'voltage',
    'current',
    'power',
    'humidity',
    'fanSpeed',
    'storageHealth',
    'networkLatency',
    'cpuUsage',
    'memoryUsage'
  ]
  return validMetrics.includes(metric as PhysicalMetric)
}

// Get valid bounds for each metric
function getMetricBounds(metric: PhysicalMetric): { min: number; max: number; unit: string } {
  switch (metric) {
    case 'temperature':
      return { min: -20, max: 100, unit: '°C' }
    case 'voltage':
      return { min: 0, max: 12, unit: 'V' }
    case 'current':
      return { min: 0, max: 3, unit: 'A' }
    case 'power':
      return { min: 0, max: 25, unit: 'W' }
    case 'humidity':
      return { min: 0, max: 100, unit: '%' }
    case 'fanSpeed':
      return { min: 0, max: 5000, unit: 'RPM' }
    case 'storageHealth':
      return { min: 0, max: 100, unit: '%' }
    case 'networkLatency':
      return { min: 0, max: 1000, unit: 'ms' }
    case 'cpuUsage':
      return { min: 0, max: 100, unit: '%' }
    case 'memoryUsage':
      return { min: 0, max: 100, unit: '%' }
  }
}

// Get minimum duration for detecting an issue with each metric
function getMinDuration(metric: PhysicalMetric): number {
  switch (metric) {
    case 'temperature':
    case 'voltage':
    case 'current':
    case 'power':
      return 30 // 30 seconds minimum for power-related issues
    case 'humidity':
    case 'fanSpeed':
      return 60 // 1 minute for environmental factors
    case 'storageHealth':
      return 300 // 5 minutes for storage health checks
    case 'networkLatency':
      return 10 // 10 seconds for network issues
    case 'cpuUsage':
    case 'memoryUsage':
      return 120 // 2 minutes for resource usage
  }
}

// Check for conflicting conditions in the same rule
function hasConflictingConditions(conditions: MaintenanceCondition[]): boolean {
  // Group conditions by metric
  const metricGroups = conditions.reduce((groups, condition) => {
    const metric = condition.metric
    if (!groups[metric]) {
      groups[metric] = []
    }
    groups[metric].push(condition)
    return groups
  }, {} as Record<string, MaintenanceCondition[]>)

  // Check each group for conflicts
  return Object.values(metricGroups).some(group => {
    if (group.length <= 1) return false

    // Sort by threshold
    const sorted = [...group].sort((a, b) => a.threshold - b.threshold)

    // Check for overlapping durations
    for (let i = 0; i < sorted.length - 1; i++) {
      const current = sorted[i]
      const next = sorted[i + 1]
      
      if (current.threshold >= next.threshold) {
        return true // Conflict found
      }
      
      // If lower threshold has longer duration than higher threshold
      if (current.duration > next.duration) {
        return true // Conflict found
      }
    }

    return false
  })
}