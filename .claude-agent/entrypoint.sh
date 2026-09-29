#!/bin/sh

set -e

PROXY_PID=""

cleanup() {
    if [ -n "$PROXY_PID" ]; then
        kill "$PROXY_PID" 2>/dev/null || true
    fi
}

trap cleanup EXIT TERM INT

node /agenttest/claude-proxy.mjs &
PROXY_PID=$!

i=0

while [ "$i" -lt 50 ]; do
    if ! kill -0 "$PROXY_PID" 2>/dev/null; then
        echo "claude proxy exited unexpectedly" >&2
        exit 1
    fi

    if node -e '
        fetch("http://127.0.0.1:8080/healthz")
            .then(() => process.exit(0))
            .catch(() => process.exit(1))
    ' >/dev/null 2>&1; then
        break
    fi

    i=$((i + 1))
    sleep 0.1
done

if ! kill -0 "$PROXY_PID" 2>/dev/null; then
    echo "claude proxy is not running" >&2
    exit 1
fi

exec claude "$@"