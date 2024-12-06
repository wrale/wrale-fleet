import type { DeviceSpec } from '@/types/device-specs'

export const RPI4_SPECS: DeviceSpec = {
  id: 'rpi4b',
  name: 'Raspberry Pi 4 Model B',
  model: 'BCM2711',
  formFactor: {
    height: 85,
    width: 56,
    depth: 17,
    weight: 46,
    rackUnits: 1
  },
  power: {
    voltage: {
      min: 4.7,
      max: 5.3,
      unit: 'V',
      recommendedMin: 4.9,
      recommendedMax: 5.1,
      sampleInterval: 1,
      defaultDuration: 30
    },
    current: {
      min: 0,
      max: 3.0,
      unit: 'A',
      recommendedMax: 2.5,
      criticalMax: 2.8,
      sampleInterval: 1,
      defaultDuration: 30
    },
    power: {
      min: 0,
      max: 15,
      unit: 'W',
      recommendedMax: 12.5,
      criticalMax: 14,
      sampleInterval: 1,
      defaultDuration: 60
    },
    efficiency: {
      min: 0,
      max: 100,
      unit: '%',
      recommendedMin: 85,
      sampleInterval: 60
    }
  },
  environmental: {
    temperature: {
      min: -20,
      max: 85,
      unit: '°C',
      recommendedMin: 10,
      recommendedMax: 50,
      criticalMax: 80,
      sampleInterval: 5,
      defaultDuration: 300
    },
    humidity: {
      min: 0,
      max: 100,
      unit: '%',
      recommendedMin: 20,
      recommendedMax: 80,
      criticalMax: 90,
      sampleInterval: 60,
      defaultDuration: 300
    },
    fanSpeed: {
      min: 0,
      max: 5000,
      unit: 'RPM',
      recommendedMin: 1000,
      sampleInterval: 5,
      defaultDuration: 60
    }
  },
  storage: {
    type: 'sd',
    size: 32,
    readSpeed: {
      min: 0,
      max: 100,
      unit: 'MB/s',
      recommendedMin: 40,
      sampleInterval: 300
    },
    writeSpeed: {
      min: 0,
      max: 100,
      unit: 'MB/s',
      recommendedMin: 20,
      sampleInterval: 300
    },
    healthThresholds: {
      warning: 80,
      critical: 90
    }
  },
  network: {
    interfaces: ['ethernet', 'wifi', 'bluetooth'],
    latency: {
      min: 0,
      max: 1000,
      unit: 'ms',
      recommendedMax: 100,
      criticalMax: 500,
      sampleInterval: 10,
      defaultDuration: 60
    },
    bandwidth: {
      min: 0,
      max: 1000,
      unit: 'Mbps',
      recommendedMin: 100,
      sampleInterval: 60
    }
  },
  metrics: {
    temperature: {
      min: -20,
      max: 85,
      unit: '°C',
      recommendedMax: 50,
      criticalMax: 80,
      sampleInterval: 5,
      defaultDuration: 300
    },
    voltage: {
      min: 4.7,
      max: 5.3,
      unit: 'V',
      recommendedMin: 4.9,
      recommendedMax: 5.1,
      sampleInterval: 1,
      defaultDuration: 30
    },
    current: {
      min: 0,
      max: 3.0,
      unit: 'A',
      recommendedMax: 2.5,
      sampleInterval: 1,
      defaultDuration: 30
    },
    power: {
      min: 0,
      max: 15,
      unit: 'W',
      recommendedMax: 12.5,
      sampleInterval: 1,
      defaultDuration: 60
    },
    humidity: {
      min: 0,
      max: 100,
      unit: '%',
      recommendedMax: 80,
      criticalMax: 90,
      sampleInterval: 60,
      defaultDuration: 300
    },
    fanSpeed: {
      min: 0,
      max: 5000,
      unit: 'RPM',
      recommendedMin: 1000,
      sampleInterval: 5,
      defaultDuration: 60
    },
    storageHealth: {
      min: 0,
      max: 100,
      unit: '%',
      recommendedMin: 20,
      criticalMin: 10,
      sampleInterval: 3600,
      defaultDuration: 3600
    },
    networkLatency: {
      min: 0,
      max: 1000,
      unit: 'ms',
      recommendedMax: 100,
      criticalMax: 500,
      sampleInterval: 10,
      defaultDuration: 60
    },
    cpuUsage: {
      min: 0,
      max: 100,
      unit: '%',
      recommendedMax: 80,
      criticalMax: 95,
      sampleInterval: 5,
      defaultDuration: 300
    },
    memoryUsage: {
      min: 0,
      max: 100,
      unit: '%',
      recommendedMax: 85,
      criticalMax: 95,
      sampleInterval: 5,
      defaultDuration: 300
    }
  },
  maintenance: {
    recommendedInterval: 90, // 90 days
    criticalComponents: [
      'SD Card',
      'Power Supply',
      'Cooling System',
      'Network Interfaces'
    ],
    commonIssues: [
      {
        issue: 'SD Card Degradation',
        indicators: ['storageHealth', 'networkLatency'],
        priority: 'high'
      },
      {
        issue: 'Overheating Risk',
        indicators: ['temperature', 'fanSpeed'],
        priority: 'high'
      },
      {
        issue: 'Power Supply Instability',
        indicators: ['voltage', 'current'],
        priority: 'medium'
      },
      {
        issue: 'Network Performance Degradation',
        indicators: ['networkLatency'],
        priority: 'medium'
      }
    ]
  }
}

const DEVICE_SPECS: { [key: string]: DeviceSpec } = {
  'rpi4b': RPI4_SPECS,
  // Add more device specs as needed
}

export function getDeviceSpecs(modelId: string): DeviceSpec | undefined {
  return DEVICE_SPECS[modelId]
}

export function getMetricSpec(modelId: string, metric: keyof DeviceSpec['metrics']): MetricSpec | undefined {
  const specs = getDeviceSpecs(modelId)
  return specs?.metrics[metric]
}

export function validateMetricValue(
  modelId: string,
  metric: keyof DeviceSpec['metrics'],
  value: number
): { valid: boolean; message?: string } {
  const spec = getMetricSpec(modelId, metric)
  if (!spec) {
    return { valid: false, message: 'Unknown metric for device model' }
  }

  if (value < spec.min || value > spec.max) {
    return {
      valid: false,
      message: `Value must be between ${spec.min} and ${spec.max} ${spec.unit}`
    }
  }

  if (spec.criticalMin && value <= spec.criticalMin) {
    return {
      valid: false,
      message: `Critical: Value below minimum threshold of ${spec.criticalMin} ${spec.unit}`
    }
  }

  if (spec.criticalMax && value >= spec.criticalMax) {
    return {
      valid: false,
      message: `Critical: Value exceeds maximum threshold of ${spec.criticalMax} ${spec.unit}`
    }
  }

  if (spec.recommendedMin && value <= spec.recommendedMin) {
    return {
      valid: true,
      message: `Warning: Value below recommended minimum of ${spec.recommendedMin} ${spec.unit}`
    }
  }

  if (spec.recommendedMax && value >= spec.recommendedMax) {
    return {
      valid: true,
      message: `Warning: Value exceeds recommended maximum of ${spec.recommendedMax} ${spec.unit}`
    }
  }

  return { valid: true }
}