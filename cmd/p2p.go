package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dexponent/geth-validator/internal/config"
	"github.com/dexponent/geth-validator/internal/p2p"
	"github.com/dexponent/geth-validator/internal/validator"
	"github.com/multiformats/go-multiaddr"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	listenAddresses []string
	bootstrapPeers  []string
)

// P2P commands
var p2pCmd = &cobra.Command{
	Use:   "p2p",
	Short: "P2P network commands for validator nodes",
	Long:  "Commands for managing the peer-to-peer network of validator nodes",
}

var p2pStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the validator with P2P networking enabled",
	Run: func(cmd *cobra.Command, args []string) {
		startP2PValidator()
	},
}

var p2pStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of the P2P network",
	Run: func(cmd *cobra.Command, args []string) {
		showP2PStatus()
	},
}

var p2pConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure P2P network settings",
	Run: func(cmd *cobra.Command, args []string) {
		configureP2P()
	},
}

func init() {
	// Add p2p command to the root command
	RootCmd.AddCommand(p2pCmd)

	// Add subcommands to the p2p command
	p2pCmd.AddCommand(p2pStartCmd)
	p2pCmd.AddCommand(p2pStatusCmd)
	p2pCmd.AddCommand(p2pConfigCmd)

	// Add flags for p2p configuration
	p2pConfigCmd.Flags().StringSliceVarP(&listenAddresses, "listen", "l", []string{"/ip4/0.0.0.0/tcp/9000", "/ip4/0.0.0.0/udp/9000/quic-v1"}, "Addresses to listen on")
	p2pConfigCmd.Flags().StringSliceVarP(&bootstrapPeers, "bootstrap", "b", []string{}, "Bootstrap peers to connect to")
}

// startP2PValidator starts a validator node with P2P networking enabled
func startP2PValidator() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create a new P2P validator
	val, err := validator.NewP2PValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create P2P validator: %v", err)
	}

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the validator with P2P networking
	log.Println("Starting validator with P2P networking...")
	if err := val.Start(ctx, 15); err != nil {
		log.Fatalf("Failed to start validator: %v", err)
	}

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Create a tabular UI for displaying validator status
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Clear the screen
				fmt.Print("\033[H\033[2J")

				// Get validator status
				status := val.GetP2PStatus()

				// Display node information
				fmt.Println("=== DXP Validator Node with P2P Networking ===")
				fmt.Printf("Node ID: %s\n", status["nodeID"])
				fmt.Println("\nP2P Addresses:")
				addrs, ok := status["addresses"].([]multiaddr.Multiaddr)
				if ok {
					for _, addr := range addrs {
						fmt.Printf("  %s\n", addr.String())
					}
				} else {
					fmt.Println("  No addresses available")
				}

				// Display peer information
				fmt.Printf("\nConnected Peers: %d\n", status["peerCount"])
				if status["peerCount"].(int) > 0 {
					table := tablewriter.NewWriter(os.Stdout)
					table.SetHeader([]string{"Peer ID", "Address", "Registered", "Last Block", "Proofs"})
					table.SetBorder(false)
					table.SetColumnColor(
						tablewriter.Colors{tablewriter.FgHiBlueColor},
						tablewriter.Colors{tablewriter.FgHiWhiteColor},
						tablewriter.Colors{tablewriter.FgHiGreenColor},
						tablewriter.Colors{tablewriter.FgHiYellowColor},
						tablewriter.Colors{tablewriter.FgHiCyanColor},
					)

					peers, ok := status["peers"].([]map[string]interface{})
					if !ok {
						fmt.Println("  No peer information available")
					} else {
						for _, peer := range peers {
							registeredStr := "No"
							if reg, ok := peer["registered"].(bool); ok && reg {
								registeredStr = "Yes"
							}

							// Safely extract peer ID
							peerID := "Unknown"
							if id, ok := peer["id"].(string); ok && len(id) > 12 {
								peerID = id[:12] + "..."
							} else if id, ok := peer["id"].(string); ok {
								peerID = id
							}

							// Safely extract address
							address := "Unknown"
							if addr, ok := peer["address"].(string); ok {
								address = addr
							}

							// Safely extract block number
							lastBlock := "0"
							if block, ok := peer["lastBlockSeen"].(uint64); ok {
								lastBlock = fmt.Sprintf("%d", block)
							} else if block, ok := peer["lastBlockSeen"].(float64); ok {
								lastBlock = fmt.Sprintf("%d", int(block))
							} else if block, ok := peer["lastBlockSeen"].(int); ok {
								lastBlock = fmt.Sprintf("%d", block)
							}

							// Safely extract proofs submitted
							proofs := "0"
							if p, ok := peer["proofsSubmitted"].(uint64); ok {
								proofs = fmt.Sprintf("%d", p)
							} else if p, ok := peer["proofsSubmitted"].(float64); ok {
								proofs = fmt.Sprintf("%d", int(p))
							} else if p, ok := peer["proofsSubmitted"].(int); ok {
								proofs = fmt.Sprintf("%d", p)
							}

							table.Append([]string{
								peerID,
								address,
								registeredStr,
								lastBlock,
								proofs,
							})
						}
					}

					table.Render()
				}

				// Display help information
				fmt.Println("\nPress Ctrl+C to stop the validator")
			}
		}
	}()

	// Wait for termination signal
	<-sigCh
	log.Println("Shutting down validator...")
	val.Stop()
}

// showP2PStatus displays the current P2P network status
func showP2PStatus() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Load P2P configuration
	p2pConfig, err := p2p.LoadP2PConfig(cfg.DataDir)
	if err != nil {
		log.Fatalf("Failed to load P2P configuration: %v", err)
	}

	// Display configuration
	fmt.Println("=== P2P Network Configuration ===")
	fmt.Println("Listen Addresses:")
	for _, addr := range p2pConfig.ListenAddresses {
		fmt.Printf("  %s\n", addr)
	}

	fmt.Println("\nBootstrap Peers:")
	if len(p2pConfig.BootstrapPeers) == 0 {
		fmt.Println("  None configured")
	} else {
		for _, peer := range p2pConfig.BootstrapPeers {
			fmt.Printf("  %s\n", peer)
		}
	}

	// TODO: Connect to the running validator to get real-time status
	fmt.Println("\nTo see live P2P network status, run the validator with 'dxp-validator p2p start'")
}

// configureP2P updates the P2P network configuration
func configureP2P() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Load existing P2P configuration
	p2pConfig, err := p2p.LoadP2PConfig(cfg.DataDir)
	if err != nil {
		log.Fatalf("Failed to load P2P configuration: %v", err)
	}

	// Update configuration with command-line flags
	if len(listenAddresses) > 0 {
		p2pConfig.ListenAddresses = listenAddresses
	}

	if len(bootstrapPeers) > 0 {
		p2pConfig.BootstrapPeers = bootstrapPeers
	}

	// Save the updated configuration
	if err := p2p.SaveP2PConfig(cfg.DataDir, p2pConfig); err != nil {
		log.Fatalf("Failed to save P2P configuration: %v", err)
	}

	fmt.Println("P2P configuration updated successfully")
	showP2PStatus()
}
