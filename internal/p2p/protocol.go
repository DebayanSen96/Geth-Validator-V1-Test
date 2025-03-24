package p2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// MessageType defines the type of message being sent between validators
type MessageType string

const (
	// MessageTypeStatus is sent when a validator's status changes
	MessageTypeStatus MessageType = "status"
	
	// MessageTypeProof is sent when a validator submits a proof
	MessageTypeProof MessageType = "proof"
	
	// MessageTypeSync is sent to request synchronization of data
	MessageTypeSync MessageType = "sync"
)

// Message represents a message sent between validators
type Message struct {
	Type      MessageType     `json:"type"`
	Sender    string          `json:"sender"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// StatusData contains validator status information
type StatusData struct {
	Address         string `json:"address"`
	Registered      bool   `json:"registered"`
	LastBlockSeen   uint64 `json:"lastBlockSeen"`
	ProofsSubmitted uint64 `json:"proofsSubmitted"`
}

// ProofData contains information about a submitted proof
type ProofData struct {
	FarmID          int64  `json:"farmId"`
	PerformanceScore int64  `json:"performanceScore"`
	TxHash          string `json:"txHash"`
	BlockNumber     uint64 `json:"blockNumber"`
}

// SyncData contains synchronization request information
type SyncData struct {
	FromBlock uint64 `json:"fromBlock"`
	ToBlock   uint64 `json:"toBlock"`
}

// MessageHandler defines a function that handles incoming messages
type MessageHandler func(peer.ID, Message) error

// Protocol manages the validator communication protocol
type Protocol struct {
	host          *Host
	protocolID    protocol.ID
	messageHandler MessageHandler
	mutex         sync.RWMutex
	peers         map[peer.ID]*bufio.ReadWriter
}

// NewProtocol creates a new validator protocol handler
func NewProtocol(host *Host, protocolID string, handler MessageHandler) *Protocol {
	p := &Protocol{
		host:          host,
		protocolID:    protocol.ID(protocolID),
		messageHandler: handler,
		peers:         make(map[peer.ID]*bufio.ReadWriter),
	}

	// Set stream handler for the protocol
	host.host.SetStreamHandler(p.protocolID, p.handleStream)

	return p
}

// handleStream is called when a new stream is opened with a peer
func (p *Protocol) handleStream(stream network.Stream) {
	// Get the peer ID
	peerID := stream.Conn().RemotePeer()
	log.Printf("New stream from peer: %s", peerID.String())

	// Create a buffered reader/writer
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	p.mutex.Lock()
	p.peers[peerID] = rw
	p.mutex.Unlock()

	// Start reading messages from the peer
	go p.readMessages(peerID, rw)
}

// readMessages continuously reads messages from a peer
func (p *Protocol) readMessages(peerID peer.ID, rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from peer %s: %v", peerID.String(), err)
			}
			
			// Remove the peer from our map
			p.mutex.Lock()
			delete(p.peers, peerID)
			p.mutex.Unlock()
			return
		}

		// Parse the message
		var msg Message
		if err := json.Unmarshal([]byte(str), &msg); err != nil {
			log.Printf("Error unmarshaling message from peer %s: %v", peerID.String(), err)
			continue
		}

		// Handle the message
		if err := p.messageHandler(peerID, msg); err != nil {
			log.Printf("Error handling message from peer %s: %v", peerID.String(), err)
		}
	}
}

// Broadcast sends a message to all connected peers
func (p *Protocol) Broadcast(msg Message) error {
	p.mutex.RLock()
	peers := make([]peer.ID, 0, len(p.peers))
	for peer := range p.peers {
		peers = append(peers, peer)
	}
	p.mutex.RUnlock()

	for _, peer := range peers {
		if err := p.SendMessage(peer, msg); err != nil {
			log.Printf("Error sending message to peer %s: %v", peer.String(), err)
		}
	}

	return nil
}

// SendMessage sends a message to a specific peer
func (p *Protocol) SendMessage(peerID peer.ID, msg Message) error {
	p.mutex.RLock()
	rw, ok := p.peers[peerID]
	p.mutex.RUnlock()

	if !ok {
		// Peer not connected, try to open a new stream
		stream, err := p.host.host.NewStream(context.Background(), peerID, p.protocolID)
		if err != nil {
			return fmt.Errorf("failed to open stream to peer %s: %w", peerID.String(), err)
		}

		rw = bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
		
		p.mutex.Lock()
		p.peers[peerID] = rw
		p.mutex.Unlock()
		
		// Start reading messages from the peer
		go p.readMessages(peerID, rw)
	}

	// Marshal the message
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write the message
	data = append(data, '\n')
	if _, err := rw.Write(data); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	// Flush the writer
	if err := rw.Flush(); err != nil {
		return fmt.Errorf("failed to flush message: %w", err)
	}

	return nil
}

// CreateStatusMessage creates a new status message
func CreateStatusMessage(address string, registered bool, lastBlockSeen, proofsSubmitted uint64) (Message, error) {
	statusData := StatusData{
		Address:         address,
		Registered:      registered,
		LastBlockSeen:   lastBlockSeen,
		ProofsSubmitted: proofsSubmitted,
	}

	dataBytes, err := json.Marshal(statusData)
	if err != nil {
		return Message{}, err
	}

	return Message{
		Type:      MessageTypeStatus,
		Sender:    address,
		Timestamp: time.Now(),
		Data:      dataBytes,
	}, nil
}

// CreateProofMessage creates a new proof message
func CreateProofMessage(sender string, farmID, performanceScore int64, txHash string, blockNumber uint64) (Message, error) {
	proofData := ProofData{
		FarmID:          farmID,
		PerformanceScore: performanceScore,
		TxHash:          txHash,
		BlockNumber:     blockNumber,
	}

	dataBytes, err := json.Marshal(proofData)
	if err != nil {
		return Message{}, err
	}

	return Message{
		Type:      MessageTypeProof,
		Sender:    sender,
		Timestamp: time.Now(),
		Data:      dataBytes,
	}, nil
}

// CreateSyncMessage creates a new sync message
func CreateSyncMessage(sender string, fromBlock, toBlock uint64) (Message, error) {
	syncData := SyncData{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
	}

	dataBytes, err := json.Marshal(syncData)
	if err != nil {
		return Message{}, err
	}

	return Message{
		Type:      MessageTypeSync,
		Sender:    sender,
		Timestamp: time.Now(),
		Data:      dataBytes,
	}, nil
}
