#!/bin/bash
# Diagnostic script for Konsole

echo "=== Terminal Environment ===" 
echo "TERM: $TERM"
echo "TERM_PROGRAM: $TERM_PROGRAM"
echo ""

echo "=== Stdin Info ===" 
if [ -t 0 ]; then
    echo "Stdin is a TTY"
else
    echo "Stdin is NOT a TTY"
fi
echo ""

echo "=== Running App ===" 
cd /home/j-p/recliner
./recliner 2>&1 | head -20
echo ""
echo "App exit code: $?"
