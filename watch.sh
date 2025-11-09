#!/bin/bash

# Watch script for dessertfrog - rebuilds and reruns on changes
# Usage: ./watch.sh [program arguments]
# Example: ./watch.sh -d postgres -P mypassword

BUILD_DIR="."
BINARY="$BUILD_DIR/dessertfrog"
PID_FILE="/tmp/dessertfrog.pid"

# Cleanup function
cleanup() {
    echo "Cleaning up..."
    if [ -f "$PID_FILE" ]; then
        kill $(cat "$PID_FILE") 2>/dev/null
        rm -f "$PID_FILE"
    fi
    exit 0
}

# Trap CTRL+C and cleanup
trap cleanup SIGINT SIGTERM

echo "Starting dessertfrog watch mode..."
echo "Press CTRL+C to stop"
echo "Program arguments: $@"
echo ""

while true; do
    echo "[$(date '+%H:%M:%S')] Building..."

    # Build the project
    cd "$BUILD_DIR"
    if go build -o "$BINARY" . 2>&1 ; then
        echo "[$(date '+%H:%M:%S')] Build successful"

        # Kill previous instance if running
        if [ -f "$PID_FILE" ]; then
            OLD_PID=$(cat "$PID_FILE")
            if kill -0 "$OLD_PID" 2>/dev/null; then
                kill "$OLD_PID" 2>/dev/null
                sleep 0.2
            fi
        fi

        # Run the new binary in background
        "$BINARY" "$@" &
        NEW_PID=$!
        echo "$NEW_PID" > "$PID_FILE"
        echo "[$(date '+%H:%M:%S')] Running (PID: $NEW_PID)"
    else
        echo "[$(date '+%H:%M:%S')] Build failed"
    fi

    # Wait 3 seconds before next rebuild
    sleep 3
done
