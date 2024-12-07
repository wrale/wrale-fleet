# Wrale Fleet Metal Hardware

[![Go](https://github.com/wrale/wrale-fleet/actions/workflows/metal-hw.yml/badge.svg)](https://github.com/wrale/wrale-fleet/actions/workflows/metal-hw.yml)
[![Lint](https://github.com/wrale/wrale-fleet/actions/workflows/metal-hw.yml/badge.svg)](https://github.com/wrale/wrale-fleet/actions/workflows/metal-hw.yml)

Pure hardware management layer for Wrale Fleet. Handles direct hardware interactions, raw sensor data, and hardware-level safety for Raspberry Pi devices. Part of the Wrale Fleet Metal project.

## Make Targets

The hardware layer provides the following make targets:

### Primary Targets
- `make all` - Build everything and run verifications
- `make build` - Build hardware components
- `make clean` - Clean build artifacts
- `make test` - Run all tests
- `make verify` - Run all verifications

### Hardware-Specific Targets
- `make hardware-test` - Run hardware-specific tests
- `make simulation` - Run in simulation mode
- `make calibrate` - Run sensor calibration

Run `make help` for a complete list of available targets.

## Feature Status

[Previous feature status table content...]

[Rest of the README content remains unchanged...]