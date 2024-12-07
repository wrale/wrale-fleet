# Contributing to Wrale Fleet

First off, thank you for considering contributing to Wrale Fleet! It's people like you that make it revolutionary fleet management possible.

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## Physical-First Philosophy

Wrale Fleet is built on a physical-first philosophy. When contributing, always consider:

1. **Physical World Implications**
   - Hardware interactions and limitations
   - Environmental conditions and constraints
   - Physical security boundaries
   - Power and thermal realities
   - Real-world deployment scenarios

2. **Safety First**
   - Hardware safety controls
   - Power management safeguards
   - Thermal protection
   - Physical security measures
   - Failsafe behaviors

3. **Environmental Awareness**
   - Temperature variations
   - Power fluctuations
   - Physical access controls
   - Network reliability
   - Environmental monitoring

## Development Process

1. Choose the right component:
   - `metal/hw/` for direct hardware interaction
   - `metal/core/` for system management
   - `metal/diag/` for diagnostics
   - `fleet/` for fleet management
   - `user/` for user interfaces
   - `shared/` for common infrastructure

2. Fork the repository

3. Create your feature branch
   ```bash
   git checkout -b feature/amazing-feature
   ```

4. Run tests for affected components
   ```bash
   # For hardware components
   cd metal/hw && go test -v ./...
   
   # For system components
   cd metal/core && go test -v ./...
   
   # For fleet management
   cd fleet && go test -v ./...
   ```

5. Ensure physical testing if modifying hardware components
   - Test with actual RPi hardware when possible
   - Verify behavior in different environmental conditions
   - Check power and thermal impacts
   - Validate physical security measures

6. Commit your changes
   ```bash
   git commit -m 'Add amazing feature'
   ```

7. Push to your branch
   ```bash
   git push origin feature/amazing-feature
   ```

8. Open a Pull Request

## Pull Request Guidelines

1. Update documentation reflecting physical implications
2. Include test coverage (both simulated and hardware when applicable)
3. Update relevant READMEs
4. Add notes about hardware requirements if any
5. Document environmental considerations

## Getting Help

- Join our community chat
- Check hardware compatibility guides
- Review physical deployment guides
- Ask questions in issues

Remember: In Wrale Fleet, hardware reality drives software architecture, not the other way around. Always start with physical constraints and build up from there.
