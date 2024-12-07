# Hardware Operator Guide

## Role Overview
As a Hardware Operator, you manage the day-to-day operations of physical devices in the Wrale Fleet. Your focus is on device health, performance, and immediate issue resolution.

## Primary Responsibilities
- Monitor device health and performance
- Respond to alerts and warnings
- Execute device operations
- Track environmental conditions
- Coordinate with maintenance team

## Key Interfaces

### Dashboard Access
- URL: `http://<dashboard>/devices`
- Required Role: `operator`
- Features:
  - Device status grid
  - Real-time metrics
  - Temperature monitoring
  - Power usage tracking

### Common Tasks
1. Device Health Check
   ```bash
   # View device status
   Dashboard → Devices → Status Grid
   
   # Check specific device
   Dashboard → Devices → [Device ID] → Metrics
   ```

2. Execute Operations
   ```bash
   # Run device command
   Dashboard → Devices → [Device ID] → Commands → Execute
   
   # Verify execution
   Dashboard → Devices → [Device ID] → History
   ```

3. Monitor Environment
   ```bash
   # Check temperature
   Dashboard → Analytics → Temperature Map
   
   # View power usage
   Dashboard → Analytics → Power Usage
   ```

### Alert Response
1. Alert Received:
   - Check alert severity
   - View affected device
   - Assess environmental impact

2. Initial Response:
   - Check device metrics
   - Review recent operations
   - Verify physical conditions

3. Actions:
   - Execute corrective commands
   - Update device status
   - Log intervention

## Best Practices

### Environmental Monitoring
- Regular temperature checks
- Power usage tracking
- Physical access verification
- Environmental log review

### Device Management
- Proactive health checks
- Regular metric review
- Operation verification
- Alert response logging

### Safety Protocol
- Check physical safety
- Verify operation safety
- Monitor environmental conditions
- Report safety concerns

## Common Issues

### High Temperature
1. Check device load
2. Verify cooling systems
3. Review environmental factors
4. Take corrective action

### Power Anomalies
1. Monitor power usage
2. Check power supply
3. Verify load distribution
4. Adjust as needed

### Device Unresponsive
1. Check physical connection
2. Verify network status
3. Review recent changes
4. Contact maintenance if needed

## Coordination

### With Maintenance
- Report physical issues
- Schedule maintenance
- Track repair status
- Verify post-maintenance

### With Security
- Report suspicious activity
- Monitor access logs
- Follow security protocols
- Update access records

## Documentation

### Required Logs
- Operation execution
- Alert responses
- Environmental changes
- Maintenance requests

### Reports
- Daily status
- Alert summary
- Environmental metrics
- Operation history

## Tools and Resources

### Dashboard Features
- Device Grid
- Metrics View
- Command Interface
- Alert Console

### CLI Access (if needed)
```bash
# Check device status
wrale device status <device-id>

# Execute command
wrale device exec <device-id> <command>

# View metrics
wrale device metrics <device-id>
```

## Emergency Procedures

### Critical Alerts
1. Assess severity
2. Check physical safety
3. Execute emergency protocol
4. Notify relevant teams

### System Failure
1. Ensure physical safety
2. Follow shutdown procedure
3. Contact maintenance
4. Document incident

## Getting Help
- Dashboard Documentation
- Team Chat Support
- Emergency Contacts
- Maintenance Schedule