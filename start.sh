#!/bin/sh

if [ -z "$SHELLED" ]; then
    export SHELLED=1
    exec "$SHELL" "$0" "$@"
fi

if ! command -v go >/dev/null 2>&1; then
    echo "Error: Go is not installed. Install Go and try again."
    exit 1
fi

if [ ! -f "go.mod" ]; then
    echo "Error: go.mod not found. Run go mod init."
    exit 1
fi

if [ ! -f "go.sum" ] || [ ! -d "vendor" ]; then
    echo "Installing/updating dependencies.."
    go mod tidy
fi

echo "Starting lark-api.."
nohup go run . > /dev/null 2>&1 &
