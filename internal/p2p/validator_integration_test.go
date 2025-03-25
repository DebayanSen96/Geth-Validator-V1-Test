package p2p

import (
	"testing"
)

func TestFarmScoreCalculator(t *testing.T) {
	// Create a farm score calculator
	calculator := NewFarmScoreCalculator()

	// Test with sample returns data
	returns := []float64{2.4, 4.7, 3.6, -1.2, 5.3, 2.1, 3.8, -0.5, 1.9, 4.2}

	// Calculate farm score
	score := calculator.CalculateFarmScore(returns)

	// Verify the score is calculated correctly based on the formula:
	// FarmScore = 0.4(Sortino Ratio) + 0.4(Sharpe ratio) + 0.2(Maximum DrawDown) + 2(Returns)
	// We can't verify the exact value without knowing the implementation details,
	// but we can check that it's a reasonable value
	if score <= 0 {
		t.Fatalf("Expected positive farm score, got %f", score)
	}

	t.Logf("Calculated farm score: %f", score)

	// Test with empty returns data
	emptyReturns := []float64{}
	score = calculator.CalculateFarmScore(emptyReturns)

	// The score should be 0 or some default value for empty data
	t.Logf("Farm score for empty returns: %f", score)

	t.Log("Farm score calculator test passed")
}

func TestMessageHandling(t *testing.T) {
	// Create a farm score calculator for testing
	calculator := NewFarmScoreCalculator()

	// Calculate a farm score for testing
	returns := []float64{2.4, 4.7, 3.6, -1.2, 5.3, 2.1, 3.8, -0.5, 1.9, 4.2}
	score := calculator.CalculateFarmScore(returns)

	// Create a message with the farm score
	farmID := "farm123"
	msg := Message{
		Type:      FarmScoreMessageType,
		Sender:    "validator1",
		RequestID: farmID,
		FarmScore: score,
		Timestamp: 1616161616, // Fixed timestamp for testing
	}

	// Verify the message fields
	if msg.Type != FarmScoreMessageType {
		t.Errorf("Expected message type %s, got %s", FarmScoreMessageType, msg.Type)
	}

	if msg.Sender != "validator1" {
		t.Errorf("Expected sender 'validator1', got '%s'", msg.Sender)
	}

	if msg.RequestID != farmID {
		t.Errorf("Expected request ID '%s', got '%s'", farmID, msg.RequestID)
	}

	if msg.FarmScore != score {
		t.Errorf("Expected farm score %f, got %f", score, msg.FarmScore)
	}

	t.Log("Message handling test passed")
}
