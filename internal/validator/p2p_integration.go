package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dexponent/geth-validator/internal/config"
	"github.com/dexponent/geth-validator/internal/p2p"
	"github.com/libp2p/go-libp2p/core/peer"
)

// P2PValidator extends the Validator with p2p networking capabilities
type P2PValidator struct {
	*Validator
	p2pHost        *p2p.Host
	p2pProtocol    *p2p.Protocol
	p2pConfig      p2p.Config
	peers          map[peer.ID]*ValidatorPeer
	peersMutex     sync.RWMutex
	lastBlockSeen   uint64
	proofsSubmitted uint64
}

// ValidatorPeer represents information about a connected validator peer
type ValidatorPeer struct {
	ID              peer.ID
	Address         string
	Registered      bool
	LastBlockSeen   uint64
	ProofsSubmitted uint64
	LastSeen        int64 // Unix timestamp
}

// NewP2PValidator creates a new validator with p2p capabilities
func NewP2PValidator(cfg *config.Config) (*P2PValidator, error) {
	// Create the base validator
	baseValidator, err := NewValidator(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create base validator: %w", err)
	}

	// Load p2p config
	p2pConfig, err := p2p.LoadP2PConfig(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load p2p config: %w", err)
	}

	return &P2PValidator{
		Validator:       baseValidator,
		p2pConfig:       p2pConfig,
		peers:           make(map[peer.ID]*ValidatorPeer),
		lastBlockSeen:   0,
		proofsSubmitted: 0,
	}, nil
}

// Start starts the validator node with p2p networking
func (v *P2PValidator) Start(ctx context.Context, blockPollingInterval int) error {
	// Initialize peers map if not already initialized
	if v.peers == nil {
		v.peers = make(map[peer.ID]*ValidatorPeer)
	}

	// Start the base validator
	if err := v.Validator.Start(ctx, blockPollingInterval); err != nil {
		return err
	}

	// Create and start the p2p host
	log.Printf("Starting p2p networking on %v", v.p2pConfig.ListenAddresses)
	p2pHost, err := p2p.NewHost(ctx, v.p2pConfig, "/dxp/validator/1.0.0")
	if err != nil {
		return fmt.Errorf("failed to create p2p host: %w", err)
	}
	v.p2pHost = p2pHost

	// Create the protocol handler
	v.p2pProtocol = p2p.NewProtocol(p2pHost, "/dxp/validator/1.0.0", v.handleMessage)

	// Log the node's addresses
	addrs := p2pHost.Addrs()
	peerID := p2pHost.ID()
	log.Printf("P2P node started with ID: %s", peerID.String())
	log.Printf("P2P node addresses:")
	for _, addr := range addrs {
		log.Printf("  %s/p2p/%s", addr.String(), peerID.String())
	}

	// Start broadcasting status updates periodically
	go v.broadcastStatus(ctx)

	// Start a goroutine to sync peers from the p2p host
	go v.syncPeersFromHost(ctx)

	return nil
}

// Stop stops the validator node and p2p networking
func (v *P2PValidator) Stop() {
	// Stop the base validator
	v.Validator.Stop()

	// Stop the p2p host if it's running
	if v.p2pHost != nil {
		if err := v.p2pHost.Close(); err != nil {
			log.Printf("Error closing p2p host: %v", err)
		}
	}
}

// handleMessage handles incoming messages from peers
func (v *P2PValidator) handleMessage(peerID peer.ID, msg p2p.Message) error {
	switch msg.Type {
	case p2p.MessageTypeStatus:
		return v.handleStatusMessage(peerID, msg)
	case p2p.MessageTypeProof:
		return v.handleProofMessage(peerID, msg)
	case p2p.MessageTypeSync:
		return v.handleSyncMessage(peerID, msg)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleStatusMessage processes a status message from a peer
func (v *P2PValidator) handleStatusMessage(peerID peer.ID, msg p2p.Message) error {
	// Parse the status data
	var statusData p2p.StatusData
	if err := json.Unmarshal(msg.Data, &statusData); err != nil {
		return fmt.Errorf("failed to unmarshal status data: %w", err)
	}

	// Update peer information
	v.peersMutex.Lock()
	defer v.peersMutex.Unlock()

	// Create or update peer info
	peer, exists := v.peers[peerID]
	if !exists {
		peer = &ValidatorPeer{
			ID: peerID,
		}
		v.peers[peerID] = peer
	}

	// Update peer data
	peer.Address = statusData.Address
	peer.Registered = statusData.Registered
	peer.LastBlockSeen = statusData.LastBlockSeen
	peer.ProofsSubmitted = statusData.ProofsSubmitted
	peer.LastSeen = msg.Timestamp.Unix()

	return nil
}

// handleProofMessage processes a proof message from a peer
func (v *P2PValidator) handleProofMessage(peerID peer.ID, msg p2p.Message) error {
	// Parse the proof data
	var proofData p2p.ProofData
	if err := json.Unmarshal(msg.Data, &proofData); err != nil {
		return fmt.Errorf("failed to unmarshal proof data: %w", err)
	}

	// Log the proof submission
	log.Printf("Peer %s submitted proof for farm %d with score %d (tx: %s, block: %d)",
		peerID.String(), proofData.FarmID, proofData.PerformanceScore,
		proofData.TxHash, proofData.BlockNumber)

	return nil
}

// handleSyncMessage processes a sync message from a peer
func (v *P2PValidator) handleSyncMessage(peerID peer.ID, msg p2p.Message) error {
	// Parse the sync data
	var syncData p2p.SyncData
	if err := json.Unmarshal(msg.Data, &syncData); err != nil {
		return fmt.Errorf("failed to unmarshal sync data: %w", err)
	}

	// TODO: Implement synchronization logic
	log.Printf("Received sync request from peer %s for blocks %d to %d",
		peerID.String(), syncData.FromBlock, syncData.ToBlock)

	return nil
}

// broadcastStatus periodically broadcasts the validator's status to all peers
func (v *P2PValidator) broadcastStatus(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if v.p2pProtocol == nil {
				continue
			}

			// Check if we're registered
			registered, err := v.IsRegistered()
			if err != nil {
				log.Printf("Error checking registration status: %v", err)
				continue
			}

			// Create and broadcast status message
			msg, err := p2p.CreateStatusMessage(
				v.address.Hex(),
				registered,
				v.lastBlockSeen,
				v.proofsSubmitted,
			)
			if err != nil {
				log.Printf("Error creating status message: %v", err)
				continue
			}

			if err := v.p2pProtocol.Broadcast(msg); err != nil {
				log.Printf("Error broadcasting status: %v", err)
			}
		}
	}
}

// GetPeers returns information about connected peers
func (v *P2PValidator) GetPeers() []*ValidatorPeer {
	v.peersMutex.RLock()
	defer v.peersMutex.RUnlock()

	peers := make([]*ValidatorPeer, 0, len(v.peers))
	for _, peer := range v.peers {
		peers = append(peers, peer)
	}

	return peers
}

// UpdateBlockProcessed updates the last block seen by this validator and broadcasts it to peers
func (v *P2PValidator) UpdateBlockProcessed(blockNum uint64) {
	v.lastBlockSeen = blockNum

	// If we have a protocol and it's a significant change, broadcast immediately
	if v.p2pProtocol != nil && blockNum%10 == 0 {
		// Check if we're registered
		registered, err := v.IsRegistered()
		if err != nil {
			log.Printf("Error checking registration status: %v", err)
			return
		}

		// Create and broadcast status message
		msg, err := p2p.CreateStatusMessage(
			v.address.Hex(),
			registered,
			v.lastBlockSeen,
			v.proofsSubmitted,
		)
		if err != nil {
			log.Printf("Error creating status message: %v", err)
			return
		}

		if err := v.p2pProtocol.Broadcast(msg); err != nil {
			log.Printf("Error broadcasting status update: %v", err)
		}
	}
}

// UpdateProofSubmitted increments the proofs submitted counter and broadcasts to peers
func (v *P2PValidator) UpdateProofSubmitted(farmID, performanceScore int64, txHash string, blockNumber uint64) {
	v.proofsSubmitted++

	// If we have a protocol, broadcast the proof submission
	if v.p2pProtocol != nil {
		// Create and broadcast proof message
		msg, err := p2p.CreateProofMessage(
			v.address.Hex(),
			farmID,
			performanceScore,
			txHash,
			blockNumber,
		)
		if err != nil {
			log.Printf("Error creating proof message: %v", err)
			return
		}

		if err := v.p2pProtocol.Broadcast(msg); err != nil {
			log.Printf("Error broadcasting proof submission: %v", err)
		}
	}
}

// syncPeersFromHost periodically syncs the peer list from the libp2p host
func (v *P2PValidator) syncPeersFromHost(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if v.p2pHost == nil {
				continue
			}
			
			// Get the list of connected peers from the host
			// This will also update the host's internal peer list
			connectedPeers := v.p2pHost.Peers()
			
			// Update our peer map
			v.peersMutex.Lock()
			
			// Add any new peers
			for _, peerID := range connectedPeers {
				if _, exists := v.peers[peerID]; !exists {
					// This is a new peer we haven't seen before
					v.peers[peerID] = &ValidatorPeer{
						ID:       peerID,
						Address:  peerID.String(),
						LastSeen: time.Now().Unix(),
					}
					log.Printf("Added new peer to tracking: %s", peerID.String())
				} else {
					// Update the last seen timestamp for existing peers
					v.peers[peerID].LastSeen = time.Now().Unix()
				}
			}
			
			// Remove peers that are no longer connected
			// We consider a peer disconnected if it's not in the connected list
			// or if we haven't seen it in the last 10 seconds
			currentTime := time.Now().Unix()
			for peerID, peer := range v.peers {
				found := false
				for _, connectedID := range connectedPeers {
					if peerID == connectedID {
						found = true
						break
					}
				}
				
				// Check if the peer is not in the connected list or if it's stale
				if !found || (currentTime - peer.LastSeen > 10) {
					// This peer is no longer connected or has timed out
					delete(v.peers, peerID)
					log.Printf("Peer disconnected: %s", peerID.String())
				}
			}
			
			v.peersMutex.Unlock()
		}
	}
}

// GetP2PStatus returns the status of the p2p network
func (v *P2PValidator) GetP2PStatus() map[string]interface{} {
	if v.p2pHost == nil {
		return map[string]interface{}{
			"running": false,
		}
	}

	peers := v.GetPeers()
	peerInfo := make([]map[string]interface{}, 0, len(peers))
	for _, p := range peers {
		peerInfo = append(peerInfo, map[string]interface{}{
			"id":              p.ID.String(),
			"address":         p.Address,
			"registered":      p.Registered,
			"lastBlockSeen":   p.LastBlockSeen,
			"proofsSubmitted": p.ProofsSubmitted,
		})
	}

	return map[string]interface{}{
		"running":   true,
		"nodeID":    v.p2pHost.ID().String(),
		"addresses": v.p2pHost.Addrs(),
		"peerCount": len(peers),
		"peers":     peerInfo,
	}
}
