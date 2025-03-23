package cmd

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/dexponent/geth-validator/internal/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

var (
	farmID          int64
	performanceScore int64
	approvalAmount   int64
	dxpTokenAddress string = "0x4ed4E862860beD51a9570b96d89aF5E1B0Efefed" // Replace with actual DXP token address
)

// Contract commands
var contractCmd = &cobra.Command{
	Use:   "contract",
	Short: "Interact with the Dexponent Protocol contract",
	Long:  "Commands for interacting with the Dexponent Protocol contract on Sepolia testnet",
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if the account is registered as a verifier",
	Run: func(cmd *cobra.Command, args []string) {
		checkRegistration()
	},
}

var checkDXPCmd = &cobra.Command{
	Use:   "check-dxp",
	Short: "Check DXP token balance and approval status",
	Run: func(cmd *cobra.Command, args []string) {
		checkDXPTokens()
	},
}

var approveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve DXP token spending for the contract",
	Run: func(cmd *cobra.Command, args []string) {
		approveDXPTokens()
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
	// Add contract command to the root command
	RootCmd.AddCommand(contractCmd)

	// Add subcommands to the contract command
	contractCmd.AddCommand(checkCmd)
	contractCmd.AddCommand(checkDXPCmd)
	contractCmd.AddCommand(approveCmd)
	contractCmd.AddCommand(registerCmd)
	contractCmd.AddCommand(submitCmd)

	// Add flags
	submitCmd.Flags().Int64VarP(&farmID, "farm-id", "f", 1, "Farm ID to submit proof for")
	submitCmd.Flags().Int64VarP(&performanceScore, "score", "s", 100, "Performance score to submit")
	approveCmd.Flags().Int64VarP(&approvalAmount, "amount", "a", 1000, "Amount of DXP tokens to approve (in tokens, not wei)")
}

// getClient establishes a connection to the Ethereum client
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

// getContract creates an instance of the Dexponent contract
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

// getAccount retrieves the account from the private key
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

// getAuthOptions creates transaction options for contract interactions
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

// checkRegistration checks if the account is registered as a verifier
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

	// Check if the account is registered as a verifier
	isRegistered, err := contract.IsRegistered(&bind.CallOpts{}, address)
	if err != nil {
		log.Fatalf("Failed to check if verifier is registered: %v", err)
	}

	if isRegistered {
		fmt.Println("Account is registered as a verifier")
	} else {
		fmt.Println("Account is not registered as a verifier")
	}

	// Check account balance
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		log.Fatalf("Failed to get account balance: %v", err)
	}

	fmt.Printf("Account balance: %s ETH\n", formatEther(balance))
}

// checkDXPRequirements checks if the address has enough DXP tokens and has approved the contract
func checkDXPRequirements(client *ethclient.Client, address common.Address) (bool, error) {
	// Get DXP contract address
	contractAddr := os.Getenv("DXP_CONTRACT_ADDRESS")
	if contractAddr == "" {
		return false, fmt.Errorf("DXP_CONTRACT_ADDRESS not set in .env file")
	}

	// Create DXP token contract instance
	tokenAddress := common.HexToAddress(dxpTokenAddress)
	tokenABI := `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`
	parsedABI, err := abi.JSON(strings.NewReader(tokenABI))
	if err != nil {
		return false, fmt.Errorf("failed to parse token ABI: %v", err)
	}
	tokenContract := bind.NewBoundContract(tokenAddress, parsedABI, client, client, client)

	// Check DXP token balance
	var balance *big.Int
	balanceResult := []interface{}{&balance}
	err = tokenContract.Call(&bind.CallOpts{}, &balanceResult, "balanceOf", address)
	if err != nil {
		return false, fmt.Errorf("failed to get DXP token balance: %v", err)
	}

	// Check allowance
	var allowance *big.Int
	allowanceResult := []interface{}{&allowance}
	contractAddress := common.HexToAddress(contractAddr)
	err = tokenContract.Call(&bind.CallOpts{}, &allowanceResult, "allowance", address, contractAddress)
	if err != nil {
		return false, fmt.Errorf("failed to get allowance: %v", err)
	}

	// Check if balance and allowance are sufficient for registration
	minStake := new(big.Int).Mul(big.NewInt(100), big.NewInt(1000000000000000000)) // 100 tokens with 18 decimals
	if balance.Cmp(minStake) < 0 {
		return false, nil // Not enough balance
	} else if allowance.Cmp(minStake) < 0 {
		return false, nil // Not enough allowance
	}

	return true, nil // Has enough balance and allowance
}

// registerVerifier registers the account as a verifier
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

	// Check DXP token balance and approval
	hasRequirements, err := checkDXPRequirements(client, address)
	if err != nil {
		fmt.Printf("Warning: Could not check DXP requirements: %v\n", err)
	} else if !hasRequirements {
		fmt.Println("Cannot proceed with registration due to insufficient DXP tokens or approval.")
		fmt.Println("Run './dxp-validator contract check-dxp' for more details.")
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

// checkDXPTokens checks DXP token balance and approval status
func checkDXPTokens() {
	// Connect to client
	client, err := getClient()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Get account
	_, address, err := getAccount()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Account address: %s\n", address.Hex())

	// Get DXP contract address
	contractAddr := os.Getenv("DXP_CONTRACT_ADDRESS")
	if contractAddr == "" {
		log.Fatalf("DXP_CONTRACT_ADDRESS not set in .env file")
	}

	// Create DXP token contract instance
	tokenAddress := common.HexToAddress(dxpTokenAddress)
	tokenABI := `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`
	parsedABI, err := abi.JSON(strings.NewReader(tokenABI))
	if err != nil {
		log.Fatalf("Failed to parse token ABI: %v", err)
	}
	tokenContract := bind.NewBoundContract(tokenAddress, parsedABI, client, client, client)

	// Check DXP token balance
	var balance *big.Int
	balanceResult := []interface{}{&balance}
	err = tokenContract.Call(&bind.CallOpts{}, &balanceResult, "balanceOf", address)
	if err != nil {
		fmt.Printf("Failed to get DXP token balance: %v\n", err)
		fmt.Println("This could mean the DXP token contract doesn't exist at the specified address.")
		return
	}

	// Check allowance
	var allowance *big.Int
	allowanceResult := []interface{}{&allowance}
	contractAddress := common.HexToAddress(contractAddr)
	err = tokenContract.Call(&bind.CallOpts{}, &allowanceResult, "allowance", address, contractAddress)
	if err != nil {
		fmt.Printf("Failed to get allowance: %v\n", err)
		return
	}

	// Format values for display (assuming 18 decimals)
	balanceInTokens := new(big.Float).Quo(
		new(big.Float).SetInt(balance),
		new(big.Float).SetInt(big.NewInt(1e18)),
	)
	allowanceInTokens := new(big.Float).Quo(
		new(big.Float).SetInt(allowance),
		new(big.Float).SetInt(big.NewInt(1e18)),
	)

	// Display results
	fmt.Printf("DXP Token Balance: %s DXP\n", balanceInTokens.Text('f', 6))
	fmt.Printf("Allowance for Contract: %s DXP\n", allowanceInTokens.Text('f', 6))

	// Check if balance and allowance are sufficient for registration
	minStake := new(big.Int).Mul(big.NewInt(100), big.NewInt(1000000000000000000)) // 100 tokens with 18 decimals
	if balance.Cmp(minStake) < 0 {
		fmt.Println("\nWARNING: Your DXP token balance is less than the required 100 DXP for registration.")
		fmt.Println("You need to acquire more DXP tokens before you can register as a verifier.")
	} else if allowance.Cmp(minStake) < 0 {
		fmt.Println("\nWARNING: Your allowance is less than the required 100 DXP for registration.")
		fmt.Println("You need to approve the contract to spend at least 100 DXP tokens.")
		fmt.Println("Run './dxp-validator contract approve --amount 100' to set the proper allowance.")
	} else {
		fmt.Println("\nYou have sufficient DXP tokens and allowance to register as a verifier.")
	}

	// Additional diagnostics
	fmt.Println("\nDiagnostic Information:")
	fmt.Printf("- DXP Token Address: %s\n", tokenAddress.Hex())
	fmt.Printf("- DXP Contract Address: %s\n", contractAddress.Hex())
	fmt.Println("- Required Stake: 100 DXP")
}

// approveDXPTokens approves the contract to spend DXP tokens
func approveDXPTokens() {
	// Connect to client
	client, err := getClient()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Get account
	privateKey, address, err := getAccount()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Account address: %s\n", address.Hex())

	// Get DXP token address from the contract
	contractAddr := os.Getenv("DXP_CONTRACT_ADDRESS")
	if contractAddr == "" {
		log.Fatalf("DXP_CONTRACT_ADDRESS not set in .env file")
	}

	// Use the global dxpTokenAddress variable

	// Create DXP token contract instance
	tokenAddress := common.HexToAddress(dxpTokenAddress)
	tokenABI := `[{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"}]`
	parsedABI, err := abi.JSON(strings.NewReader(tokenABI))
	if err != nil {
		log.Fatalf("Failed to parse token ABI: %v", err)
	}
	tokenContract := bind.NewBoundContract(tokenAddress, parsedABI, client, client, client)

	// Get auth options
	auth, err := getAuthOptions(client, privateKey)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Convert approval amount to wei (assuming 18 decimals)
	amount := new(big.Int).Mul(big.NewInt(approvalAmount), big.NewInt(1000000000000000000))
	fmt.Printf("Approving %d DXP tokens for contract %s...\n", approvalAmount, contractAddr)

	// Call approve function on the token contract
	contractAddress := common.HexToAddress(contractAddr)
	tx, err := tokenContract.Transact(auth, "approve", contractAddress, amount)
	if err != nil {
		log.Fatalf("Failed to approve tokens: %v", err)
	}

	fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())
	fmt.Println("Check the transaction status on Sepolia block explorer")
}

// submitProof submits a proof to the contract
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
