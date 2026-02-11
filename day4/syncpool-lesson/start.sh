#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="${SCRIPT_DIR}/server"
PORT="8080"
if pgrep -f "${SCRIPT_DIR}/server" > /dev/null 2>&1 || lsof -i ":${PORT}" 2>/dev/null | grep -q LISTEN; then
    echo "Server already running. Stop first with: bash $SCRIPT_DIR/../stop.sh"
    exit 1
fi
if [ ! -x "$BINARY" ]; then
    echo "Binary not found: $BINARY. Run setup.sh first."
    exit 1
fi
echo "Starting server from $BINARY on port ${PORT}..."
cd "$SCRIPT_DIR" && exec "$BINARY"
