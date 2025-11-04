#!/bin/bash
# Test build hash injection locally

set -e

echo "========================================"
echo "Testing Build Hash Injection"
echo "========================================"
echo ""

# Get current commit and date
COMMIT_SHA=$(git rev-parse HEAD)
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

echo "Current commit: $COMMIT_SHA"
echo "Build date: $BUILD_DATE"
echo ""

# Build backend with build args
echo "Building backend..."
cd backend
COMMIT_HASH="$COMMIT_SHA"
BUILD_TIMESTAMP="$BUILD_DATE"
go build -ldflags "-X 'main.BuildCommit=$COMMIT_HASH' -X 'main.BuildDate=$BUILD_TIMESTAMP'" -o /tmp/addressiq-test .
cd ..

# Start backend
echo "Starting backend on port 8085..."
PORT=8085 \
FRONTEND_BUILD_COMMIT="$COMMIT_SHA" \
FRONTEND_BUILD_DATE="$BUILD_DATE" \
/tmp/addressiq-test &
BACKEND_PID=$!

# Wait for backend to start
sleep 3

# Test build-info endpoint
echo ""
echo "Testing /build-info endpoint:"
echo "========================================"
curl -s http://localhost:8085/build-info | jq .
echo ""
echo "========================================"

# Cleanup
kill $BACKEND_PID 2>/dev/null || true
rm -f /tmp/addressiq-test

echo ""
echo "✅ Test completed successfully!"
echo ""
echo "In production (Coolify), configure:"
echo "  Build Args: COMMIT_SHA=\$SOURCE_COMMIT BUILD_DATE=\$BUILD_DATE"
echo "  Env Vars: FRONTEND_BUILD_COMMIT=\$SOURCE_COMMIT FRONTEND_BUILD_DATE=\$BUILD_DATE"
