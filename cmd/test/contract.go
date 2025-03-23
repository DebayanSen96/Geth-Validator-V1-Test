package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/dexponent/geth-validator/internal/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	farmID          int64
	performanceScore int64
)

var rootCmd = &cobra.Command{
	Use:   "contract-test",
	Short: "Test interaction with the DXP contract",
	Long:  "Test interaction with the Dexponent Protocol contract on Sepolia testnet",
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if the account is registered as a verifier",
	Run: func(cmd *cobra.Command, args []string) {
		checkRegistration()
	},
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register as a verifier",
	Run: func(cmd *cobra.Command, args []string) {
		registerVerifier()
	},
}

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit proof to the contract",
	Run: func(cmd *cobra.Command, args []string) {
		submitProof()
	},
}

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
	}

	// Add subcommands
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(submitCmd)

	// Add flags
	submitCmd.Flags().Int64VarP(&farmID, "farm-id", "f", 1, "Farm ID to submit proof for")
	submitCmd.Flags().Int64VarP(&performanceScore, "score", "s", 100, "Performance score to submit")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getClient() (*ethclient.Client, error) {
	rpcURL := os.Getenv("BASE_RPC_URL")
	if rpcURL == "" {
		return nil, fmt.Errorf("BASE_RPC_URL not set in .env file")
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}

	return client, nil
}

func getContract(client *ethclient.Client) (*contracts.DexponentContractWrapper, error) {
	contractAddr := os.Getenv("DXP_CONTRACT_ADDRESS")
	if contractAddr == "" {
		return nil, fmt.Errorf("DXP_CONTRACT_ADDRESS not set in .env file")
	}

	contractAddress := common.HexToAddress(contractAddr)
	contract, err := contracts.NewDexponentContractWrapper(contractAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate contract: %v", err)
	}

	return contract, nil
}

func getAccount() (*ecdsa.PrivateKey, common.Address, error) {
	privateKeyStr := os.Getenv("WALLET_PRIVATE_KEY")
	if privateKeyStr == "" {
		return nil, common.Address{}, fmt.Errorf("WALLET_PRIVATE_KEY not set in .env file")
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("invalid private key: %v", err)
	}

	// Get account address from private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, common.Address{}, fmt.Errorf("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return privateKey, address, nil
}

func getAuthOptions(client *ethclient.Client, privateKey *ecdsa.PrivateKey) (*bind.TransactOpts, error) {
	chainIDStr := os.Getenv("CHAIN_ID")
	if chainIDStr == "" {
		return nil, fmt.Errorf("CHAIN_ID not set in .env file")
	}

	// Default to Sepolia chain ID
	chainID := big.NewInt(11155111)

	// Create transaction options
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction options: %v", err)
	}

	// Set gas price and limit
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %v", err)
	}

	auth.GasPrice = gasPrice
	auth.GasLimit = 3000000

	return auth, nil
}

// formatEther converts wei to ether
func formatEther(wei *big.Int) string {
	ether := new(big.Float).Quo(
		new(big.Float).SetInt(wei),
		new(big.Float).SetInt(big.NewInt(1e18)),
	)
	return ether.String()
}

func checkRegistration() {
	// Connect to client
	client, err := getClient()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Get contract
	contract, err := getContract(client)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Get account
	_, address, err := getAccount()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Account address: %s\n", address.Hex())

	// Check if the account is already registered as a verifier
	isRegistered, err := contract.IsRegistered(&bind.CallOpts{}, address)
	if err != nil {
		log.Fatalf("Failed to check if verifier is registered: %v", err)
	}

	fmt.Printf("Is account registered as verifier: %v\n", isRegistered)

	// Get account balance
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		log.Fatalf("Failed to get account balance: %v", err)
	}

	fmt.Printf("Account balance: %s ETH\n", formatEther(balance))
}

func registerVerifier() {
	// Connect to client
	client, err := getClient()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Get contract
	contract, err := getContract(client)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Get account
	privateKey, address, err := getAccount()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Account address: %s\n", address.Hex())

	// Check if the account is already registered as a verifier
	isRegistered, err := contract.IsRegistered(&bind.CallOpts{}, address)
	if err != nil {
		log.Fatalf("Failed to check if verifier is registered: %v", err)
	}

	if isRegistered {
		fmt.Println("Account is already registered as a verifier")
		return
	}

	// Get auth options
	auth, err := getAuthOptions(client, privateKey)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Note: Registration requires DXP tokens for staking. Make sure your account has approved the contract to spend DXP tokens.")
	fmt.Println("Attempting to register as verifier...")

	tx, err := contract.RegisterValidator(auth)
	if err != nil {
		log.Fatalf("Failed to register as verifier: %v", err)
	}

	fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())
	fmt.Println("Check the transaction status on Sepolia block explorer")
}

func submitProof() {
	// Connect to client
	client, err := getClient()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Get contract
	contract, err := getContract(client)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Get account
	privateKey, address, err := getAccount()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Account address: %s\n", address.Hex())

	// Check if the account is already registered as a verifier
	isRegistered, err := contract.IsRegistered(&bind.CallOpts{}, address)
	if err != nil {
		log.Fatalf("Failed to check if verifier is registered: %v", err)
	}

	if !isRegistered {
		fmt.Println("Account is not registered as a verifier. Please register first.")
		return
	}

	// Get auth options
	auth, err := getAuthOptions(client, privateKey)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Attempting to submit proof for farm ID %d with performance score %d...\n", farmID, performanceScore)

	// For the wrapper, we need to use SubmitVerificationResult
	// The wrapper will convert this to a submitProof call
	tx, err := contract.SubmitVerificationResult(auth, big.NewInt(farmID), []byte{}, []byte{})
	if err != nil {
		log.Fatalf("Failed to submit proof: %v", err)
	}

	fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())
	fmt.Println("Check the transaction status on Sepolia block explorer")
}
