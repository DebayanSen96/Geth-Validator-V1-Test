package validator

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/dexponent/geth-validator/internal/compute"
	"github.com/dexponent/geth-validator/internal/config"
	"github.com/dexponent/geth-validator/internal/consensus"
	"github.com/dexponent/geth-validator/internal/contracts"
	"github.com/dexponent/geth-validator/internal/proof"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// DXPContract is an interface for the DXP smart contract
type DXPContract interface {
	RegisterValidator(opts *bind.TransactOpts) (*types.Transaction, error)
	IsRegistered(opts *bind.CallOpts, address common.Address) (bool, error)
	GetPendingRewards(opts *bind.CallOpts, address common.Address) (*big.Int, error)
	ClaimRewards(opts *bind.TransactOpts) (*types.Transaction, error)
	SubmitVerificationResult(opts *bind.TransactOpts, requestID *big.Int, result []byte, proof []byte) (*types.Transaction, error)
}

// VerificationRequest represents a request for verification
type VerificationRequest struct {
	ID        *big.Int
	Requester common.Address
	Data      []byte
	Timestamp *big.Int
}

// Validator represents a GETH-based validator node
type Validator struct {
	client          *ethclient.Client
	contract        DXPContract
	config          *config.Config
	privateKey      *ecdsa.PrivateKey
	address         common.Address
	nodeID          string
	running         bool
	registered      bool
	lastBlock       uint64
	verificationQueue []VerificationRequest
	consensusEngine  *consensus.Engine
	computeEngine    *compute.Engine
	proofGenerator   *proof.Generator
	mutex            sync.Mutex
	cancel          context.CancelFunc
}

// ValidatorStatus represents the status of the validator node
type ValidatorStatus struct {
	Running              bool
	NodeID               string
	Account              string
	Balance              float64
	Registered           bool
	LastBlockProcessed   uint64
	VerificationQueueSize int
	ConsensusParticipants int
}

// NewValidator creates a new validator instance
func NewValidator(cfg *config.Config) (*Validator, error) {
	// Connect to Ethereum client
	client, err := ethclient.Dial(cfg.BaseRPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(cfg.WalletPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}

	// Get public address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Create contract instance
	contractAddress := common.HexToAddress(cfg.DXPContractAddress)
	contract, err := contracts.NewDexponentContractWrapper(contractAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract instance: %v", err)
	}

	// Generate a unique node ID
	nodeID := hexutil.Encode(crypto.Keccak256([]byte(address.Hex() + time.Now().String())))[2:10]

	// Create consensus engine
	consensusEngine := consensus.NewEngine()

	// Create compute engine
	computeEngine := compute.NewEngine()

	// Create proof generator
	proofGenerator := proof.NewGenerator()

	return &Validator{
		client:          client,
		contract:        contract,
		config:          cfg,
		privateKey:      privateKey,
		address:         address,
		nodeID:          nodeID,
		running:         false,
		registered:      false,
		lastBlock:       0,
		verificationQueue: make([]VerificationRequest, 0),
		consensusEngine:  consensusEngine,
		computeEngine:    computeEngine,
		proofGenerator:   proofGenerator,
		mutex:            sync.Mutex{},
	}, nil
}

// IsRegistered checks if the validator is registered with the DXP contract
func (v *Validator) IsRegistered() (bool, error) {
	auth := &bind.CallOpts{
		From: v.address,
	}

	isRegistered, err := v.contract.IsRegistered(auth, v.address)
	if err != nil {
		return false, fmt.Errorf("failed to check registration status: %v", err)
	}

	v.registered = isRegistered
	return isRegistered, nil
}

// RegisterValidator registers the validator with the DXP contract
func (v *Validator) RegisterValidator() (string, error) {
	// Create transaction options
	chainID := big.NewInt(v.config.ChainID)
	auth, err := bind.NewKeyedTransactorWithChainID(v.privateKey, chainID)
	if err != nil {
		return "", fmt.Errorf("failed to create transaction options: %v", err)
	}

	// Set gas price and limit
	gasPrice, err := v.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to suggest gas price: %v", err)
	}

	// Apply gas price multiplier
	multiplier := big.NewFloat(v.config.GasPriceMultiplier)
	adjustedGasPrice := new(big.Float).Mul(new(big.Float).SetInt(gasPrice), multiplier)
	adjustedGasPriceInt, _ := adjustedGasPrice.Int(nil)

	auth.GasPrice = adjustedGasPriceInt
	auth.GasLimit = v.config.GasLimit

	// Register validator
	tx, err := v.contract.RegisterValidator(auth)
	if err != nil {
		return "", fmt.Errorf("failed to register validator: %v", err)
	}

	v.registered = true
	return tx.Hash().Hex(), nil
}

// Start starts the validator node
func (v *Validator) Start(ctx context.Context, blockPollingInterval int) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.running {
		return errors.New("validator is already running")
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(ctx)
	v.cancel = cancel

	// Start block processing
	go v.processBlocks(ctx, blockPollingInterval)

	// Start verification processing
	go v.processVerifications(ctx)

	v.running = true
	return nil
}

// Stop stops the validator node
func (v *Validator) Stop() {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if !v.running {
		return
	}

	// Cancel context to stop all goroutines
	if v.cancel != nil {
		v.cancel()
	}

	v.running = false
}

// processBlocks continuously processes new blocks
func (v *Validator) processBlocks(ctx context.Context, blockPollingInterval int) {
	ticker := time.NewTicker(time.Duration(blockPollingInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get latest block number
			latestBlock, err := v.client.BlockNumber(ctx)
			if err != nil {
				log.Printf("Error getting latest block: %v", err)
				continue
			}

			// Process new blocks
			if v.lastBlock == 0 {
				v.lastBlock = latestBlock
				continue
			}

			for blockNum := v.lastBlock + 1; blockNum <= latestBlock; blockNum++ {
				if err := v.processBlock(ctx, blockNum); err != nil {
					log.Printf("Error processing block %d: %v", blockNum, err)
					continue
				}
				v.lastBlock = blockNum
			}
		}
	}
}

// processBlock processes a single block
func (v *Validator) processBlock(ctx context.Context, blockNum uint64) error {
	// In a real implementation, we would filter events from the DXP contract
	// For this example, we'll simulate finding verification requests
	
	// Simulate finding a verification request every 10 blocks
	if blockNum%10 == 0 {
		// Create a simulated verification request
		request := VerificationRequest{
			ID:        big.NewInt(int64(blockNum)),
			Requester: common.HexToAddress("0x1234567890123456789012345678901234567890"),
			Data:      []byte(fmt.Sprintf("verification_data_%d", blockNum)),
			Timestamp: big.NewInt(time.Now().Unix()),
		}

		// Add to verification queue
		v.mutex.Lock()
		v.verificationQueue = append(v.verificationQueue, request)
		v.mutex.Unlock()

		log.Printf("Found verification request: %s", request.ID.String())
	}

	return nil
}

// processVerifications processes verification requests in the queue
func (v *Validator) processVerifications(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			v.mutex.Lock()
			if len(v.verificationQueue) > 0 {
				// Get the next verification request
				request := v.verificationQueue[0]
				v.verificationQueue = v.verificationQueue[1:]
				v.mutex.Unlock()

				// Process the verification request
				go v.verifyRequest(ctx, request)
			} else {
				v.mutex.Unlock()
			}
		}
	}
}

// verifyRequest processes a single verification request
func (v *Validator) verifyRequest(ctx context.Context, request VerificationRequest) {
	log.Printf("Processing verification request: %s", request.ID.String())

	// 1. Submit the verification task to the compute engine
	taskID := v.computeEngine.SubmitTask(request.ID.String(), request.Data)

	// 2. Wait for the computation to complete
	result, err := v.computeEngine.WaitForResult(taskID, 30*time.Second)
	if err != nil {
		log.Printf("Error computing result: %v", err)
		return
	}

	// 3. Submit the result to the consensus engine
	v.consensusEngine.SubmitResult(request.ID.String(), v.nodeID, result)

	// 4. Wait for consensus
	consensusReached, consensusResult := v.consensusEngine.CheckConsensus(request.ID.String())
	if !consensusReached {
		log.Printf("Consensus not reached for request: %s", request.ID.String())
		return
	}

	// 5. Generate proof for the consensus result
	proof, err := v.proofGenerator.GenerateProof(request.ID.String(), consensusResult)
	if err != nil {
		log.Printf("Error generating proof: %v", err)
		return
	}

	// 6. Submit the result and proof to the smart contract
	if err := v.submitResult(request.ID, consensusResult, proof); err != nil {
		log.Printf("Error submitting result: %v", err)
		return
	}

	log.Printf("Successfully processed verification request: %s", request.ID.String())
}

// submitResult submits the verification result and proof to the smart contract
func (v *Validator) submitResult(requestID *big.Int, result []byte, proof []byte) error {
	// Create transaction options
	chainID := big.NewInt(v.config.ChainID)
	auth, err := bind.NewKeyedTransactorWithChainID(v.privateKey, chainID)
	if err != nil {
		return fmt.Errorf("failed to create transaction options: %v", err)
	}

	// Set gas price and limit
	gasPrice, err := v.client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to suggest gas price: %v", err)
	}

	// Apply gas price multiplier
	multiplier := big.NewFloat(v.config.GasPriceMultiplier)
	adjustedGasPrice := new(big.Float).Mul(new(big.Float).SetInt(gasPrice), multiplier)
	adjustedGasPriceInt, _ := adjustedGasPrice.Int(nil)

	auth.GasPrice = adjustedGasPriceInt
	auth.GasLimit = v.config.GasLimit

	// Submit result and proof
	tx, err := v.contract.SubmitVerificationResult(auth, requestID, result, proof)
	if err != nil {
		return fmt.Errorf("failed to submit verification result: %v", err)
	}

	log.Printf("Submitted verification result, tx: %s", tx.Hash().Hex())
	return nil
}

// GetValidatorStatus returns the status of a validator node
func GetValidatorStatus(cfg *config.Config) (*ValidatorStatus, error) {
	// In a real implementation, we would check the status of a running validator
	// For this example, we'll return a simulated status
	return &ValidatorStatus{
		Running:              true,
		NodeID:               "0x1234abcd",
		Account:              "0x5678efgh",
		Balance:              1.234,
		Registered:           true,
		LastBlockProcessed:   12345,
		VerificationQueueSize: 5,
		ConsensusParticipants: 3,
	}, nil
}

// GetValidatorRewards returns the pending rewards for a validator
func GetValidatorRewards(cfg *config.Config) (float64, error) {
	// In a real implementation, we would check the rewards from the contract
	// For this example, we'll return a simulated value
	return 0.5, nil
}

// ClaimValidatorRewards claims the pending rewards for a validator
func ClaimValidatorRewards(cfg *config.Config) (string, error) {
	// In a real implementation, we would call the contract to claim rewards
	// For this example, we'll return a simulated transaction hash
	return "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", nil
}

// StopValidator stops a running validator node
func StopValidator(cfg *config.Config) error {
	// In a real implementation, we would stop a running validator
	// For this example, we'll just return success
	return nil
}

// MockDXPContract is a mock implementation of the DXPContract interface
type MockDXPContract struct{}

// NewMockDXPContract creates a new mock contract
func NewMockDXPContract() *MockDXPContract {
	return &MockDXPContract{}
}

// RegisterValidator mock implementation
func (m *MockDXPContract) RegisterValidator(opts *bind.TransactOpts) (*types.Transaction, error) {
	// Create a dummy transaction
	return types.NewTransaction(
		0,
		common.HexToAddress("0x0000000000000000000000000000000000000000"),
		big.NewInt(0),
		0,
		big.NewInt(0),
		nil,
	), nil
}

// IsRegistered mock implementation
func (m *MockDXPContract) IsRegistered(opts *bind.CallOpts, address common.Address) (bool, error) {
	return true, nil
}

// GetPendingRewards mock implementation
func (m *MockDXPContract) GetPendingRewards(opts *bind.CallOpts, address common.Address) (*big.Int, error) {
	return big.NewInt(500000000000000000), nil // 0.5 ETH
}

// ClaimRewards mock implementation
func (m *MockDXPContract) ClaimRewards(opts *bind.TransactOpts) (*types.Transaction, error) {
	// Create a dummy transaction
	return types.NewTransaction(
		0,
		common.HexToAddress("0x0000000000000000000000000000000000000000"),
		big.NewInt(0),
		0,
		big.NewInt(0),
		nil,
	), nil
}

// SubmitVerificationResult mock implementation
func (m *MockDXPContract) SubmitVerificationResult(opts *bind.TransactOpts, requestID *big.Int, result []byte, proof []byte) (*types.Transaction, error) {
	// Create a dummy transaction
	return types.NewTransaction(
		0,
		common.HexToAddress("0x0000000000000000000000000000000000000000"),
		big.NewInt(0),
		0,
		big.NewInt(0),
		nil,
	), nil
}
