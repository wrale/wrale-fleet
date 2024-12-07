export type AlertSeverity = 'low' | 'medium' | 'high'
export type AlertChannelType = 'email' | 'slack' | 'webhook'
export type AlertAction = 'restart' | 'ticket' | 'notification'

export interface AlertChannel {
  id: string
  type: AlertChannelType
  name: string
  config: {
    recipients?: string[]
    webhook_url?: string
    channel?: string
  }
  enabled: boolean
}

export interface AlertRule {
  id: string
  name: string
  condition: string
  severity: AlertSeverity
  channels: string[]
  actions?: AlertAction[]
  enabled: boolean
}