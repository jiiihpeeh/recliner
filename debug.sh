#!/bin/bash
# Debug script to capture app output

# Build the app
go build -o recliner main.go

# Run with output redirection
echo "Starting app - output will be saved to debug.log"
./recliner > debug.log 2>&1 &
APP_PID=$!

# Wait a bit for the app to potentially fail
sleep 3

# Check if app is still running
if ps -p $APP_PID > /dev/null; then
    echo "App is still running (PID: $APP_PID)"
    kill $APP_PID 2>/dev/null || true
else
    echo "App exited (likely crashed)"
fi

# Show the debug output
echo ""
echo "=== App Output ==="
cat debug.log
