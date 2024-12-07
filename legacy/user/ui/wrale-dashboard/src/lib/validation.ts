export function validateNumber(value: string | number, min?: number, max?: number): string | undefined {
  const num = typeof value === 'string' ? parseFloat(value) : value

  if (isNaN(num)) {
    return 'Must be a valid number'
  }

  if (min !== undefined && num < min) {
    return `Must be at least ${min}`
  }

  if (max !== undefined && num > max) {
    return `Must be no more than ${max}`
  }

  return undefined
}

export function validateIP(value: string): string | undefined {
  const pattern = /^(\d{1,3}\.){3}\d{1,3}(\/\d{1,2})?$/
  if (!pattern.test(value)) {
    return 'Must be a valid IP address'
  }

  const parts = value.split('.')
  for (const part of parts) {
    const num = parseInt(part, 10)
    if (num < 0 || num > 255) {
      return 'Each part must be between 0 and 255'
    }
  }

  return undefined
}

export function validateRequired(value: string): string | undefined {
  if (!value || value.trim().length === 0) {
    return 'This field is required'
  }
  return undefined
}