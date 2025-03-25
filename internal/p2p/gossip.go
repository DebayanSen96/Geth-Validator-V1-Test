package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Message types
const (
	FarmDataMessageType  = "farm_data"
	FarmScoreMessageType = "farm_score"
	PeerDiscoveryType    = "peer_discovery"
)

// Message represents a message in the gossip protocol
type Message struct {
	Type      string          `json:"type"`
	Sender    string          `json:"sender"`
	RequestID string          `json:"request_id,omitempty"`
	FarmData  []float64       `json:"farm_data,omitempty"`
	FarmScore float64         `json:"farm_score,omitempty"`
	Timestamp int64           `json:"timestamp"`
}

// Peer represents a peer in the network
type Peer struct {
	ID        string
	Address   string
	LastSeen  time.Time
}

// GossipEngine represents a p2p gossip protocol engine
type GossipEngine struct {
	nodeID           string
	listenAddr       string
	peers            map[string]Peer
	knownMessages    map[string]bool
	messageCallbacks map[string]func(Message)
	scoreResults     map[string]map[string]float64
	mutex            sync.RWMutex
	listener         net.Listener
	running          bool
	cancel           context.CancelFunc
}

// NewGossipEngine creates a new gossip protocol engine
func NewGossipEngine(nodeID, listenAddr string) *GossipEngine {
	return &GossipEngine{
		nodeID:           nodeID,
		listenAddr:       listenAddr,
		peers:            make(map[string]Peer),
		knownMessages:    make(map[string]bool),
		messageCallbacks: make(map[string]func(Message)),
		scoreResults:     make(map[string]map[string]float64),
		mutex:            sync.RWMutex{},
	}
}

// Start starts the gossip protocol engine
func (g *GossipEngine) Start(ctx context.Context) error {
	g.mutex.Lock()
	if g.running {
		g.mutex.Unlock()
		return nil
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(ctx)
	g.cancel = cancel
	g.running = true
	g.mutex.Unlock()

	// Start listening for incoming connections
	listener, err := net.Listen("tcp", g.listenAddr)
	if err != nil {
		g.running = false
		return fmt.Errorf("failed to start listener: %v", err)
	}
	g.listener = listener

	log.Printf("P2P Gossip engine started on %s with node ID %s", g.listenAddr, g.nodeID)

	// Accept incoming connections
	go func() {
		for g.running {
			conn, err := listener.Accept()
			if err != nil {
				if g.running {
					log.Printf("Error accepting connection: %v", err)
				}
				break
			}
			go g.handleConnection(conn)
		}
	}()

	// Start peer discovery
	go g.discoverPeers(ctx)

	// Start periodic message broadcasting
	go g.periodicBroadcast(ctx)

	return nil
}

// Stop stops the gossip protocol engine
func (g *GossipEngine) Stop() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if !g.running {
		return
	}

	g.running = false
	if g.cancel != nil {
		g.cancel()
	}

	if g.listener != nil {
		g.listener.Close()
	}

	log.Printf("P2P Gossip engine stopped")
}

// AddPeer adds a peer to the gossip network
func (g *GossipEngine) AddPeer(id, address string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.peers[id] = Peer{
		ID:       id,
		Address:  address,
		LastSeen: time.Now(),
	}

	log.Printf("Added peer %s at %s", id, address)
}

// RegisterMessageCallback registers a callback for a specific message type
func (g *GossipEngine) RegisterMessageCallback(messageType string, callback func(Message)) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.messageCallbacks[messageType] = callback
}

// BroadcastFarmData broadcasts farm data to all peers
func (g *GossipEngine) BroadcastFarmData(requestID string, farmData []float64) {
	msg := Message{
		Type:      "farm_data",
		Sender:    g.nodeID,
		RequestID: requestID,
		FarmData:  farmData,
		Timestamp: time.Now().Unix(),
	}

	g.broadcastMessage(msg)
}

// BroadcastFarmScore broadcasts a calculated farm score to all peers
func (g *GossipEngine) BroadcastFarmScore(requestID string, farmScore float64) {
	msg := Message{
		Type:      "farm_score",
		Sender:    g.nodeID,
		RequestID: requestID,
		FarmScore: farmScore,
		Timestamp: time.Now().Unix(),
	}

	g.broadcastMessage(msg)

	// Store our own score result
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if _, ok := g.scoreResults[requestID]; !ok {
		g.scoreResults[requestID] = make(map[string]float64)
	}
	g.scoreResults[requestID][g.nodeID] = farmScore
}

// GetScoreResults gets all farm score results for a request
func (g *GossipEngine) GetScoreResults(requestID string) map[string]float64 {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	if results, ok := g.scoreResults[requestID]; ok {
		// Create a copy to avoid concurrent map access
		copy := make(map[string]float64)
		for k, v := range results {
			copy[k] = v
		}
		return copy
	}

	return make(map[string]float64)
}

// CheckConsensus checks if consensus has been reached for a farm score
func (g *GossipEngine) CheckConsensus(requestID string) (bool, float64) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	results, ok := g.scoreResults[requestID]
	if !ok || len(results) == 0 {
		return false, 0
	}

	// Count occurrences of each score
	scoreCounts := make(map[float64]int)
	for _, score := range results {
		scoreCounts[score]++
	}

	// Find the score with the most votes
	maxCount := 0
	var consensusScore float64
	for score, count := range scoreCounts {
		if count > maxCount {
			maxCount = count
			consensusScore = score
		}
	}

	// Check if we have a 2/3 majority
	totalParticipants := len(results)
	if maxCount*3 >= totalParticipants*2 {
		return true, consensusScore
	}

	return false, 0
}

// broadcastMessage broadcasts a message to all peers
func (g *GossipEngine) broadcastMessage(msg Message) {
	g.mutex.RLock()
	peers := make([]Peer, 0, len(g.peers))
	for _, peer := range g.peers {
		peers = append(peers, peer)
	}
	g.mutex.RUnlock()

	// Generate a unique message ID
	msgID := fmt.Sprintf("%s-%s-%d", msg.Type, msg.Sender, msg.Timestamp)

	// Check if we've already seen this message
	g.mutex.Lock()
	if g.knownMessages[msgID] {
		g.mutex.Unlock()
		return
	}
	g.knownMessages[msgID] = true
	g.mutex.Unlock()

	// Send the message to all peers
	for _, peer := range peers {
		go func(p Peer) {
			conn, err := net.Dial("tcp", p.Address)
			if err != nil {
				log.Printf("Failed to connect to peer %s at %s: %v", p.ID, p.Address, err)
				return
			}
			defer conn.Close()

			if err := json.NewEncoder(conn).Encode(msg); err != nil {
				log.Printf("Failed to send message to peer %s: %v", p.ID, err)
			}
		}(peer)
	}
}

// Broadcast broadcasts a message to all peers
func (g *GossipEngine) Broadcast(msg Message) {
	// Set timestamp if not already set
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().Unix()
	}
	
	log.Printf("Broadcasting message of type %s from %s", msg.Type, msg.Sender)
	
	// Use the internal broadcastMessage method
	g.broadcastMessage(msg)

	// Process the message locally
	g.processMessage(msg)
}

// handleConnection handles an incoming connection
func (g *GossipEngine) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Set a read deadline to prevent hanging connections
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Decode the message
	var msg Message
	if err := json.NewDecoder(conn).Decode(&msg); err != nil {
		log.Printf("Error decoding message: %v", err)
		return
	}

	// Update peer last seen time
	g.mutex.Lock()
	if peer, ok := g.peers[msg.Sender]; ok {
		peer.LastSeen = time.Now()
		g.peers[msg.Sender] = peer
	}
	g.mutex.Unlock()

	// Process the message
	g.processMessage(msg)

	// Relay the message to other peers
	g.broadcastMessage(msg)
}

// processMessage processes a received message
func (g *GossipEngine) processMessage(msg Message) {
	// Handle based on message type
	switch msg.Type {
	case "peer_discovery":
		// Add the sender as a peer if not already known
		g.mutex.Lock()
		if _, ok := g.peers[msg.Sender]; !ok {
			// Extract port from the message and format as a proper address
			port := fmt.Sprintf("%v", msg.FarmData[0])
			address := fmt.Sprintf("127.0.0.1:%s", port)
			g.peers[msg.Sender] = Peer{
				ID:       msg.Sender,
				Address:  address,
				LastSeen: time.Now(),
			}
			log.Printf("Discovered new peer %s at %s", msg.Sender, address)
		}
		g.mutex.Unlock()

	case "farm_data":
		// Nothing to do here, as we'll handle this in the callback

	case "farm_score":
		// Store the farm score result
		g.mutex.Lock()
		if _, ok := g.scoreResults[msg.RequestID]; !ok {
			g.scoreResults[msg.RequestID] = make(map[string]float64)
		}
		g.scoreResults[msg.RequestID][msg.Sender] = msg.FarmScore
		g.mutex.Unlock()
	}

	// Call the registered callback for this message type if any
	g.mutex.RLock()
	callback, ok := g.messageCallbacks[msg.Type]
	g.mutex.RUnlock()

	if ok {
		callback(msg)
	}
}

// discoverPeers periodically broadcasts peer discovery messages
func (g *GossipEngine) discoverPeers(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Create a peer discovery message with our address as the first element in FarmData
			msg := Message{
				Type:      "peer_discovery",
				Sender:    g.nodeID,
				FarmData:  []float64{float64(parseIPToInt(g.listenAddr))},
				Timestamp: time.Now().Unix(),
			}

			g.broadcastMessage(msg)

			// Clean up old peers
			g.cleanupOldPeers()
		}
	}
}

// periodicBroadcast periodically broadcasts a heartbeat message
func (g *GossipEngine) periodicBroadcast(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Send a heartbeat message
			msg := Message{
				Type:      "heartbeat",
				Sender:    g.nodeID,
				Timestamp: time.Now().Unix(),
			}

			g.broadcastMessage(msg)
		}
	}
}

// cleanupOldPeers removes peers that haven't been seen recently
func (g *GossipEngine) cleanupOldPeers() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	expiration := time.Now().Add(-2 * time.Minute)
	for id, peer := range g.peers {
		if peer.LastSeen.Before(expiration) {
			delete(g.peers, id)
			log.Printf("Removed stale peer %s", id)
		}
	}
}

// parseIPToInt converts an IP:port address to an integer for easy transmission
func parseIPToInt(addr string) int64 {
	// Parse the IP:port address properly
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		log.Printf("Invalid address format: %s", addr)
		return 0
	}
	
	// Just return the port as an integer, which is more reliable for connections
	port, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		log.Printf("Error parsing port from address %s: %v", addr, err)
		return 0
	}
	
	return port
}
