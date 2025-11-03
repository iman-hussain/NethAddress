#!/bin/sh
set -eu

PORT="${PORT:-8080}"
URL="http://127.0.0.1:${PORT}/healthz"

if ! curl -fsS --max-time 5 "$URL" >/dev/null; then
    echo "healthcheck failed: unable to reach $URL" >&2
    exit 1
fi
