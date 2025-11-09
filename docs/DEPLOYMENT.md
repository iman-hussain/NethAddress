
# AddressIQ Deployment Guide

Deploy to Hetzner VPS with Coolify. Assumes basic familiarity with servers.

## Prerequisites

- Hetzner account and domain.
- AddressIQ GitHub access.
- API keys (see `API_REFERENCE.md`).

## Server Setup (Hetzner)

1. Create VPS (Ubuntu 22.04, CPX31 recommended).
2. SSH in: `ssh root@<ip>`.
3. Update: `apt update && apt upgrade -y`.

## Install Coolify

Run:

```bash
curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash
```

Access at `http://<ip>:8000`. Create admin account.

## Databases

In Coolify:

- PostgreSQL: Name `addressiq`, user `addressiq`, strong password.
- Redis: Default settings.

## Deploy AddressIQ

1. Add app from GitHub repo.
2. Branch: `main`.
3. Build pack: Dockerfile.

## Environment Variables

Set in Coolify (bulk add):

```bash
PORT=8080
DB_HOST=postgres
DB_PORT=5432
DB_USER=addressiq
DB_PASSWORD=<password>
DB_NAME=addressiq
REDIS_URL=redis://redis:6379
# Add API keys from [.env.example](.env.example)
```

## Build Hashes

For commit tracking:

- Build args (both services): `COMMIT_SHA=$SOURCE_COMMIT`, `BUILD_DATE=$BUILD_DATE`.
- Runtime env (backend): `FRONTEND_BUILD_COMMIT=$SOURCE_COMMIT`, `FRONTEND_BUILD_DATE=$BUILD_DATE`.

## Domain & SSL

1. Add A record to domain pointing to server IP.
2. In Coolify, add domain and enable SSL.

## Testing

- Health: `curl https://<domain>/healthz`.
- Build info: `curl https://<domain>/build-info`.
- Search: Visit domain and test address lookup.

## Troubleshooting

- App won't start: Check env vars, DB connection, logs.
- SSL issues: Verify DNS, wait 10-15 min.
- API errors: Validate keys, test locally.
- Build fails: Check Dockerfile, dependencies.

## Costs

- Server: ~€13/month.
- Domain: ~€1.25/month.
- APIs: Free/paid as needed.

## Security

- Strong passwords everywhere.
- No secrets in code/GitHub.
- Enable SSL, backups, updates.

Full docs: [Coolify](https://coolify.io/docs), [Hetzner](https://docs.hetzner.com).
