#!/bin/bash
# Cleanup: stop app, stop/remove Docker resources, remove cache/venv artifacts.
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "--- Stopping application ---"
[ -f "$SCRIPT_DIR/stop.sh" ] && bash "$SCRIPT_DIR/stop.sh" || true

echo "--- Stopping and removing Docker containers ---"
docker stop $(docker ps -aq) 2>/dev/null || true
docker rm $(docker ps -aq) 2>/dev/null || true

echo "--- Removing unused Docker resources (images, containers, networks, build cache) ---"
docker system prune -af --volumes 2>/dev/null || true

echo "--- Removing cache/artifact directories and files ---"
rm -rf node_modules .pytest_cache venv .venv __pycache__ 2>/dev/null || true
find . -type d -name "node_modules" -exec rm -rf {} + 2>/dev/null || true
find . -type d -name ".pytest_cache" -exec rm -rf {} + 2>/dev/null || true
find . -type d -name "venv" -exec rm -rf {} + 2>/dev/null || true
find . -type d -name ".venv" -exec rm -rf {} + 2>/dev/null || true
find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
find . -type f -name "*.pyc" -delete 2>/dev/null || true
find . -type f -name "*.pyo" -delete 2>/dev/null || true
for d in ./istio-*; do [ -d "$d" ] && rm -rf "$d"; done 2>/dev/null || true

echo "--- Cleanup complete ---"
