package cmd

import (
	"fmt"
	"os"

	"github.com/dexponent/geth-validator/internal/config"
	"github.com/dexponent/geth-validator/internal/validator"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of the validator node",
	Long:  `Check the status of the validator node, including registration, block processing, and more.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			os.Exit(1)
		}

		// Get validator status
		status, err := validator.GetValidatorStatus(cfg)
		if err != nil {
			fmt.Printf("Error getting validator status: %v\n", err)
			os.Exit(1)
		}

		// Print status information
		fmt.Println("Validator Node Status")
		fmt.Println("=====================")
		fmt.Printf("Running: %v\n", status.Running)
		fmt.Printf("Node ID: %s\n", status.NodeID)
		fmt.Printf("Account: %s\n", status.Account)
		fmt.Printf("ETH Balance: %.6f ETH\n", status.Balance)
		fmt.Printf("Registered: %v\n", status.Registered)
		fmt.Printf("Last Block Processed: %d\n", status.LastBlockProcessed)
		fmt.Printf("Verification Queue: %d\n", status.VerificationQueueSize)
		fmt.Printf("Consensus Participants: %d\n", status.ConsensusParticipants)
	},
}
