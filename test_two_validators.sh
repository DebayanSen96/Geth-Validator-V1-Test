#!/bin/bash

# Create temporary directories for each validator
mkdir -p ./tmp/validator1
mkdir -p ./tmp/validator2

# Create .env files for each validator
cat > ./tmp/validator1/.env << EOL
BASE_RPC_URL=https://sepolia.infura.io/v3/your-infura-key
WALLET_PRIVATE_KEY=your-private-key-1
VALIDATOR_ADDRESS=0x1111111111111111111111111111111111111111
DXP_CONTRACT_ADDRESS=0x2222222222222222222222222222222222222222
P2P_PORT=9001
P2P_PEERS=localhost:9002
EOL

cat > ./tmp/validator2/.env << EOL
BASE_RPC_URL=https://sepolia.infura.io/v3/your-infura-key
WALLET_PRIVATE_KEY=your-private-key-2
VALIDATOR_ADDRESS=0x3333333333333333333333333333333333333333
DXP_CONTRACT_ADDRESS=0x2222222222222222222222222222222222222222
P2P_PORT=9002
P2P_PEERS=localhost:9001
EOL

# Build the validator binary
echo "Building validator..."
go build -o validator

# Instructions for running the validators
echo "To run the validators, open two terminal windows and execute the following commands:"
echo ""
echo "Terminal 1:"
echo "cd $(pwd) && ./validator start --config=./tmp/validator1/.env > ./tmp/validator1/validator1.log 2>&1"
echo ""
echo "Terminal 2:"
echo "cd $(pwd) && ./validator start --config=./tmp/validator2/.env > ./tmp/validator2/validator2.log 2>&1"
echo ""
echo "To monitor the logs in real-time, open additional terminal windows and run:"
echo "tail -f ./tmp/validator1/validator1.log"
echo "tail -f ./tmp/validator2/validator2.log"

chmod +x validator
