#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GATEWAY_DIR="$ROOT/gateway"

HOST="${SENTRYMESH_INTEGRATION_HOST:-127.0.0.1}"
PORT="${SENTRYMESH_INTEGRATION_PORT:-8080}"
BASE_URL="${SENTRYMESH_INTEGRATION_URL:-http://${HOST}:${PORT}}"

BINARY="${TMPDIR:-/tmp}/sentrymesh-integration"
LOG="${TMPDIR:-/tmp}/sentrymesh-integration.log"

cleanup() {
    if [[ -n "${GATEWAY_PID:-}" ]]; then
        kill "$GATEWAY_PID" 2>/dev/null || true
        wait "$GATEWAY_PID" 2>/dev/null || true
    fi
}

trap cleanup EXIT INT TERM

echo "==> Building gateway"
(
    cd "$GATEWAY_DIR"
    go build -o "$BINARY" ./cmd/sentrymesh
)

echo "==> Starting gateway"
SENTRYMESH_ROOT="$ROOT" \
"$BINARY" >"$LOG" 2>&1 &

GATEWAY_PID=$!

echo "==> Waiting for $BASE_URL/ready"

ready=0

for _ in $(seq 1 40); do
    if ! kill -0 "$GATEWAY_PID" 2>/dev/null; then
        echo "Gateway exited before becoming ready"
        cat "$LOG"
        exit 1
    fi

    if curl \
        --fail \
        --silent \
        --show-error \
        "$BASE_URL/ready" \
        >/dev/null 2>&1; then
        ready=1
        break
    fi

    sleep 0.5
done

if [[ "$ready" -ne 1 ]]; then
    echo "Gateway did not become ready"
    cat "$LOG"
    exit 1
fi

echo "==> Gateway ready"

(
    cd "$GATEWAY_DIR"

    SENTRYMESH_INTEGRATION_URL="$BASE_URL" \
        go test -tags=integration -count=1 -v ./integration
)

echo "==> Integration suite passed"
