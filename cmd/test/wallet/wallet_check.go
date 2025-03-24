package wallet_check

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get RPC URL and private key from environment
	rpcURL := os.Getenv("BASE_RPC_URL")
	privateKeyHex := os.Getenv("WALLET_PRIVATE_KEY")
	dxpContractAddr := os.Getenv("DXP_CONTRACT_ADDRESS")

	if rpcURL == "" || privateKeyHex == "" || dxpContractAddr == "" {
		log.Fatalf("Required environment variables not set")
	}

	// Connect to Ethereum client
	fmt.Printf("Connecting to %s...\n", rpcURL)
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// Get the private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	// Get the public key and address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatalf("Failed to get public key")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("Wallet address: %s\n", address.Hex())

	// Check the balance
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	balance, err := client.BalanceAt(ctx, address, nil)
	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}

	// Convert wei to ether
	ether := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
	fmt.Printf("Balance: %s ETH\n", ether.Text('f', 6))

	// Check if the DXP contract address is valid
	dxpAddress := common.HexToAddress(dxpContractAddr)
	fmt.Printf("DXP Contract address: %s\n", dxpAddress.Hex())

	// Get the contract code to verify it exists
	code, err := client.CodeAt(ctx, dxpAddress, nil)
	if err != nil {
		log.Fatalf("Failed to get contract code: %v", err)
	}

	if len(code) == 0 {
		fmt.Printf("WARNING: No code found at the contract address. This might not be a valid contract.\n")
	} else {
		fmt.Printf("Contract verified: Code found at the contract address (length: %d bytes)\n", len(code))
	}

	// Get the latest nonce for the account
	nonce, err := client.PendingNonceAt(ctx, address)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}

	fmt.Printf("Current nonce: %d\n", nonce)

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatalf("Failed to get gas price: %v", err)
	}

	gasPriceGwei := new(big.Float).Quo(new(big.Float).SetInt(gasPrice), big.NewFloat(1e9))
	fmt.Printf("Current gas price: %s Gwei\n", gasPriceGwei.Text('f', 2))
}
