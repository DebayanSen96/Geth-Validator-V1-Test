package cmd

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

var forceRegisterCmd = &cobra.Command{
	Use:   "force-register",
	Short: "Force registration attempt without checking DXP balance",
	Run: func(cmd *cobra.Command, args []string) {
		forceRegisterVerifier()
	},
}

func init() {
	// Add force-register command to the contract command
	contractCmd.AddCommand(forceRegisterCmd)
}

// forceRegisterVerifier attempts to register without checking DXP requirements
func forceRegisterVerifier() {
	// Connect to client
	client, err := getClient()
	if err != nil {
		log.Fatalf("Error connecting to blockchain: %v", err)
	}

	// Get contract
	contract, err := getContract(client)
	if err != nil {
		log.Fatalf("Error getting contract: %v", err)
	}

	// Get account
	privateKey, address, err := getAccount()
	if err != nil {
		log.Fatalf("Error getting account: %v", err)
	}

	fmt.Printf("Account address: %s\n", address.Hex())

	// Check if the account is already registered as a verifier
	isRegistered, err := contract.IsRegistered(&bind.CallOpts{}, address)
	if err != nil {
		log.Printf("Warning: Failed to check if verifier is registered: %v", err)
	} else if isRegistered {
		fmt.Println("Account is already registered as a verifier")
		return
	}

	// Get auth options
	auth, err := getAuthOptions(client, privateKey)
	if err != nil {
		log.Fatalf("Error creating transaction options: %v", err)
	}

	// Check wallet balance
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		log.Printf("Warning: Failed to get wallet balance: %v", err)
	} else {
		// Convert wei to ether for logging
		ether := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
		fmt.Printf("Wallet balance: %s ETH\n", ether.Text('f', 6))
		
		// Check if balance is sufficient for gas
		if balance.Cmp(big.NewInt(1000000000000000)) < 0 { // 0.001 ETH minimum
			fmt.Printf("WARNING: Wallet balance may be too low for transaction fees\n")
		}
	}
	
	// Get current nonce
	nonce, err := client.PendingNonceAt(context.Background(), address)
	if err != nil {
		log.Printf("Warning: Failed to get nonce: %v", err)
	} else {
		fmt.Printf("Current nonce: %d\n", nonce)
	}
	
	// Get gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to get gas price: %v", err)
	} else {
		// Convert wei to gwei for logging
		gwei := new(big.Float).Quo(new(big.Float).SetInt(gasPrice), big.NewFloat(1e9))
		fmt.Printf("Current gas price: %s Gwei\n", gwei.Text('f', 2))
	}
	
	// Check connection to blockchain
	blockNumber, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to get block number: %v", err)
	} else {
		fmt.Printf("Connected to blockchain, current block: %d\n", blockNumber)
	}

	fmt.Println("WARNING: Bypassing DXP token checks. This transaction will likely fail on-chain.")
	fmt.Println("Forcing registration attempt...")

	// Set higher gas price to ensure transaction is picked up
	// Get suggested gas price
	suggestedGasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to get suggested gas price: %v", err)
	} else {
		// Increase gas price by 20% to ensure transaction is picked up
		multiplier := big.NewFloat(1.2)
		adjustedGasPrice := new(big.Float).Mul(new(big.Float).SetInt(suggestedGasPrice), multiplier)
		adjustedGasPriceInt, _ := adjustedGasPrice.Int(nil)
		auth.GasPrice = adjustedGasPriceInt
		
		// Convert to Gwei for logging
		gwei := new(big.Float).Quo(new(big.Float).SetInt(adjustedGasPriceInt), big.NewFloat(1e9))
		fmt.Printf("Setting gas price to: %s Gwei (increased by 20%%)\n", gwei.Text('f', 2))
	}
	
	// Ensure gas limit is sufficient
	auth.GasLimit = 300000 // Higher gas limit to ensure transaction goes through
	fmt.Printf("Setting gas limit to: %d\n", auth.GasLimit)

	tx, err := contract.RegisterValidator(auth)
	if err != nil {
		log.Fatalf("Failed to send registration transaction: %v", err)
	}

	txHash := tx.Hash().Hex()
	fmt.Printf("Transaction successfully sent to blockchain!\n")
	fmt.Printf("Transaction hash: %s\n", txHash)
	
	// Wait for transaction receipt with timeout
	fmt.Println("Waiting for transaction confirmation (this may take a minute)...")
	ctxReceipt, cancelReceipt := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelReceipt()
	
	receipt, err := bind.WaitMined(ctxReceipt, client, tx)
	if err != nil {
		fmt.Printf("Failed to get transaction receipt: %v\n", err)
		fmt.Printf("Transaction may still be pending or dropped. Check the transaction hash: %s\n", txHash)
	} else {
		// Check transaction status
		if receipt.Status == types.ReceiptStatusSuccessful { // 1 = success, 0 = failure
			fmt.Printf("Transaction confirmed successfully in block %d!\n", receipt.BlockNumber)
			fmt.Printf("Gas used: %d\n", receipt.GasUsed)
		} else {
			fmt.Printf("Transaction failed on-chain (status: 0). Check block explorer for details.\n")
			fmt.Printf("Block number: %d\n", receipt.BlockNumber)
			fmt.Printf("Gas used: %d\n", receipt.GasUsed)
		}
	}
	
	fmt.Println("\nNote: The transaction may fail on-chain due to contract requirements.")
	fmt.Println("Check the transaction status on Sepolia block explorer:")
	fmt.Printf("https://sepolia.etherscan.io/tx/%s\n", txHash)
}
