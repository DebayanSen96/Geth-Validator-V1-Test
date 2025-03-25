#!/bin/bash

# Build the validator
go build -o validator ./cmd/start.go

# Create temporary directories for each validator
mkdir -p ./tmp/validator1
mkdir -p ./tmp/validator2
mkdir -p ./tmp/validator3

# Create .env files for each validator with different ports
cat > ./tmp/validator1/.env << EOL
BASE_RPC_URL=https://sepolia.base.org
WALLET_PRIVATE_KEY=0x1111111111111111111111111111111111111111111111111111111111111111
VALIDATOR_ADDRESS=0x1111111111111111111111111111111111111111
DXP_CONTRACT_ADDRESS=0x2222222222222222222222222222222222222222
P2P_PORT=8881
P2P_PEERS=localhost:8882,localhost:8883
EOL

cat > ./tmp/validator2/.env << EOL
BASE_RPC_URL=https://sepolia.base.org
WALLET_PRIVATE_KEY=0x2222222222222222222222222222222222222222222222222222222222222222
VALIDATOR_ADDRESS=0x2222222222222222222222222222222222222222
DXP_CONTRACT_ADDRESS=0x2222222222222222222222222222222222222222
P2P_PORT=8882
P2P_PEERS=localhost:8881,localhost:8883
EOL

cat > ./tmp/validator3/.env << EOL
BASE_RPC_URL=https://sepolia.base.org
WALLET_PRIVATE_KEY=0x3333333333333333333333333333333333333333333333333333333333333333
VALIDATOR_ADDRESS=0x3333333333333333333333333333333333333333
DXP_CONTRACT_ADDRESS=0x2222222222222222222222222222222222222222
P2P_PORT=8883
P2P_PEERS=localhost:8881,localhost:8882
EOL

# Start validators in separate terminals
echo "Starting validator 1..."
cd ./tmp/validator1 && ../../validator start > validator1.log 2>&1 &
VALIDATOR1_PID=$!

echo "Starting validator 2..."
cd ./tmp/validator2 && ../../validator start > validator2.log 2>&1 &
VALIDATOR2_PID=$!

echo "Starting validator 3..."
cd ./tmp/validator3 && ../../validator start > validator3.log 2>&1 &
VALIDATOR3_PID=$!

echo "All validators started. Press Ctrl+C to stop all validators."

# Wait for user to press Ctrl+C
trap "kill $VALIDATOR1_PID $VALIDATOR2_PID $VALIDATOR3_PID; echo 'Validators stopped.'; exit 0" INT
wait
