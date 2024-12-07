// Physical subsystems that map directly to hardware components
export type PowerSubsystem = {
  type: 'power'
  voltage: {
    current: number    // Current voltage reading
    min: number       // Minimum acceptable voltage
    max: number       // Maximum acceptable voltage
    nominal: number   // Nominal (expected) voltage
    ripple?: number   // Voltage ripple measurement if available
  }
  current: {
    current: number   // Current amperage reading
    max: number      // Maximum rated current
    typical: number  // Typical operating current
  }
  efficiency: {
    current: number  // Current efficiency percentage
    target: number  // Target efficiency
  }
  protection: {
    overcurrent: boolean
    overvoltage: boolean
    undervoltage: boolean
    thermal: boolean
  }
  status: 'nominal' | 'warning' | 'critical'
}

export type ThermalSubsystem = {
  type: 'thermal'
  temperature: {
    current: number     // Current temperature
    max: number        // Maximum rated temperature
    critical: number   // Critical shutdown temperature
    gradient?: number  // Temperature change rate if available
  }
  cooling: {
    type: 'passive' | 'active'
    fanSpeed?: number  // RPM if active cooling
    fanStatus?: 'ok' | 'warning' | 'failed'
    airflow?: number   // CFM if measurable
  }
  zones: {
    id: string
    name: string
    temperature: number
    maxTemp: number
  }[]
  status: 'nominal' | 'warning' | 'critical'
}

export type StorageSubsystem = {
  type: 'storage'
  device: {
    type: 'sd' | 'emmc' | 'nvme'
    size: number      // Total size in bytes
    manufacturer: string
    model: string
  }
  health: {
    lifeRemaining: number  // Percentage
    readErrors: number
    writeErrors: number
    badBlocks: number
  }
  performance: {
    readSpeed: number    // Current read speed in MB/s
    writeSpeed: number   // Current write speed in MB/s
    iops: number        // Current IOPS
  }
  status: 'nominal' | 'warning' | 'critical'
}

export type NetworkSubsystem = {
  type: 'network'
  interfaces: {
    name: string
    type: 'ethernet' | 'wifi' | 'bluetooth'
    status: 'up' | 'down'
    metrics: {
      rxRate: number     // Current receive rate in bps
      txRate: number     // Current transmit rate in bps
      latency: number    // Current latency in ms
      errors: number     // Error count
      drops: number      // Drop count
    }
  }[]
  status: 'nominal' | 'warning' | 'critical'
}

// Physical location and environment information
export type PhysicalContext = {
  location: {
    rack: string
    unit: number
    position: {
      x: number
      y: number
      z: number
    }
  }
  environment: {
    temperature: number    // Ambient temperature
    humidity: number      // Relative humidity
    airflow: number       // Air flow rate in CFM
    noise: number         // Noise level in dB
    vibration?: number    // Vibration level if sensor available
  }
  neighbors: {
    above?: string    // Device ID above
    below?: string    // Device ID below
    left?: string     // Device ID to left
    right?: string    // Device ID to right
  }
}

// Complete physical device representation
export type PhysicalDevice = {
  id: string
  model: string
  serialNumber: string
  manufacturingDate: string
  physicalContext: PhysicalContext
  subsystems: {
    power: PowerSubsystem
    thermal: ThermalSubsystem
    storage: StorageSubsystem
    network: NetworkSubsystem
  }
  dimensions: {
    height: number    // mm
    width: number     // mm
    depth: number     // mm
    weight: number    // g
  }
  maintenance: {
    lastService: string
    nextService: string
    serviceHistory: {
      date: string
      type: string
      notes: string
    }[]
    healthScore: number  // 0-100
    alerts: {
      subsystem: string
      severity: 'warning' | 'critical'
      message: string
      timestamp: string
    }[]
  }
}