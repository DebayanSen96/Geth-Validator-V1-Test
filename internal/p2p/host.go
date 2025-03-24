package p2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = 1 * time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other validators.
const DiscoveryServiceTag = "dxp-validators"

// Config holds the p2p configuration.
type Config struct {
	ListenAddresses []string
	BootstrapPeers  []string
	PrivateKeyFile  string
}

// Host represents the p2p network host.
type Host struct {
	host     host.Host
	config   Config
	protocol string
	mutex    sync.RWMutex
	peers    map[peer.ID]peer.AddrInfo
}

// NewHost creates a new p2p host with the given configuration.
func NewHost(ctx context.Context, config Config, protocol string) (*Host, error) {
	// Generate or load a private key
	priv, err := generateOrLoadPrivateKey(config.PrivateKeyFile)
	if err != nil {
		return nil, err
	}

	// Parse the multiaddresses to listen on
	addrs := make([]multiaddr.Multiaddr, 0, len(config.ListenAddresses))
	for _, addrStr := range config.ListenAddresses {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			return nil, fmt.Errorf("invalid listen address: %s: %w", addrStr, err)
		}
		addrs = append(addrs, addr)
	}

	// Create the libp2p host
	h, err := libp2p.New(
		libp2p.ListenAddrs(addrs...),
		libp2p.Identity(priv),
		libp2p.NATPortMap(),
		libp2p.EnableRelay(),
	)
	if err != nil {
		return nil, err
	}

	host := &Host{
		host:     h,
		config:   config,
		protocol: protocol,
		peers:    make(map[peer.ID]peer.AddrInfo),
	}

	// Subscribe to network notifications for peer connections/disconnections
	notifyBundle := &network.NotifyBundle{
		DisconnectedF: func(n network.Network, conn network.Conn) {
			peerID := conn.RemotePeer()
			log.Printf("Network notification: Peer disconnected: %s", peerID.String())
			
			// Remove the peer from our tracking map
			host.mutex.Lock()
			delete(host.peers, peerID)
			host.mutex.Unlock()
		},
		ConnectedF: func(n network.Network, conn network.Conn) {
			peerID := conn.RemotePeer()
			log.Printf("Network notification: Peer connected: %s", peerID.String())
			
			// Add the peer to our tracking map
			host.mutex.Lock()
			host.peers[peerID] = peer.AddrInfo{ID: peerID}
			host.mutex.Unlock()
		},
	}
	
	// Register the notification handlers
	h.Network().Notify(notifyBundle)

	// Set up local mDNS discovery
	if err := host.setupDiscovery(ctx); err != nil {
		return nil, err
	}

	// Connect to bootstrap peers
	if err := host.connectToBootstrapPeers(ctx); err != nil {
		log.Printf("Warning: failed to connect to some bootstrap peers: %v", err)
	}

	return host, nil
}

// generateOrLoadPrivateKey generates a new private key or loads an existing one.
func generateOrLoadPrivateKey(keyFile string) (crypto.PrivKey, error) {
	// TODO: Implement loading from file if keyFile is provided
	// For now, just generate a new key
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, -1, rand.Reader)
	if err != nil {
		return nil, err
	}
	return priv, nil
}

// setupDiscovery configures mDNS discovery to find other validators on the local network.
func (h *Host) setupDiscovery(ctx context.Context) error {
	// Setup local mDNS discovery
	discovery := mdns.NewMdnsService(h.host, DiscoveryServiceTag, h)
	return discovery.Start()
}

// HandlePeerFound is called when a peer is discovered via mDNS.
func (h *Host) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == h.host.ID() {
		return // Skip ourselves
	}

	log.Printf("Discovered new peer: %s", pi.ID.String())

	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.peers[pi.ID] = pi

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.host.Connect(ctx, pi); err != nil {
		log.Printf("Failed to connect to peer %s: %v", pi.ID.String(), err)
		return
	}

	log.Printf("Connected to peer: %s", pi.ID.String())
}

// connectToBootstrapPeers connects to the configured bootstrap peers.
func (h *Host) connectToBootstrapPeers(ctx context.Context) error {
	var wg sync.WaitGroup
	var errors []error
	var errorsMutex sync.Mutex

	for _, addrStr := range h.config.BootstrapPeers {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			return fmt.Errorf("invalid bootstrap peer address: %s: %w", addrStr, err)
		}

		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return fmt.Errorf("invalid peer info from address: %s: %w", addrStr, err)
		}

		wg.Add(1)
		go func(pi peer.AddrInfo) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			h.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
			if err := h.host.Connect(ctx, pi); err != nil {
				errorsMutex.Lock()
				errors = append(errors, fmt.Errorf("failed to connect to bootstrap peer %s: %w", pi.ID.String(), err))
				errorsMutex.Unlock()
				return
			}

			log.Printf("Connected to bootstrap peer: %s", pi.ID.String())
		}(*peerInfo)
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("failed to connect to some bootstrap peers: %v", errors)
	}

	return nil
}

// ID returns the host's peer ID.
func (h *Host) ID() peer.ID {
	return h.host.ID()
}

// Addrs returns the host's listen addresses.
func (h *Host) Addrs() []multiaddr.Multiaddr {
	return h.host.Addrs()
}

// Peers returns the list of connected peers.
func (h *Host) Peers() []peer.ID {
	// Get the list of peers that libp2p considers connected
	connectedPeers := h.host.Network().Peers()
	
	// Update our internal peer map based on actual connections
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	// Remove any peers that are no longer connected
	for id := range h.peers {
		found := false
		for _, connectedID := range connectedPeers {
			if id == connectedID {
				found = true
				break
			}
		}
		
		if !found {
			// This peer is no longer connected according to libp2p
			delete(h.peers, id)
			log.Printf("Peer disconnected: %s", id.String())
		}
	}
	
	// Add new peers to our tracking map
	for _, id := range connectedPeers {
		// Make sure the peer is in our map
		if _, exists := h.peers[id]; !exists {
			// Add any newly connected peers to our map
			h.peers[id] = peer.AddrInfo{ID: id}
			log.Printf("New peer connected: %s", id.String())
		}
	}
	
	return connectedPeers
}

// Close shuts down the host.
func (h *Host) Close() error {
	return h.host.Close()
}
