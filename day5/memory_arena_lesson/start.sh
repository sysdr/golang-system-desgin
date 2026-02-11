#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="${SCRIPT_DIR}/main"
if [ ! -x "$BINARY" ]; then
    echo "Binary not found or not executable: $BINARY. Run setup.sh first."
    exit 1
fi
if pgrep -f "${SCRIPT_DIR}/main" > /dev/null 2>&1; then
    echo "Demo already running (pgrep matched ${SCRIPT_DIR}/main). Stop first with: bash ${SCRIPT_DIR}/stop.sh"
    exit 1
fi
echo "Running demo from $BINARY..."
exec "$BINARY"
