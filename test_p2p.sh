#!/bin/bash

# Test script to demonstrate P2P communication between two validator instances

echo "Starting P2P communication test between two validators"

# Kill any existing validator processes
pkill -f dxp-validator

# Start the first validator in the background
echo "Starting validator 1..."
TEST_MODE=true P2P_PORT=30301 ./dxp-validator start --config=.env.validator1 > validator1.log 2>&1 &
PID1=$!

# Give it a moment to start
sleep 2

# Start the second validator in the background
echo "Starting validator 2..."
TEST_MODE=true P2P_PORT=30302 ./dxp-validator start --config=.env.validator2 > validator2.log 2>&1 &
PID2=$!

# Wait for both validators to initialize and connect
echo "Waiting for validators to connect..."
sleep 5

# Display logs from both validators
echo "\nValidator 1 log:"
tail -n 20 validator1.log

echo "\nValidator 2 log:"
tail -n 20 validator2.log

# Keep the script running to observe the validators
echo "\nTest running. Press Ctrl+C to stop the test and kill the validators."

# Wait for user to press Ctrl+C
trap "kill $PID1 $PID2; echo 'Test stopped.'; exit 0" INT
while true; do
  sleep 1
done
