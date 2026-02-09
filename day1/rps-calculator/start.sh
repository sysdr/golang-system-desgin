#!/bin/bash
set -e

# Run from the directory containing this script (rps-calculator project root)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

PROJECT_NAME="rps-calculator"
BINARY="$SCRIPT_DIR/$PROJECT_NAME"

# Build if binary missing or source is newer
if [ ! -x "$BINARY" ] || [ main.go -nt "$BINARY" ]; then
	echo "Building $PROJECT_NAME..."
	go build -o "$PROJECT_NAME" main.go
	echo "Build complete."
fi

# Run the application for successful project output
echo "Starting $PROJECT_NAME..."
exec "$BINARY"
