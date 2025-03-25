package main

import (
	"fmt"
	"log"
	"math"
	"time"
)

// FarmData represents the performance metrics of a farm
type FarmData struct {
	SortinoRatio   float64
	SharpeRatio    float64
	MaxDrawdown    float64
	Returns        float64
	Timestamp      int64
	ValidatorID    string
	FarmID         string
}

// FarmScoreCalculator calculates farm scores based on performance metrics
type FarmScoreCalculator struct{}

// CalculateFarmScore calculates the farm score based on the formula:
// FarmScore = 0.4(Sortino Ratio) + 0.4(Sharpe ratio) + 0.2(Maximum DrawDown) + 2(Returns)
func (f *FarmScoreCalculator) CalculateFarmScore(data *FarmData) float64 {
	return 0.4*data.SortinoRatio + 0.4*data.SharpeRatio + 0.2*data.MaxDrawdown + 2*data.Returns
}

// Validator represents a validator node in the network
type Validator struct {
	ID            string
	FarmScores    map[string]float64
	PeerScores    map[string]map[string]float64
	Calculator    *FarmScoreCalculator
	Peers         map[string]*Validator
}

// NewValidator creates a new validator instance
func NewValidator(id string) *Validator {
	return &Validator{
		ID:         id,
		FarmScores: make(map[string]float64),
		PeerScores: make(map[string]map[string]float64),
		Calculator: &FarmScoreCalculator{},
		Peers:      make(map[string]*Validator),
	}
}

// AddPeer adds a peer validator to this validator's peer list
func (v *Validator) AddPeer(peer *Validator) {
	v.Peers[peer.ID] = peer
	log.Printf("Validator %s added peer %s", v.ID, peer.ID)
}

// CalculateAndBroadcastFarmScore calculates a farm score and broadcasts it to peers
func (v *Validator) CalculateAndBroadcastFarmScore(data *FarmData) float64 {
	// Calculate the farm score
	score := v.Calculator.CalculateFarmScore(data)
	
	// Store the score locally
	v.FarmScores[data.FarmID] = score
	
	// Broadcast to peers
	for _, peer := range v.Peers {
		peer.ReceiveFarmScore(v.ID, data.FarmID, score)
	}
	
	log.Printf("Validator %s calculated farm score %.4f for farm %s and broadcast to peers", 
		v.ID, score, data.FarmID)
	
	return score
}

// ReceiveFarmScore receives a farm score from a peer
func (v *Validator) ReceiveFarmScore(peerID, farmID string, score float64) {
	// Initialize the map for this farm if it doesn't exist
	if _, ok := v.PeerScores[farmID]; !ok {
		v.PeerScores[farmID] = make(map[string]float64)
	}
	
	// Store the peer's score
	v.PeerScores[farmID][peerID] = score
	
	log.Printf("Validator %s received farm score %.4f for farm %s from peer %s", 
		v.ID, score, farmID, peerID)
}

// CheckConsensus checks if consensus has been reached for a farm
func (v *Validator) CheckConsensus(farmID string) (bool, float64) {
	// Get all scores for this farm (including our own)
	scores := make(map[string]float64)
	
	// Add our own score if we have one
	if ownScore, ok := v.FarmScores[farmID]; ok {
		scores[v.ID] = ownScore
	}
	
	// Add peer scores
	if peerScores, ok := v.PeerScores[farmID]; ok {
		for peerID, score := range peerScores {
			scores[peerID] = score
		}
	}
	
	// Count how many validators have scores within 1% of each other
	totalValidators := len(scores)
	if totalValidators < 2 {
		// Need at least 2 validators to reach consensus
		return false, 0
	}
	
	// Group scores that are within 1% of each other
	consensusGroups := make(map[int][]string)
	scoreToGroup := make(map[string]int)
	nextGroupID := 0
	
	for validatorID, score := range scores {
		// Check if this score is within 1% of any existing group
		foundGroup := false
		for groupID, members := range consensusGroups {
			// Use the first member's score as the reference
			referenceScore := scores[members[0]]
			diff := math.Abs(score - referenceScore) / referenceScore
			
			if diff <= 0.01 { // Within 1%
				consensusGroups[groupID] = append(consensusGroups[groupID], validatorID)
				scoreToGroup[validatorID] = groupID
				foundGroup = true
				break
			}
		}
		
		// If no matching group, create a new one
		if !foundGroup {
			consensusGroups[nextGroupID] = []string{validatorID}
			scoreToGroup[validatorID] = nextGroupID
			nextGroupID++
		}
	}
	
	// Find the largest consensus group
	largestGroupID := -1
	largestGroupSize := 0
	
	for groupID, members := range consensusGroups {
		if len(members) > largestGroupSize {
			largestGroupID = groupID
			largestGroupSize = len(members)
		}
	}
	
	// Check if the largest group has at least 2/3 of validators
	minRequiredSize := int(math.Ceil(float64(totalValidators) * 2 / 3))
	if largestGroupSize >= minRequiredSize {
		// Calculate the average score of the consensus group
		sum := 0.0
		for _, validatorID := range consensusGroups[largestGroupID] {
			sum += scores[validatorID]
		}
		averageScore := sum / float64(largestGroupSize)
		
		log.Printf("Validator %s: Consensus reached for farm %s with score %.4f (%d/%d validators agree)", 
			v.ID, farmID, averageScore, largestGroupSize, totalValidators)
		
		return true, averageScore
	}
	
	log.Printf("Validator %s: No consensus reached for farm %s (largest group: %d/%d validators)", 
		v.ID, farmID, largestGroupSize, totalValidators)
	
	return false, 0
}

func main() {
	// Create three validators
	validator1 := NewValidator("validator1")
	validator2 := NewValidator("validator2")
	validator3 := NewValidator("validator3")
	
	// Connect the validators in a mesh network
	validator1.AddPeer(validator2)
	validator1.AddPeer(validator3)
	validator2.AddPeer(validator1)
	validator2.AddPeer(validator3)
	validator3.AddPeer(validator1)
	validator3.AddPeer(validator2)
	
	fmt.Println("\n=== Test 1: All validators agree (should reach consensus) ===")
	
	// Create farm data with identical metrics for all validators
	farmData := &FarmData{
		SortinoRatio:   25.5,
		SharpeRatio:    30.2,
		MaxDrawdown:    15.8,
		Returns:        12.5,
		Timestamp:      time.Now().Unix(),
		ValidatorID:    "test",
		FarmID:         "farm1",
	}
	
	// Each validator calculates and broadcasts the farm score
	score1 := validator1.CalculateAndBroadcastFarmScore(farmData)
	score2 := validator2.CalculateAndBroadcastFarmScore(farmData)
	score3 := validator3.CalculateAndBroadcastFarmScore(farmData)
	
	fmt.Printf("Validator1 farm score: %.4f\n", score1)
	fmt.Printf("Validator2 farm score: %.4f\n", score2)
	fmt.Printf("Validator3 farm score: %.4f\n", score3)
	
	// Check if consensus was reached
	consensus1, consensusScore1 := validator1.CheckConsensus("farm1")
	consensus2, consensusScore2 := validator2.CheckConsensus("farm1")
	consensus3, consensusScore3 := validator3.CheckConsensus("farm1")
	
	fmt.Printf("Validator1 consensus reached: %v (score: %.4f)\n", consensus1, consensusScore1)
	fmt.Printf("Validator2 consensus reached: %v (score: %.4f)\n", consensus2, consensusScore2)
	fmt.Printf("Validator3 consensus reached: %v (score: %.4f)\n", consensus3, consensusScore3)
	
	fmt.Println("\n=== Test 2: One validator disagrees (should still reach consensus) ===")
	
	// Create a different farm with slightly different metrics for validator3
	farmData2 := &FarmData{
		SortinoRatio:   25.5,
		SharpeRatio:    30.2,
		MaxDrawdown:    15.8,
		Returns:        12.5,
		Timestamp:      time.Now().Unix(),
		ValidatorID:    "test",
		FarmID:         "farm2",
	}
	
	// Slightly different data for validator3 (within 1% difference)
	farmData2Variant := &FarmData{
		SortinoRatio:   25.7, // Slightly different
		SharpeRatio:    30.2,
		MaxDrawdown:    15.8,
		Returns:        12.5,
		Timestamp:      time.Now().Unix(),
		ValidatorID:    "test",
		FarmID:         "farm2",
	}
	
	// Each validator calculates and broadcasts the farm score
	score1 = validator1.CalculateAndBroadcastFarmScore(farmData2)
	score2 = validator2.CalculateAndBroadcastFarmScore(farmData2)
	score3 = validator3.CalculateAndBroadcastFarmScore(farmData2Variant) // Validator3 uses variant data
	
	fmt.Printf("Validator1 farm score: %.4f\n", score1)
	fmt.Printf("Validator2 farm score: %.4f\n", score2)
	fmt.Printf("Validator3 farm score: %.4f\n", score3)
	
	// Check if consensus was reached
	consensus1, consensusScore1 = validator1.CheckConsensus("farm2")
	consensus2, consensusScore2 = validator2.CheckConsensus("farm2")
	consensus3, consensusScore3 = validator3.CheckConsensus("farm2")
	
	fmt.Printf("Validator1 consensus reached: %v (score: %.4f)\n", consensus1, consensusScore1)
	fmt.Printf("Validator2 consensus reached: %v (score: %.4f)\n", consensus2, consensusScore2)
	fmt.Printf("Validator3 consensus reached: %v (score: %.4f)\n", consensus3, consensusScore3)
	
	fmt.Println("\n=== Test 3: Validators significantly disagree (should not reach consensus) ===")
	
	// Create a farm with significantly different metrics for each validator
	farmData3 := &FarmData{
		SortinoRatio:   25.5,
		SharpeRatio:    30.2,
		MaxDrawdown:    15.8,
		Returns:        12.5,
		Timestamp:      time.Now().Unix(),
		ValidatorID:    "test",
		FarmID:         "farm3",
	}
	
	farmData3Variant1 := &FarmData{
		SortinoRatio:   28.0, // >5% different
		SharpeRatio:    30.2,
		MaxDrawdown:    15.8,
		Returns:        12.5,
		Timestamp:      time.Now().Unix(),
		ValidatorID:    "test",
		FarmID:         "farm3",
	}
	
	farmData3Variant2 := &FarmData{
		SortinoRatio:   25.5,
		SharpeRatio:    32.0, // >5% different
		MaxDrawdown:    15.8,
		Returns:        13.5, // >5% different
		Timestamp:      time.Now().Unix(),
		ValidatorID:    "test",
		FarmID:         "farm3",
	}
	
	// Each validator calculates and broadcasts the farm score
	score1 = validator1.CalculateAndBroadcastFarmScore(farmData3)
	score2 = validator2.CalculateAndBroadcastFarmScore(farmData3Variant1)
	score3 = validator3.CalculateAndBroadcastFarmScore(farmData3Variant2)
	
	fmt.Printf("Validator1 farm score: %.4f\n", score1)
	fmt.Printf("Validator2 farm score: %.4f\n", score2)
	fmt.Printf("Validator3 farm score: %.4f\n", score3)
	
	// Check if consensus was reached
	consensus1, consensusScore1 = validator1.CheckConsensus("farm3")
	consensus2, consensusScore2 = validator2.CheckConsensus("farm3")
	consensus3, consensusScore3 = validator3.CheckConsensus("farm3")
	
	fmt.Printf("Validator1 consensus reached: %v (score: %.4f)\n", consensus1, consensusScore1)
	fmt.Printf("Validator2 consensus reached: %v (score: %.4f)\n", consensus2, consensusScore2)
	fmt.Printf("Validator3 consensus reached: %v (score: %.4f)\n", consensus3, consensusScore3)
}
