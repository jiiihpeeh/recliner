#!/bin/bash

# Script to run ReCLIner with debug output in a separate terminal window

echo "Starting ReCLIner in debug mode..."
echo "Debug output will appear in a separate terminal window."

# Start the debug client in a new terminal window
if command -v xterm &> /dev/null; then
    xterm -e "cd /home/j-p/recliner && ./cmd/debug/debug; read -p 'Press Enter to exit...'" &
elif command -v konsole &> /dev/null; then
    konsole --workdir /home/j-p/recliner --noclose -e ./cmd/debug/debug &
elif command -v gnome-terminal &> /dev/null; then
    gnome-terminal -- bash -c "cd /home/j-p/recliner && ./cmd/debug/debug; read -p 'Press Enter to exit...'" &
else
    echo "No supported terminal emulator found. Please run './cmd/debug/debug' in another terminal."
    exit 1
fi

# Small delay to let the debug client start
sleep 1

echo "Starting ReCLIner application..."
# Run ReCLIner with debug mode
cd /home/j-p/recliner
./recliner --debug