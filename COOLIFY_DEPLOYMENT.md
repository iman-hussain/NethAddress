# Coolify Deployment Guide - Build Hash Configuration

This guide explains how to configure Coolify to automatically inject Git commit hashes during the build process so they appear in the UI.

## Overview

The AddressIQ application displays build hashes for both frontend and backend in the top right corner of the website. These hashes are hyperlinks to the corresponding GitHub commits.

## Backend Configuration

### Dockerfile Build Arguments

The backend Dockerfile (`backend/Dockerfile`) accepts the following build arguments:

- `COMMIT_SHA` - The full Git commit hash
- `BUILD_DATE` - The build timestamp in ISO 8601 format

### Coolify Configuration for Backend

In Coolify, configure the backend service with the following:

1. **Build Arguments** (set these in the Coolify UI under Build settings):
   ```
   COMMIT_SHA=$SOURCE_COMMIT
   BUILD_DATE=$BUILD_DATE
   ```

2. Coolify automatically provides these variables:
   - `$SOURCE_COMMIT` - The Git commit SHA being deployed
   - `$BUILD_DATE` - The current timestamp

3. The Dockerfile will use these to inject the values via Go ldflags during compilation.

### Runtime Environment Variables for Backend

The backend also needs to know about the frontend build info. Set these environment variables in Coolify:

```bash
FRONTEND_BUILD_COMMIT=$SOURCE_COMMIT
FRONTEND_BUILD_DATE=$BUILD_DATE
```

**Note:** If you deploy frontend and backend separately, you'll need to manually set the frontend's commit hash in the backend's environment variables, or use a shared environment variable.

## Frontend Configuration

### Dockerfile Build Arguments

The frontend Dockerfile (`frontend/Dockerfile`) also accepts:

- `COMMIT_SHA` - The full Git commit hash
- `BUILD_DATE` - The build timestamp in ISO 8601 format

### Coolify Configuration for Frontend

Since the frontend is static HTML, it doesn't need the build arguments at runtime. However, the backend needs to know the frontend's build info.

**Option 1: Separate Deployments**

If frontend and backend are deployed separately:

1. Deploy frontend with build args:
   ```
   COMMIT_SHA=$SOURCE_COMMIT
   BUILD_DATE=$BUILD_DATE
   ```

2. In backend deployment, manually set:
   ```
   FRONTEND_BUILD_COMMIT=<frontend-commit-sha>
   FRONTEND_BUILD_DATE=<frontend-build-date>
   ```

**Option 2: Same Commit Deployment (Recommended)**

If both are deployed from the same commit (monorepo):

In backend deployment environment variables:
```bash
FRONTEND_BUILD_COMMIT=$SOURCE_COMMIT
FRONTEND_BUILD_DATE=$BUILD_DATE
```

This works because both services are built from the same repository commit.

## Testing the Configuration

### 1. Check Backend Build Info Endpoint

```bash
curl https://api.addressiq.imanhussain.com/build-info
```

Expected response:
```json
{
  "backend": {
    "commit": "19476a08087fe18e720ae5f9ef7c705c1d587f6a",
    "date": "2025-11-04T11:01:41Z"
  },
  "frontend": {
    "commit": "19476a08087fe18e720ae5f9ef7c705c1d587f6a",
    "date": "2025-11-04T11:01:41Z"
  }
}
```

### 2. Check Frontend Display

Open the website and look at the top right corner. You should see:

```
Frontend Build: (19476a0) Backend Build: (19476a0) | 04/11/2025 11:01
```

The hashes should be clickable links to GitHub commits.

### 3. Verify Links Work

Click on either hash - it should open:
```
https://github.com/iman-hussain/AddressIQ/commit/<full-commit-hash>
```

## Troubleshooting

### "unknown" appears for hashes

**Cause:** Build arguments are not being passed from Coolify.

**Solution:** 
1. Check Coolify build logs to verify `$SOURCE_COMMIT` is available
2. Verify build arguments are set in Coolify UI
3. For backend, check that ldflags are correctly set in Dockerfile

### Hashes are "unknown" in API but Dockerfile is correct

**Cause:** Runtime environment variables not set for frontend build info.

**Solution:** Set `FRONTEND_BUILD_COMMIT` and `FRONTEND_BUILD_DATE` in backend's runtime environment variables.

### Frontend shows backend hash but not frontend hash

**Cause:** Backend doesn't have frontend build info.

**Solution:** Set `FRONTEND_BUILD_COMMIT` and `FRONTEND_BUILD_DATE` environment variables in backend deployment.

## Coolify Build Command Examples

### Backend Build

```bash
docker build \
  --build-arg COMMIT_SHA=$SOURCE_COMMIT \
  --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t addressiq-backend:latest \
  -f backend/Dockerfile \
  ./backend
```

### Frontend Build

```bash
docker build \
  --build-arg COMMIT_SHA=$SOURCE_COMMIT \
  --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t addressiq-frontend:latest \
  -f frontend/Dockerfile \
  ./frontend
```

## Architecture Notes

1. **Backend** receives build info via:
   - Compile-time: `ldflags` inject `BuildCommit` and `BuildDate`
   - Runtime: Environment variables for frontend build info

2. **Frontend** is static HTML that:
   - Calls `/build-info` API endpoint
   - Displays both frontend and backend hashes
   - Creates clickable GitHub commit links

3. **Coolify** should provide:
   - `$SOURCE_COMMIT` - Git commit being deployed
   - Pass as build args to Docker
   - Set as environment variables for runtime

## Manual Testing Locally

### Backend
```bash
cd backend
docker build \
  --build-arg COMMIT_SHA=$(git rev-parse HEAD) \
  --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t addressiq-backend-local .

docker run -p 8080:8080 \
  -e FRONTEND_BUILD_COMMIT=$(git rev-parse HEAD) \
  -e FRONTEND_BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  addressiq-backend-local
```

### Frontend
```bash
cd frontend
docker build \
  --build-arg COMMIT_SHA=$(git rev-parse HEAD) \
  --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t addressiq-frontend-local .

docker run -p 3000:80 addressiq-frontend-local
```
