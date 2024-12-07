# Contributing to Wrale Fleet v1.0

This document extends CONTRIBUTING.md with v1.0-specific processes and requirements.

## v1.0 Development Workflow

### Component Status
- Metal Layer: Feature complete
- Fleet Layer: Feature complete
- User Layer: Feature complete
- Integration: Feature complete

### Development Process
1. Focus on bug fixes and stability
2. Critical security issues only
3. Documentation improvements
4. No new features until v1.1

### Branch Strategy
```
main
 │
 ├── release/v1.0
 │    ├── hotfix/security-fix
 │    └── hotfix/critical-bug
 │
 └── develop
      ├── feature/v1.1-feature
      └── feature/v1.1-enhancement
```

## v1.0 Testing Requirements

### Unit Tests
- Metal: 90% coverage
- Fleet: 85% coverage
- API: 80% coverage
- UI: Component tests

### Integration Tests
- Metal → Fleet
- Fleet Brain → Edge
- API → Fleet
- UI → API

### System Tests
- Full deployment
- Stress testing
- Security scanning
- UI/UX testing

### Hardware Tests
```
Required Hardware:
- Raspberry Pi 5
- GPIO peripherals
- Temperature sensors
- Power monitoring
```

## v1.0 Release Process

### Pre-Release
1. Version bump
2. Changelog update
3. Documentation review
4. Test suite completion

### Release Build
1. Clean build environment
2. Run all tests
3. Build containers
4. Tag release

### Deployment
1. Environment validation
2. Container deployment
3. Health verification
4. Monitoring setup

### Post-Release
1. Migration validation
2. Performance monitoring
3. Error tracking
4. User feedback

## v1.0 Issue Management

### Bug Reports
```yaml
Title: [Component] Brief description
Body:
  - Environment: [prod/dev]
  - Version: [exact version]
  - Hardware: [device details]
  - Steps to reproduce
  - Expected behavior
  - Actual behavior
  - Logs/Screenshots
```

### Feature Requests
- Defer to v1.1
- Critical enhancements only
- Security improvements
- Documentation updates

## v1.0 Code Review

### Review Focus
1. Security implications
2. Hardware safety
3. Performance impact
4. Error handling

### Review Checklist
```
□ Physical safety checks
□ Error handling complete
□ Tests included
□ Documentation updated
□ Performance validated
□ Security reviewed
```

## v1.0 Documentation

### Required Updates
- README version info
- Setup instructions
- Configuration guide
- Troubleshooting steps

### API Documentation
- Endpoint details
- Request/Response formats
- Error codes
- Examples

### Architecture Updates
- Deployment diagrams
- Integration flows
- Security model
- Error handling

## Getting Help

### v1.0 Support
- GitHub Issues
- Documentation
- Community chat
- Email support

### Emergency Support
- Security issues
- Hardware safety
- Data loss
- System failure

## Future Development

### v1.1 Planning
- Feature proposals
- Architecture changes
- Performance improvements
- Security enhancements

### Contributing Path
1. Bug reports welcome
2. Documentation help
3. Testing assistance
4. Feature ideas for v1.1