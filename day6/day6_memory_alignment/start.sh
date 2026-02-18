#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="${SCRIPT_DIR}/server"
PORT="8080"
if pgrep -f "${SCRIPT_DIR}/server" > /dev/null 2>&1 || (command -v lsof >/dev/null 2>&1 && lsof -i ":${PORT}" 2>/dev/null | grep -q LISTEN); then
    echo "Server already running. Stop first with: bash $SCRIPT_DIR/stop.sh"
    exit 1
fi
if [ ! -f "$BINARY" ]; then
    echo "Building..."
    (cd "$SCRIPT_DIR" && go build -o server .)
fi
[ -x "$BINARY" ] || chmod +x "$BINARY" 2>/dev/null || true
echo "Starting server from $SCRIPT_DIR on port ${PORT}..."
cd "$SCRIPT_DIR" && exec "$BINARY"
