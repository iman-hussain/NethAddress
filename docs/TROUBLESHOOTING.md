# Troubleshooting Guide

Testing, diagnostics, and fixes for NethAddress.

## API Testing

Run free API test:

```bash
cd backend && go run ./cmd/test-free-apis/main.go
```

Working (4/33): BAG, Open-Meteo Weather, Open-Meteo Solar, Luchtmeetnet.
Errors (13): CBS Population, BRO Soil, etc. (outdated URLs).
Not Configured (16): Paid APIs (Kadaster, Altum, etc.).

## Common Issues

- **Hashes show "unknown"**: Set build args/env vars in Coolify.
- **Few API results**: Most APIs fail/error; check env config.
- **Frontend shows errors**: Backend APIs failing; verify URLs/keys.
- **Build fails**: Check Dockerfile, dependencies.

## Fixes

- Luchtmeetnet: Use `/open_api` endpoint.
- CBS/PDOK: Update URLs (check pdok.nl for versions).
- Paid APIs: Add keys to [.env.example](.env.example).
- Logs: Check backend for [AGGREGATOR] messages.

## Full Stack Test

Run:

```powershell
test-full-stack.ps1
```

for end-to-end validation.
