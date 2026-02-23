#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PORT="8080"
pkill -f "$SCRIPT_DIR/server" 2>/dev/null || true
if command -v lsof >/dev/null 2>&1; then
    P=$(lsof -t -i ":$PORT" 2>/dev/null)
    [ -n "$P" ] && kill -9 "$P" 2>/dev/null || true
fi
echo "Stopped."
