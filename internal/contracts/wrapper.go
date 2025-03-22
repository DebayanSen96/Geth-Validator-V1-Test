package contracts

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// DexponentContractWrapper implements the validator.DXPContract interface
// for the Dexponent Protocol contract
type DexponentContractWrapper struct {
	contract *DexponentProtocol
}

// NewDexponentContractWrapper creates a new wrapper for the Dexponent Protocol contract
func NewDexponentContractWrapper(address common.Address, backend bind.ContractBackend) (*DexponentContractWrapper, error) {
	contract, err := NewDexponentProtocol(address, backend)
	if err != nil {
		return nil, err
	}

	return &DexponentContractWrapper{
		contract: contract,
	}, nil
}

// RegisterValidator registers the validator with the Dexponent Protocol contract
func (w *DexponentContractWrapper) RegisterValidator(opts *bind.TransactOpts) (*types.Transaction, error) {
	return w.contract.RegisterVerifier(opts)
}

// IsRegistered checks if the validator is registered with the Dexponent Protocol contract
func (w *DexponentContractWrapper) IsRegistered(opts *bind.CallOpts, address common.Address) (bool, error) {
	return w.contract.RegisteredVerifiers(opts, address)
}

// GetPendingRewards gets the pending rewards for the validator
// Note: This is a mock implementation as the actual contract doesn't have this method
func (w *DexponentContractWrapper) GetPendingRewards(opts *bind.CallOpts, address common.Address) (*big.Int, error) {
	// This is a mock implementation since the contract doesn't have this method
	return big.NewInt(0), nil
}

// ClaimRewards claims the pending rewards for the validator
// Note: This is a mock implementation as the actual contract doesn't have this method
func (w *DexponentContractWrapper) ClaimRewards(opts *bind.TransactOpts) (*types.Transaction, error) {
	// This is a mock implementation since the contract doesn't have this method
	return nil, nil
}

// SubmitVerificationResult submits the verification result to the Dexponent Protocol contract
func (w *DexponentContractWrapper) SubmitVerificationResult(opts *bind.TransactOpts, requestID *big.Int, result []byte, proof []byte) (*types.Transaction, error) {
	// Convert the result to a performance score
	// For simplicity, we'll use a fixed performance score of 100
	performanceScore := big.NewInt(100)

	// Submit the proof to the contract
	return w.contract.SubmitProof(opts, requestID, performanceScore)
}
