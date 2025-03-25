package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dexponent/geth-validator/internal/p2p"
)

func main() {
	// Create two validator P2P instances
	log.Println("Creating validator P2P instances...")

	// Create the first validator instance
	validator1, err := p2p.NewValidatorP2PIntegration("validator1", "localhost:9001", nil, nil)
	if err != nil {
		log.Fatalf("Failed to create validator1: %v", err)
	}

	// Create the second validator instance
	validator2, err := p2p.NewValidatorP2PIntegration("validator2", "localhost:9002", nil, nil)
	if err != nil {
		log.Fatalf("Failed to create validator2: %v", err)
	}

	// Start the validators
	log.Println("Starting validators...")
	go validator1.Start()
	go validator2.Start()

	// Wait for the validators to start
	time.Sleep(2 * time.Second)

	// Connect the validators to each other
	log.Println("Connecting validators...")
	validator1.AddPeer("validator2", "localhost:9002")
	validator2.AddPeer("validator1", "localhost:9001")

	// Wait for the connection to establish
	time.Sleep(2 * time.Second)

	// Simulate farm data for validator1
	farmData1 := &p2p.FarmData{
		SortinoRatio:   25.5,
		SharpeRatio:    30.2,
		MaxDrawdown:    15.8,
		Returns:        12.5,
		Timestamp:      time.Now().Unix(),
		ValidatorID:    "validator1",
		FarmID:         "farm1",
	}

	// Simulate farm data for validator2
	farmData2 := &p2p.FarmData{
		SortinoRatio:   25.5,
		SharpeRatio:    30.2,
		MaxDrawdown:    15.8,
		Returns:        12.5,
		Timestamp:      time.Now().Unix(),
		ValidatorID:    "validator2",
		FarmID:         "farm1",
	}

	// Calculate farm scores
	log.Println("Calculating farm scores...")
	farmScore1 := validator1.CalculateFarmScore(farmData1)
	farmScore2 := validator2.CalculateFarmScore(farmData2)

	log.Printf("Validator1 farm score: %f", farmScore1)
	log.Printf("Validator2 farm score: %f", farmScore2)

	// Broadcast farm data from validator1
	log.Println("Broadcasting farm data from validator1...")
	validator1.BroadcastFarmData(farmData1)

	// Wait for the data to be received by validator2
	time.Sleep(2 * time.Second)

	// Broadcast farm score from validator1
	log.Println("Broadcasting farm score from validator1...")
	validator1.BroadcastFarmScore("farm1", farmScore1)

	// Broadcast farm score from validator2
	log.Println("Broadcasting farm score from validator2...")
	validator2.BroadcastFarmScore("farm1", farmScore2)

	// Wait for consensus to be reached
	time.Sleep(5 * time.Second)

	// Check if consensus was reached
	log.Println("Checking consensus...")
	consensus1 := validator1.CheckConsensus("farm1")
	consensus2 := validator2.CheckConsensus("farm1")

	log.Printf("Validator1 consensus reached: %v", consensus1)
	log.Printf("Validator2 consensus reached: %v", consensus2)

	// Print final results
	fmt.Println("\nTest Results:")
	fmt.Printf("Validator1 farm score: %f\n", farmScore1)
	fmt.Printf("Validator2 farm score: %f\n", farmScore2)
	fmt.Printf("Consensus reached: %v\n", consensus1 && consensus2)

	// Keep the program running to observe logs
	log.Println("Test completed. Press Ctrl+C to exit.")
	select {}
}
