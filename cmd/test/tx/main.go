package main

import (
	"context"
	"fmt"
	"log"
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

	// Connect to Ethereum client
	client, err := ethclient.Dial("https://sepolia.infura.io/v3/713cd361276943ec9d192bf3ce177f4c")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// Define transaction hashes to check
	txHashes := []string{
		"0x7353953bb4e8dbf0cb3bc8e60f79feceb20b848891c77992c5b738309d6d31b6",
		"0xd62ba79c9ae786c9bc5d130b092c0b6db248d8957d958cd35e721b7587a27020",
		"0x4607f5e64877651340317f39efddf17a5d1ae3c34189ad4c2d429b7fc3c7dc1e",
		"0x6a0ecea41b11098ced0f451ef055f026272cb9fff090db7750eaae594983e936",
		"0x1874f68f9fda32f33fb9f9ce31b24dd819ba9706275c6423966760d0fcf92b28",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check each transaction
	for _, hashStr := range txHashes {
		txHash := common.HexToHash(hashStr)
		fmt.Printf("\nChecking transaction: %s\n", txHash.Hex())

		tx, isPending, err := client.TransactionByHash(ctx, txHash)
		if err != nil {
			fmt.Printf("Error retrieving transaction: %v\n", err)
			continue
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
			continue
		}

		fmt.Printf("Transaction receipt:\n")
		fmt.Printf("Status: %d (0=failed, 1=success)\n", receipt.Status)
		fmt.Printf("Block number: %d\n", receipt.BlockNumber)
		fmt.Printf("Gas used: %d\n", receipt.GasUsed)
	}
}
