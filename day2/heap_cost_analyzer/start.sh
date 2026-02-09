#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"
BINARY="${SCRIPT_DIR}/cmd/server/server"
PORT="8080"
if pgrep -f "${BINARY}" > /dev/null 2>&1 || lsof -i ":${PORT}" > /dev/null 2>&1; then
	echo "Server already running (binary or port ${PORT} in use). Stop it first with ./stop.sh"
	exit 1
fi
if [ ! -x "$BINARY" ]; then
	echo "Binary not found: $BINARY. Run setup.sh first."
	exit 1
fi
echo "Starting server from $BINARY on port ${PORT}..."
exec "$BINARY"
