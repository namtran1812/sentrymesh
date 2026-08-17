#!/usr/bin/env bash
set -euo pipefail

VARIANT="${1:?usage: benchmark-variant.sh <full|no-log|async|no-log-no-audit>}"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GATEWAY="$ROOT/gateway"
RESULTS="$ROOT/benchmarks/results"

mkdir -p "$RESULTS"

export DATABASE_URL="${DATABASE_URL:-postgresql://sentrymesh:sentrymesh@localhost:5432/sentrymesh}"
export SENTRYMESH_BENCHMARK_MODE=1
export SENTRYMESH_ROOT="$ROOT"

unset SENTRYMESH_DISABLE_ACCESS_LOG
unset SENTRYMESH_DISABLE_AUDIT_WRITE
unset SENTRYMESH_AUDIT_MODE
unset SENTRYMESH_AUDIT_QUEUE_SIZE
unset SENTRYMESH_AUDIT_BATCH_SIZE
unset SENTRYMESH_AUDIT_FLUSH_MS

case "$VARIANT" in
  full)
    ;;
  no-log)
    export SENTRYMESH_DISABLE_ACCESS_LOG=1
    ;;
  async)
    export SENTRYMESH_DISABLE_ACCESS_LOG=1
    export SENTRYMESH_AUDIT_MODE=async
    export SENTRYMESH_AUDIT_QUEUE_SIZE=16384
    export SENTRYMESH_AUDIT_BATCH_SIZE=128
    export SENTRYMESH_AUDIT_FLUSH_MS=10
    ;;
  no-log-no-audit)
    export SENTRYMESH_DISABLE_ACCESS_LOG=1
    export SENTRYMESH_DISABLE_AUDIT_WRITE=1
    ;;
  *)
    echo "unknown variant: $VARIANT"
    exit 1
    ;;
esac

PID=""

cleanup() {
  if [[ -n "$PID" ]]; then
    kill "$PID" 2>/dev/null || true
    wait "$PID" 2>/dev/null || true
  fi
}

trap cleanup EXIT INT TERM

# Ensure a stale gateway cannot contaminate the run.
if lsof -tiTCP:8080 -sTCP:LISTEN >/dev/null 2>&1; then
  echo "port 8080 is already in use"
  lsof -nP -iTCP:8080 -sTCP:LISTEN
  exit 1
fi

cd "$GATEWAY"

go build -o /tmp/sentrymesh-bench ./cmd/sentrymesh

echo "==> starting variant: $VARIANT"

env | grep '^SENTRYMESH_' | sort

/tmp/sentrymesh-bench \
  >"$RESULTS/${VARIANT}-server.log" \
  2>&1 &

PID=$!

ready=0
for _ in $(seq 1 40); do
  if ! kill -0 "$PID" 2>/dev/null; then
    echo "gateway exited unexpectedly"
    cat "$RESULTS/${VARIANT}-server.log"
    exit 1
  fi

  if curl -fsS http://127.0.0.1:8080/ready >/dev/null; then
    ready=1
    break
  fi

  sleep 0.25
done

if [[ "$ready" -ne 1 ]]; then
  echo "gateway did not become ready"
  cat "$RESULTS/${VARIANT}-server.log"
  exit 1
fi

echo "==> gateway ready"

go run ./cmd/bench \
  -requests 5000 \
  -warmup 500 \
  -output "$RESULTS/${VARIANT}.json"

echo
echo "==> server configuration"
grep -E \
  'benchmark mode|access log|audit event writes|primary persistence' \
  "$RESULTS/${VARIANT}-server.log" \
  || true
