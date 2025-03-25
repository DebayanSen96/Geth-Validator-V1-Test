package p2p

import (
	"testing"
)

func TestMessageTypes(t *testing.T) {
	// Test that message type constants are defined correctly
	if FarmDataMessageType != "farm_data" {
		t.Errorf("Expected FarmDataMessageType to be 'farm_data', got '%s'", FarmDataMessageType)
	}

	if FarmScoreMessageType != "farm_score" {
		t.Errorf("Expected FarmScoreMessageType to be 'farm_score', got '%s'", FarmScoreMessageType)
	}

	if PeerDiscoveryType != "peer_discovery" {
		t.Errorf("Expected PeerDiscoveryType to be 'peer_discovery', got '%s'", PeerDiscoveryType)
	}

	t.Log("Message type constants are defined correctly")
}

func TestCheckConsensus(t *testing.T) {
	// Create a gossip engine
	gossip := NewGossipEngine("test-node", "localhost:9000")

	// Add scores manually
	gossip.mutex.Lock()
	requestID := "test-request"
	if gossip.scoreResults[requestID] == nil {
		gossip.scoreResults[requestID] = make(map[string]float64)
	}

	// Add 3 identical scores (should reach consensus)
	gossip.scoreResults[requestID]["node1"] = 85.5
	gossip.scoreResults[requestID]["node2"] = 85.5
	gossip.scoreResults[requestID]["node3"] = 85.5
	gossip.mutex.Unlock()

	// Check consensus
	hasConsensus, consensusScore := gossip.CheckConsensus(requestID)
	if !hasConsensus {
		t.Fatalf("Expected consensus but got none")
	}

	if consensusScore != 85.5 {
		t.Fatalf("Expected consensus score of 85.5 but got %f", consensusScore)
	}

	// Test with different scores (should not reach consensus)
	gossip.mutex.Lock()
	requestID = "test-request-2"
	if gossip.scoreResults[requestID] == nil {
		gossip.scoreResults[requestID] = make(map[string]float64)
	}

	gossip.scoreResults[requestID]["node1"] = 85.5
	gossip.scoreResults[requestID]["node2"] = 75.5
	gossip.scoreResults[requestID]["node3"] = 65.5
	gossip.mutex.Unlock()

	hasConsensus, _ = gossip.CheckConsensus(requestID)
	if hasConsensus {
		t.Fatalf("Expected no consensus but got consensus")
	}

	t.Log("Consensus check works correctly")
}
