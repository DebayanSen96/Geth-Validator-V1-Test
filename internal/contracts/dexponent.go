// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// DexponentProtocolABI is the input ABI used to generate the binding from.
const DexponentProtocolABI = "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"farmId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"performanceScore\",\"type\":\"uint256\"}],\"name\":\"submitProof\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"registerVerifier\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"verifier\",\"type\":\"address\"}],\"name\":\"registeredVerifiers\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"triggerEmission\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// DexponentProtocol is an auto generated Go binding around an Ethereum contract.
type DexponentProtocol struct {
	DexponentProtocolCaller     // Read-only binding to the contract
	DexponentProtocolTransactor // Write-only binding to the contract
	DexponentProtocolFilterer   // Log filterer for contract events
}

// DexponentProtocolCaller is an auto generated read-only Go binding around an Ethereum contract.
type DexponentProtocolCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DexponentProtocolTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DexponentProtocolTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DexponentProtocolFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DexponentProtocolFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// NewDexponentProtocol creates a new instance of DexponentProtocol, bound to a specific deployed contract.
func NewDexponentProtocol(address common.Address, backend bind.ContractBackend) (*DexponentProtocol, error) {
	parsed, err := abi.JSON(strings.NewReader(DexponentProtocolABI))
	if err != nil {
		return nil, err
	}
	contract := bind.NewBoundContract(address, parsed, backend, backend, backend)
	return &DexponentProtocol{
		DexponentProtocolCaller:     DexponentProtocolCaller{contract: contract},
		DexponentProtocolTransactor: DexponentProtocolTransactor{contract: contract},
		DexponentProtocolFilterer:   DexponentProtocolFilterer{contract: contract},
	}, nil
}

// RegisteredVerifiers is a free data retrieval call binding the contract method 0x5f7a7e6a.
func (_DexponentProtocol *DexponentProtocolCaller) RegisteredVerifiers(opts *bind.CallOpts, verifier common.Address) (bool, error) {
	var out []interface{}
	err := _DexponentProtocol.contract.Call(opts, &out, "registeredVerifiers", verifier)
	return *abi.ConvertType(out[0], new(bool)).(*bool), err
}

// RegisterVerifier is a paid mutator transaction binding the contract method 0xb7b4a0e2.
func (_DexponentProtocol *DexponentProtocolTransactor) RegisterVerifier(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DexponentProtocol.contract.Transact(opts, "registerVerifier")
}

// SubmitProof is a paid mutator transaction binding the contract method 0x3c2c5f0a.
func (_DexponentProtocol *DexponentProtocolTransactor) SubmitProof(opts *bind.TransactOpts, farmId *big.Int, performanceScore *big.Int) (*types.Transaction, error) {
	return _DexponentProtocol.contract.Transact(opts, "submitProof", farmId, performanceScore)
}

// TriggerEmission is a paid mutator transaction binding the contract method 0x8a9c3d0d.
func (_DexponentProtocol *DexponentProtocolTransactor) TriggerEmission(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DexponentProtocol.contract.Transact(opts, "triggerEmission")
}
