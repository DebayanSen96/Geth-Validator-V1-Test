# Dexponent GETH-Based Validator Node

A Go Ethereum (GETH) based validator node for the Dexponent protocol. This validator allows verifiers to register, start verification processes, run a consensus mechanism, and submit proofs to smart contracts deployed on the Base chain.

## Features

- Connect to Base chain using custom RPC endpoints
- Register as a validator with the DXP smart contract
- Listen for verification requests from the blockchain
- Perform off-chain computations for complex verification tasks
- Participate in consensus with other validators
- Generate cryptographic proofs of verification results
- Submit verification results and proofs to the DXP contract
- Claim rewards for successful verifications

## Prerequisites

- Go 1.20 or higher
- Access to a Base chain RPC endpoint
- A wallet with ETH for gas fees

## Installation

```bash
# Clone the repository
git clone https://github.com/dexponent/geth-validator.git
cd geth-validator

# Install dependencies
go mod download

# Build the validator
go build -o dxp-validator
```

## Configuration

Copy the `.env.example` file to `.env` and update the values with your specific configuration:

```bash
cp .env.example .env
```

Edit the `.env` file to include your RPC provider URL, smart contract address, and wallet private key.

## Usage

```bash
# Get help
./dxp-validator --help

# Start a validator node
./dxp-validator start

# Start with custom block polling interval
./dxp-validator start --block-polling-interval 5

# Run in detached mode
./dxp-validator start --detached

# Check validator status
./dxp-validator status

# Stop a running validator
./dxp-validator stop

# Check pending rewards
./dxp-validator rewards

# Claim accumulated rewards
./dxp-validator claim
```

## Architecture

The validator node consists of several components:

1. **Validator Core**: Handles registration, block processing, and verification requests.
2. **Consensus Engine**: Manages consensus among validators to agree on verification results.
3. **Compute Engine**: Performs off-chain computations for verification tasks.
4. **Proof Generator**: Creates cryptographic proofs of verification results.

## Consensus Mechanism

The validator node implements a 2/3 majority consensus mechanism. For a verification result to be considered valid, at least 2/3 of the participating validators must agree on the result.

## Development

```bash
# Run tests
go test ./...

# Build for development
go build -tags dev -o dxp-validator
```

## Testing Contract Interaction

The repository includes a tool for testing interaction with the Dexponent Protocol contract on Sepolia testnet. This tool allows you to check registration status, register as a verifier, and submit proofs.

### Setup for Testing

1. Ensure your `.env` file is properly configured with the following values:
   ```
   BASE_RPC_URL=https://sepolia.infura.io/v3/YOUR_INFURA_KEY
   DXP_CONTRACT_ADDRESS=0x8437ab3cCb485D2a3793F97f58c6e3F926039684
   WALLET_PRIVATE_KEY=YOUR_PRIVATE_KEY_HERE
   CHAIN_ID=11155111
   ```

2. Replace `YOUR_INFURA_KEY` with your Infura API key
3. Replace `YOUR_PRIVATE_KEY_HERE` with the private key of the wallet you want to use for testing
   - **IMPORTANT**: This private key will be used to derive your wallet address
   - Never share your private key or commit it to version control
   - The account must have Sepolia ETH for gas fees and DXP tokens for staking

### Building the Test Tool

```bash
# Build the contract test tool
go build -o bin/contract-test ./cmd/test/contract.go
```

### Using the Test Tool

```bash
# Check if your account is registered as a verifier
./bin/contract-test check

# Register as a verifier (requires DXP tokens for staking)
./bin/contract-test register

# Submit proof for a farm (requires being registered first)
./bin/contract-test submit --farm-id 1 --score 100
```

### Getting Sepolia ETH and DXP Tokens

- Sepolia ETH can be obtained from faucets like:
  - [Alchemy Sepolia Faucet](https://sepoliafaucet.com/)
  - [Infura Sepolia Faucet](https://www.infura.io/faucet/sepolia)
- Contact the Dexponent team to receive test DXP tokens on Sepolia

## License

MIT
# Geth-Valdator-V1-test
