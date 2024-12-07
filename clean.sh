#!/bin/bash
set -e

# Clean build artifacts
go clean -cache
go clean -testcache

# Ensure modules are up to date
go mod tidy

# Clean make artifacts
make clean

# Rebuild
make all