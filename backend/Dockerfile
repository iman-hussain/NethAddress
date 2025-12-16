# Build stage
FROM golang:1.23-alpine AS builder

# Accept build arguments from Coolify (or use git fallback)
ARG COMMIT_SHA
ARG BUILD_DATE

# Install build dependencies in a single layer
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files first for better layer caching
# This layer is cached unless go.mod/go.sum changes
COPY go.mod go.sum ./

# Download dependencies with cache mount for faster rebuilds
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build with cache mounts for faster compilation
# Use build args if provided, otherwise fallback to git commands
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    set -ex; \
    if [ -z "$COMMIT_SHA" ]; then \
        COMMIT_HASH=$(git rev-parse HEAD 2>/dev/null || echo "unknown"); \
    else \
        COMMIT_HASH="$COMMIT_SHA"; \
    fi; \
    if [ -z "$BUILD_DATE" ]; then \
        BUILD_TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ); \
    else \
        BUILD_TIMESTAMP="$BUILD_DATE"; \
    fi; \
    echo "Building with commit: $COMMIT_HASH"; \
    echo "Building with date: $BUILD_TIMESTAMP"; \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X 'main.BuildCommit=$COMMIT_HASH' -X 'main.BuildDate=$BUILD_TIMESTAMP'" \
    -trimpath \
    -o addressiq .

# Final stage - use latest Alpine for security patches
FROM alpine:3.19

# Install runtime dependencies in a single layer
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /home/appuser

# Copy binary with correct permissions from builder
COPY --from=builder --chown=appuser:appgroup /app/addressiq ./addressiq

USER appuser

EXPOSE 8080

HEALTHCHECK NONE

CMD ["./addressiq"]