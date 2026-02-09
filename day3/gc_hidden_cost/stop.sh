#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_FILE="${SCRIPT_DIR}/gc_hidden_cost.pid"
if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE" 2>/dev/null)
    if [ -n "$PID" ] && kill -0 "$PID" 2>/dev/null; then
        echo "Stopping server PID $PID..."
        kill "$PID" 2>/dev/null || true
    fi
    rm -f "$PID_FILE"
fi
pkill -f "gc_hidden_cost.out" 2>/dev/null || true
docker stop gc_hidden_cost 2>/dev/null || true
docker rm gc_hidden_cost 2>/dev/null || true
echo "Stopped."
