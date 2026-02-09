#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"
CONTAINER="heap_cost_analyzer-container"
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER}$"; then
	echo "Stopping Docker container ${CONTAINER}..."
	docker stop "$CONTAINER" 2>/dev/null || true
	docker rm "$CONTAINER" 2>/dev/null || true
fi
PIDFILE="${SCRIPT_DIR}/server.pid"
if [ -f "$PIDFILE" ]; then
	PID=$(cat "$PIDFILE")
	if kill -0 "$PID" 2>/dev/null; then
		echo "Stopping local server PID $PID..."
		kill "$PID" 2>/dev/null || true
	fi
	rm -f "$PIDFILE"
fi
echo "Stopped."
