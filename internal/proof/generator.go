package proof

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
)

// Generator represents a cryptographic proof generator
type Generator struct {
	proofs map[string][]byte
	mutex  sync.Mutex
}

// NewGenerator creates a new proof generator
func NewGenerator() *Generator {
	return &Generator{
		proofs: make(map[string][]byte),
		mutex:  sync.Mutex{},
	}
}

// GenerateProof generates a cryptographic proof for a result
func (g *Generator) GenerateProof(requestID string, result []byte) ([]byte, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Check if we already generated a proof for this request
	if proof, ok := g.proofs[requestID]; ok {
		return proof, nil
	}

	// For this example, we'll generate a simple proof by hashing the result
	// In a real implementation, this would be a more complex cryptographic proof
	if len(result) == 0 {
		return nil, errors.New("cannot generate proof for empty result")
	}

	// Generate a proof by double-hashing the result
	hash1 := sha256.Sum256(result)
	hash2 := sha256.Sum256(hash1[:])
	proof := []byte(fmt.Sprintf("proof:%s", hex.EncodeToString(hash2[:])))

	// Store the proof
	g.proofs[requestID] = proof

	return proof, nil
}

// VerifyProof verifies a cryptographic proof against a result
func (g *Generator) VerifyProof(result []byte, proof []byte) (bool, error) {
	// For this example, we'll verify the simple proof by recreating it
	// In a real implementation, this would be a more complex verification
	if len(result) == 0 || len(proof) == 0 {
		return false, errors.New("cannot verify proof with empty result or proof")
	}

	// Recreate the proof
	hash1 := sha256.Sum256(result)
	hash2 := sha256.Sum256(hash1[:])
	expectedProof := []byte(fmt.Sprintf("proof:%s", hex.EncodeToString(hash2[:])))

	// Compare the proofs
	return string(proof) == string(expectedProof), nil
}
