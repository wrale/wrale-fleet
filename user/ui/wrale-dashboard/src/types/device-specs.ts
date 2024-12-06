import type { PhysicalMetric } from './maintenance'

export interface MetricSpec {
  min: number
  max: number
  unit: string
  recommendedMin?: number
  recommendedMax?: number
  criticalMin?: number
  criticalMax?: number
  sampleInterval?: number  // in seconds
  defaultDuration?: number // in seconds for maintenance checks
}

export interface PowerSpec {
  voltage: MetricSpec
  current: MetricSpec
  power: MetricSpec
  efficiency: MetricSpec
}

export interface EnvironmentalSpec {
  temperature: MetricSpec
  humidity: MetricSpec
  fanSpeed?: MetricSpec
}

export interface StorageSpec {
  type: 'sd' | 'emmc' | 'nvme'
  size: number  // in GB
  readSpeed: MetricSpec
  writeSpeed: MetricSpec
  healthThresholds: {
    warning: number
    critical: number
  }
}

export interface NetworkSpec {
  interfaces: ('ethernet' | 'wifi' | 'bluetooth')[]
  latency: MetricSpec
  bandwidth: MetricSpec
}

export interface DeviceSpec {
  id: string
  name: string
  model: string
  formFactor: {
    height: number  // mm
    width: number   // mm
    depth: number   // mm
    weight: number  // g
    rackUnits: number
  }
  power: PowerSpec
  environmental: EnvironmentalSpec
  storage: StorageSpec
  network: NetworkSpec
  metrics: {
    [key in PhysicalMetric]: MetricSpec
  }
  maintenance: {
    recommendedInterval: number  // days
    criticalComponents: string[]
    commonIssues: {
      issue: string
      indicators: PhysicalMetric[]
      priority: 'low' | 'medium' | 'high'
    }[]
  }
}