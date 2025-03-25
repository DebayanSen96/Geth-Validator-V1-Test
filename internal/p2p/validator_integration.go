package p2p

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// FarmDataFetcher fetches farm data from a smart contract
type FarmDataFetcher struct {
	client         *ethclient.Client
	contractAddr   common.Address
	mutex          sync.Mutex
}

// NewFarmDataFetcher creates a new farm data fetcher
func NewFarmDataFetcher(client *ethclient.Client, contractAddr common.Address) *FarmDataFetcher {
	return &FarmDataFetcher{
		client:       client,
		contractAddr: contractAddr,
		mutex:        sync.Mutex{},
	}
}

// FetchFarmData fetches farm returns data from the smart contract
// For now, this is a dummy implementation that returns mock data
func (f *FarmDataFetcher) FetchFarmData(ctx context.Context, farmID string) ([]float64, error) {
	// TODO: Implement actual contract call to fetch farm returns
	// For now, return mock data
	return []float64{2.4, 4.7, 3.6, -1.2, 5.3, 2.1, 3.8, -0.5, 1.9, 4.2}, nil
}

// FetchFarmReturns fetches the returns data for a specific farm
func (f *FarmDataFetcher) FetchFarmReturns(farmID string) ([]float64, error) {
	log.Printf("Fetching returns data for farm %s from protocol master contract", farmID)
	
	// Call FetchFarmData with a background context
	return f.FetchFarmData(context.Background(), farmID)
}

// FarmScoreCallback is a function that handles farm scores from the p2p network
type FarmScoreCallback func(farmID string, score float64)

// ValidatorP2PIntegration integrates the p2p gossip protocol with the validator
type ValidatorP2PIntegration struct {
	nodeID          string
	gossipEngine    *GossipEngine
	farmCalculator  *FarmScoreCalculator
	farmDataFetcher *FarmDataFetcher
	protocolMaster  common.Address
	client          *ethclient.Client
	pendingRequests map[string]time.Time
	resultsMutex    sync.Mutex
	// Callback for farm scores to be processed by the validator
	farmScoreCallback FarmScoreCallback
}

// NewValidatorP2PIntegration creates a new validator p2p integration
func NewValidatorP2PIntegration(
	nodeID string,
	listenAddr string,
	client *ethclient.Client,
	protocolMaster common.Address,
) *ValidatorP2PIntegration {
	gossipEngine := NewGossipEngine(nodeID, listenAddr)
	farmCalculator := NewFarmScoreCalculator()
	farmDataFetcher := NewFarmDataFetcher(client, protocolMaster)

	integration := &ValidatorP2PIntegration{
		nodeID:          nodeID,
		gossipEngine:    gossipEngine,
		farmCalculator:  farmCalculator,
		farmDataFetcher: farmDataFetcher,
		protocolMaster:  protocolMaster,
		client:          client,
		pendingRequests: make(map[string]time.Time),
		resultsMutex:    sync.Mutex{},
	}

	// Register message callbacks
	gossipEngine.RegisterMessageCallback("farm_data", integration.handleFarmDataMessage)
	gossipEngine.RegisterMessageCallback("farm_score", integration.handleFarmScoreMessage)

	return integration
}

// Start starts the validator p2p integration
func (v *ValidatorP2PIntegration) Start(ctx context.Context) error {
	// Start the gossip engine
	if err := v.gossipEngine.Start(ctx); err != nil {
		return err
	}

	// Start the farm data processing loop
	go v.processFarmData(ctx)

	// Start the consensus checking loop
	go v.checkConsensus(ctx)

	return nil
}

// Stop stops the validator p2p integration
func (v *ValidatorP2PIntegration) Stop() {
	v.gossipEngine.Stop()
}

// SetFarmScoreCallback sets the callback function for handling farm scores
func (v *ValidatorP2PIntegration) SetFarmScoreCallback(callback FarmScoreCallback) {
	v.farmScoreCallback = callback
	log.Printf("Farm score callback registered for validator %s", v.nodeID)
}

// AddPeer adds a peer to the gossip network
func (v *ValidatorP2PIntegration) AddPeer(id, address string) {
	v.gossipEngine.AddPeer(id, address)
}

// GetActiveFarmIDs returns a list of active farm IDs
func (v *ValidatorP2PIntegration) GetActiveFarmIDs() ([]string, error) {
	log.Printf("Fetching active farm IDs from protocol master contract")
	
	// In a real implementation, this would query the protocol master contract
	// For now, we'll return a mock list of farm IDs
	return []string{"1", "2", "3"}, nil
}

// GetFarmReturns fetches the returns data for a specific farm
func (v *ValidatorP2PIntegration) GetFarmReturns(farmID string) ([]float64, error) {
	log.Printf("Fetching returns data for farm %s", farmID)
	
	// In a real implementation, this would query the protocol master contract
	// For now, we'll use the farm data fetcher to get mock data
	return v.farmDataFetcher.FetchFarmReturns(farmID)
}

// CalculateFarmScore calculates the farm score based on returns data
func (v *ValidatorP2PIntegration) CalculateFarmScore(returns []float64) float64 {
	// Use the farm calculator to calculate the score based on the Dexponent formula
	return v.farmCalculator.CalculateFarmScore(returns)
}

// BroadcastFarmScore broadcasts a farm score to all peers
func (v *ValidatorP2PIntegration) BroadcastFarmScore(farmID string, score float64) {
	log.Printf("Broadcasting farm score %f for farm %s to peers", score, farmID)
	
	// Create a farm score message
	msg := Message{
		Type:      FarmScoreMessageType,
		Sender:    v.nodeID,
		RequestID: farmID,
		FarmScore: score,
		Timestamp: time.Now().Unix(),
	}
	
	// Broadcast the message to all peers
	v.gossipEngine.Broadcast(msg)
}

// processFarmData periodically fetches and processes farm data
func (v *ValidatorP2PIntegration) processFarmData(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Generate a request ID based on current time
			requestID := fmt.Sprintf("farm-data-%d", time.Now().Unix())

			// Fetch farm data from the smart contract
			farmData, err := v.farmDataFetcher.FetchFarmData(ctx, "farm1")
			if err != nil {
				log.Printf("Error fetching farm data: %v", err)
				continue
			}

			// Broadcast the farm data to all peers
			v.gossipEngine.BroadcastFarmData(requestID, farmData)

			// Calculate our own farm score
			farmScore := v.farmCalculator.CalculateFarmScore(farmData)

			// Broadcast our farm score
			v.gossipEngine.BroadcastFarmScore(requestID, farmScore)

			// Add to pending requests
			v.resultsMutex.Lock()
			v.pendingRequests[requestID] = time.Now()
			v.resultsMutex.Unlock()

			log.Printf("Processed farm data for request %s, calculated score: %f", requestID, farmScore)
		}
	}
}

// checkConsensus periodically checks for consensus on farm scores
func (v *ValidatorP2PIntegration) checkConsensus(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			v.resultsMutex.Lock()
			pendingRequests := make([]string, 0, len(v.pendingRequests))
			for requestID := range v.pendingRequests {
				pendingRequests = append(pendingRequests, requestID)
			}
			v.resultsMutex.Unlock()

			for _, requestID := range pendingRequests {
				// Check if consensus has been reached
				consensusReached, consensusScore := v.gossipEngine.CheckConsensus(requestID)
				if consensusReached {
					log.Printf("Consensus reached for request %s with score %f", requestID, consensusScore)

					// Submit the result to the protocol master contract
					err := v.submitConsensusResult(requestID, consensusScore)
					if err != nil {
						log.Printf("Error submitting consensus result: %v", err)
					} else {
						log.Printf("Successfully submitted consensus result for request %s", requestID)

						// Remove from pending requests
						v.resultsMutex.Lock()
						delete(v.pendingRequests, requestID)
						v.resultsMutex.Unlock()
					}
				} else {
					// Check if the request has timed out
					v.resultsMutex.Lock()
					requestTime, ok := v.pendingRequests[requestID]
					v.resultsMutex.Unlock()

					if ok && time.Since(requestTime) > 5*time.Minute {
						log.Printf("Request %s timed out without reaching consensus", requestID)

						// Remove from pending requests
						v.resultsMutex.Lock()
						delete(v.pendingRequests, requestID)
						v.resultsMutex.Unlock()
					}
				}
			}
		}
	}
}

// submitConsensusResult submits the consensus result to the protocol master contract
func (v *ValidatorP2PIntegration) submitConsensusResult(requestID string, score float64) error {
	// TODO: Implement actual contract call to submit the consensus result
	log.Printf("Would submit consensus result to protocol master contract: requestID=%s, score=%f", requestID, score)
	return nil
}

// handleFarmDataMessage handles a farm data message from a peer
func (v *ValidatorP2PIntegration) handleFarmDataMessage(msg Message) {
	// Validate the message
	if msg.FarmData == nil || len(msg.FarmData) == 0 {
		return
	}

	log.Printf("Received farm data from peer %s for request %s", msg.Sender, msg.RequestID)

	// Calculate our farm score based on the received data
	farmScore := v.farmCalculator.CalculateFarmScore(msg.FarmData)

	// Broadcast our farm score
	v.gossipEngine.BroadcastFarmScore(msg.RequestID, farmScore)

	// Add to pending requests if not already there
	v.resultsMutex.Lock()
	if _, ok := v.pendingRequests[msg.RequestID]; !ok {
		v.pendingRequests[msg.RequestID] = time.Now()
	}
	v.resultsMutex.Unlock()

	log.Printf("Calculated farm score %f for request %s based on peer data", farmScore, msg.RequestID)
}

// handleFarmScoreMessage handles a farm score message from a peer
func (v *ValidatorP2PIntegration) handleFarmScoreMessage(msg Message) {
	log.Printf("Received farm score %f from peer %s for request %s", msg.FarmScore, msg.Sender, msg.RequestID)

	// Add to pending requests if not already there
	v.resultsMutex.Lock()
	if _, ok := v.pendingRequests[msg.RequestID]; !ok {
		v.pendingRequests[msg.RequestID] = time.Now()
	}
	v.resultsMutex.Unlock()

	// If a callback is registered, send the farm score to the validator
	if v.farmScoreCallback != nil {
		log.Printf("Forwarding farm score %f for farm %s to validator", msg.FarmScore, msg.RequestID)
		v.farmScoreCallback(msg.RequestID, msg.FarmScore)
	} else {
		log.Printf("No callback registered for farm scores, score %f for farm %s will not be processed", msg.FarmScore, msg.RequestID)
	}
}
