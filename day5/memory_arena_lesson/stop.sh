#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
pkill -f "${SCRIPT_DIR}/main" 2>/dev/null || true
echo "Stopped any running demo."
