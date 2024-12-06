export type MaintenancePriority = 'low' | 'medium' | 'high'

export type PhysicalMetric = 
  | 'temperature'
  | 'voltage'
  | 'current'
  | 'power'
  | 'humidity'
  | 'fanSpeed'
  | 'storageHealth'
  | 'networkLatency'
  | 'cpuUsage'
  | 'memoryUsage'

export type ComparisonOperator = '>' | '<' | '>=' | '<=' | '=' | '!='

export interface MaintenanceCondition {
  metric: PhysicalMetric
  operator: ComparisonOperator
  threshold: number
  duration: number // in seconds
  unit?: string
}

export type MaintenanceAction = 
  | 'inspect'
  | 'replace'
  | 'clean'
  | 'calibrate'
  | 'update'
  | 'reboot'

export interface MaintenanceRule {
  id: string
  name: string
  description: string
  conditions: MaintenanceCondition[]
  priority: MaintenancePriority
  recommendedActions: MaintenanceAction[]
  enabled: boolean
  lastUpdated: string
  deviceTypes?: string[]
  locations?: string[]
}