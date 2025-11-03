# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o addressiq .

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