package config

import (
	"errors"
	"os"
	"strconv"
)

// Config holds the configuration for the validator node
type Config struct {
	BaseRPCURL        string
	DXPContractAddress string
	WalletPrivateKey  string
	GasPriceMultiplier float64
	GasLimit          uint64
	ChainID           int64
	LogLevel          string
	DataDir           string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Get required environment variables
	baseRPCURL := os.Getenv("BASE_RPC_URL")
	dxpContractAddress := os.Getenv("DXP_CONTRACT_ADDRESS")
	walletPrivateKey := os.Getenv("WALLET_PRIVATE_KEY")

	// Check required variables
	if baseRPCURL == "" || dxpContractAddress == "" || walletPrivateKey == "" {
		return nil, errors.New("missing required environment variables: BASE_RPC_URL, DXP_CONTRACT_ADDRESS, WALLET_PRIVATE_KEY")
	}

	// Get optional variables with defaults
	gasPriceMultiplier := 1.0
	if value := os.Getenv("GAS_PRICE_MULTIPLIER"); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			gasPriceMultiplier = parsed
		}
	}

	gasLimit := uint64(3000000)
	if value := os.Getenv("GAS_LIMIT"); value != "" {
		if parsed, err := strconv.ParseUint(value, 10, 64); err == nil {
			gasLimit = parsed
		}
	}

	chainID := int64(8453) // Default to Base chain ID
	if value := os.Getenv("CHAIN_ID"); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			chainID = parsed
		}
	}

	logLevel := "info"
	if value := os.Getenv("LOG_LEVEL"); value != "" {
		logLevel = value
	}

	dataDir := "./data"
	if value := os.Getenv("DATA_DIR"); value != "" {
		dataDir = value
	}

	return &Config{
		BaseRPCURL:        baseRPCURL,
		DXPContractAddress: dxpContractAddress,
		WalletPrivateKey:  walletPrivateKey,
		GasPriceMultiplier: gasPriceMultiplier,
		GasLimit:          gasLimit,
		ChainID:           chainID,
		LogLevel:          logLevel,
		DataDir:           dataDir,
	}, nil
}
