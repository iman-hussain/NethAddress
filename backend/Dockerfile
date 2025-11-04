# Build stage
FROM golang:1.23-alpine AS builder

# Accept build arguments from Coolify (or use git fallback)
ARG COMMIT_SHA
ARG BUILD_DATE

RUN apk add --no-cache git
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Use build args if provided, otherwise fallback to git commands
RUN COMMIT_HASH=${COMMIT_SHA:-$(git rev-parse HEAD 2>/dev/null || echo "unknown")} && \
    BUILD_TIMESTAMP=${BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)} && \
    echo "Building with commit: $COMMIT_HASH" && \
    echo "Building with date: $BUILD_TIMESTAMP" && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X 'main.BuildCommit=$COMMIT_HASH' -X 'main.BuildDate=$BUILD_TIMESTAMP'" \
    -o addressiq .

# Final stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates curl wget

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /home/appuser

COPY --from=builder /app/addressiq ./addressiq
RUN chown appuser:appgroup ./addressiq && chmod +x ./addressiq

USER appuser

EXPOSE 8080

HEALTHCHECK NONE

CMD ["./addressiq"]