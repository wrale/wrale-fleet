# Hardware Layer Architecture

The hardware layer (`metal/hw/`) provides direct interaction with physical Raspberry Pi hardware components. This layer emphasizes the physical-first philosophy by treating hardware interactions as first-class concerns.

## Core Components

### GPIO Controller
Primary interface for physical GPIO pin management.

- **Controller** (`gpio/controller.go`)
  - Direct pin access and control
  - Pin state management
  - Safety interlocks
  
- **Interrupt Handler** (`gpio/interrupt.go`)
  - Hardware interrupt processing
  - Event dispatch
  - Debouncing logic

- **PWM Controller** (`gpio/pwm.go`)
  - PWM signal generation
  - Duty cycle management
  - Frequency control

### Power Management
Handles power-related hardware interactions.

- **Power Manager** (`power/manager.go`)
  - Power state coordination
  - Voltage monitoring
  - Current monitoring
  - Battery management (where applicable)

- **State Handler** (`power/types.go`)
  - Power state transitions
  - Safe shutdown procedures
  - Boot sequence management

### Thermal Control
Manages temperature monitoring and cooling.

- **Thermal Monitor** (`thermal/monitor.go`)
  - Temperature sensor reading
  - Thermal zone management
  - Over-temperature protection

- **Cooling Controller** (`thermal/cooling.go`)
  - Fan speed control
  - Thermal throttling
  - Temperature prediction

### Security Control
Manages physical security aspects.

- **Security Manager** (`secure/manager.go`)
  - Physical tampering detection
  - Secure boot verification
  - Hardware encryption support

- **Monitor** (`secure/monitor.go`)
  - Continuous security monitoring
  - Threat detection
  - Alert generation

### Diagnostics
System-wide hardware diagnostics.

- **Diagnostics Manager** (`diag/manager.go`)
  - Hardware test coordination
  - Performance measurement
  - Error detection

## Hardware Interaction Patterns

### Direct Hardware Access
1. GPIO pins accessed through memory-mapped I/O
2. Hardware registers accessed through sysfs
3. Device tree overlays for specialized hardware

### Event Handling
1. Hardware interrupts processed in real-time
2. Event aggregation and dispatch
3. Priority-based event processing

### Safety Mechanisms
1. Hardware interlocks
2. Thermal protection
3. Power protection
4. Physical security monitoring

## Integration Points

### metal/core Integration
- Hardware state updates
- Command processing
- Event streaming
- Configuration management

### Hardware API
- Subsystem status reporting
- Command interface
- Configuration interface
- Diagnostic interface

## Physical Considerations

### Environmental Factors
1. Temperature monitoring and management
2. Humidity awareness (if sensors available)
3. Physical location awareness
4. Vibration monitoring (if sensors available)

### Hardware Limitations
1. GPIO capabilities and limitations
2. Power supply constraints
3. Thermal constraints
4. Physical security boundaries

## Error Handling

### Hardware Failures
1. Graceful degradation
2. Failsafe modes
3. Error reporting
4. Recovery procedures

### Safety Critical Operations
1. Pre-operation validation
2. Operation monitoring
3. Post-operation verification
4. Emergency shutdown procedures

## Testing Considerations

### Hardware Testing
1. GPIO pin testing
2. Power system testing
3. Thermal system testing
4. Security system testing

### Simulation Support
1. Hardware simulation interfaces
2. Test fixtures
3. Mocked hardware responses

## Future Considerations

1. Additional sensor support
2. Enhanced power management
3. Advanced cooling algorithms
4. Expanded diagnostic capabilities