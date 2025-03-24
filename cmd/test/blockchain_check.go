package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get RPC URL from environment
	rpcURL := os.Getenv("BASE_RPC_URL")
	if rpcURL == "" {
		log.Fatalf("BASE_RPC_URL environment variable not set")
	}

	// Connect to Ethereum client
	fmt.Printf("Connecting to %s...\n", rpcURL)
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// Check connection by getting the latest block number
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	blockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		log.Fatalf("Failed to get latest block number: %v", err)
	}
	fmt.Printf("Successfully connected to blockchain. Current block number: %d\n", blockNumber)

	// Check the transaction
	txHash := common.HexToHash("0x4607f5e64877651340317f39efddf17a5d1ae3c34189ad4c2d429b7fc3c7dc1e")
	fmt.Printf("Checking transaction: %s\n", txHash.Hex())

	tx, isPending, err := client.TransactionByHash(ctx, txHash)
	if err != nil {
		fmt.Printf("Error retrieving transaction: %v\n", err)
		return
	}

	fmt.Printf("Transaction found!\n")
	fmt.Printf("Is pending: %v\n", isPending)
	fmt.Printf("Gas: %d\n", tx.Gas())
	fmt.Printf("Gas price: %s\n", tx.GasPrice().String())
	fmt.Printf("Value: %s\n", tx.Value().String())
	fmt.Printf("Nonce: %d\n", tx.Nonce())
	fmt.Printf("To: %s\n", tx.To().Hex())

	// Get transaction receipt
	receipt, err := client.TransactionReceipt(ctx, txHash)
	if err != nil {
		fmt.Printf("Error retrieving transaction receipt: %v\n", err)
		return
	}

	fmt.Printf("Transaction receipt:\n")
	fmt.Printf("Status: %d (0=failed, 1=success)\n", receipt.Status)
	fmt.Printf("Block number: %d\n", receipt.BlockNumber)
	fmt.Printf("Gas used: %d\n", receipt.GasUsed)
}
