package consensus

import (
	"sync"
)

// Engine represents a consensus engine for validators
type Engine struct {
	participants     map[string]bool
	consensusResults map[string]map[string][]byte
	resultCounts     map[string]map[string]int
	mutex            sync.Mutex
}

// NewEngine creates a new consensus engine
func NewEngine() *Engine {
	return &Engine{
		participants:     make(map[string]bool),
		consensusResults: make(map[string]map[string][]byte),
		resultCounts:     make(map[string]map[string]int),
		mutex:            sync.Mutex{},
	}
}

// RegisterParticipant registers a participant in the consensus
func (e *Engine) RegisterParticipant(participantID string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.participants[participantID] = true
}

// SubmitResult submits a result for consensus
func (e *Engine) SubmitResult(requestID string, participantID string, result []byte) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Initialize maps if needed
	if _, ok := e.consensusResults[requestID]; !ok {
		e.consensusResults[requestID] = make(map[string][]byte)
		e.resultCounts[requestID] = make(map[string]int)
	}

	// Store the result
	resultKey := string(result)
	e.consensusResults[requestID][participantID] = result

	// Update the count for this result
	e.resultCounts[requestID][resultKey]++
}

// CheckConsensus checks if consensus has been reached for a request
func (e *Engine) CheckConsensus(requestID string) (bool, []byte) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Check if we have results for this request
	if _, ok := e.resultCounts[requestID]; !ok {
		return false, nil
	}

	// Count total participants who submitted results
	totalParticipants := len(e.consensusResults[requestID])
	if totalParticipants == 0 {
		return false, nil
	}

	// Find the result with the most votes
	maxCount := 0
	var consensusResult []byte

	for resultKey, count := range e.resultCounts[requestID] {
		if count > maxCount {
			maxCount = count
			// Find the actual result bytes from any participant
			for _, result := range e.consensusResults[requestID] {
				if string(result) == resultKey {
					consensusResult = result
					break
				}
			}
		}
	}

	// Check if we have a 2/3 majority
	if maxCount*3 >= totalParticipants*2 {
		return true, consensusResult
	}

	return false, nil
}
