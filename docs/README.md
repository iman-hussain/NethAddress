# NethAddress Documentation

Welcome to the NethAddress documentation wiki. This repository contains comprehensive guides and references for working with the NethAddress property intelligence toolkit.

## Documentation Index

### [API Reference](API_REFERENCE.md)

Complete reference for all external API integrations used by NethAddress. Includes:

- List of 35+ integrated APIs (cadastral, demographic, environmental, market data)
- Environment variable configuration for each API
- Authentication requirements and pricing information
- REST API endpoints for querying property data

### [Deployment Guide](DEPLOYMENT.md)

Production deployment instructions and configuration. Covers:

- Coolify deployment with automatic build hash injection
- Docker and docker-compose setup
- Environment variable configuration
- Build arguments (COMMIT_SHA, BUILD_DATE)
- Frontend/backend build info integration

### [CBS Reference](CBS_REFERENCE.md)

Netherlands Statistics (CBS) dataset documentation. Includes:

- CBS OData API integration details
- Field definitions and data structure
- Statistical tables and neighbourhood data
- JSON metadata snapshots (`cbs_fields.json`, `cbs_metadata.json`)

### [Troubleshooting Guide](TROUBLESHOOTING.md)

Testing and debugging resources. Contains:

- Backend test commands (`go test`)
- Free API testing scripts
- Full-stack integration test procedures
- Common issues and solutions

## Quick Links

- [Main README](../README.md) - Project overview and quick start
- [Backend Source](../backend/) - Go backend implementation
- [Frontend Source](../frontend/) - HTML/JS frontend

## Contributing

When updating documentation:

1. Keep content concise and actionable
2. Use proper markdown formatting and follow markdownlint rules
3. Update this index when adding new documentation files
4. Use relative links for cross-references
