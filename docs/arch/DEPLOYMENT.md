# Wrale Fleet Deployment Architecture

## System Components

### Container Structure
```
┌─────────────────┐     ┌─────────────────┐
│    Dashboard    │     │       API       │
│   (Next.js)    │────▶│      (Go)       │
└─────────────────┘     └────────┬────────┘
                               ▲  │
                               │  ▼
┌─────────────────┐     ┌─────────────────┐
│  Fleet Brain    │◀───▶│   Fleet Edge    │
│      (Go)      │     │      (Go)       │
└────────┬────────┘     └────────┬────────┘
         │                       │
         ▼                       ▼
┌─────────────────────────────────────────┐
│               Metal Layer               │
│                  (Go)                   │
└─────────────────────────────────────────┘
```

## Runtime Components

### Metal Runtime
- metald daemon
- Hardware management services
- Diagnostic services
- Security services

### Fleet Runtime
- fleetd brain daemon
- Edge agents
- Sync services
- State management

### User Runtime
- API service
- Next.js dashboard
- WebSocket servers
- Static assets

## Deployment Requirements

### Hardware Requirements
1. Raspberry Pi hardware
2. Networking infrastructure
3. Power infrastructure
4. Physical security

### Network Requirements
1. Inter-service connectivity
2. Edge communication
3. API access
4. WebSocket support

### Storage Requirements
1. Configuration storage
2. State persistence
3. Metrics storage
4. Log management

## Environmental Considerations

### Physical Environment
1. Temperature control
2. Power availability
3. Network connectivity
4. Physical security

### Resource Management
1. CPU allocation
2. Memory limits
3. Network bandwidth
4. Storage capacity

## Service Configuration

### Metal Configuration
1. Hardware settings
2. Security policies
3. Diagnostic settings
4. Performance tuning

### Fleet Configuration
1. Brain settings
2. Edge policies
3. Sync configuration
4. Resource limits

### User Configuration
1. API settings
2. UI configuration
3. Authentication
4. Authorization

## Monitoring & Operations

### Health Monitoring
1. Service health checks
2. Hardware monitoring
3. Resource monitoring
4. Environmental monitoring

### Metric Collection
1. Performance metrics
2. Resource usage
3. Environmental data
4. Operation logs

### Alert Management
1. Critical alerts
2. Warning notifications
3. Event monitoring
4. Status updates

## Security Configuration

### Network Security
1. Service isolation
2. Access control
3. Traffic encryption
4. Firewall rules

### Physical Security
1. Hardware security
2. Access monitoring
3. Tamper detection
4. Environmental monitoring

### Authentication
1. Service accounts
2. User authentication
3. Device authentication
4. Token management

## Disaster Recovery

### Backup Procedures
1. Configuration backup
2. State backup
3. Data retention
4. Recovery testing

### Recovery Procedures
1. Service restoration
2. State recovery
3. Configuration recovery
4. Hardware recovery

## Scaling Considerations

### Horizontal Scaling
1. Edge node addition
2. API scaling
3. Brain scaling
4. Storage scaling

### Resource Scaling
1. CPU allocation
2. Memory allocation
3. Network capacity
4. Storage capacity

## Operational Procedures

### Deployment
1. Component deployment
2. Configuration management
3. Version control
4. Rollback procedures

### Maintenance
1. Service updates
2. Configuration updates
3. Hardware maintenance
4. Security updates

### Monitoring
1. Service monitoring
2. Resource monitoring
3. Security monitoring
4. Environmental monitoring