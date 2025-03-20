package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dexponent/geth-validator/internal/config"
	"github.com/dexponent/geth-validator/internal/validator"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the validator node",
	Long:  `Start the GETH-based validator node to participate in the Dexponent protocol.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags
		blockPollingInterval, _ := cmd.Flags().GetInt("block-polling-interval")
		detached, _ := cmd.Flags().GetBool("detached")

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			os.Exit(1)
		}

		// Create validator instance
		validatorNode, err := validator.NewValidator(cfg)
		if err != nil {
			fmt.Printf("Error creating validator: %v\n", err)
			os.Exit(1)
		}

		// Check if validator is registered
		isRegistered, err := validatorNode.IsRegistered()
		if err != nil {
			fmt.Printf("Error checking registration status: %v\n", err)
			os.Exit(1)
		}

		if !isRegistered {
			fmt.Println("Validator is not registered. Attempting to register...")
			txHash, err := validatorNode.RegisterValidator()
			if err != nil {
				fmt.Printf("Error registering validator: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Validator registered successfully! TX: %s\n", txHash)
		} else {
			fmt.Println("Validator is already registered with the DXP contract.")
		}

		// Create context with cancellation
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start the validator
		fmt.Println("Starting validator node...")
		if err := validatorNode.Start(ctx, blockPollingInterval); err != nil {
			fmt.Printf("Error starting validator: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Validator node started successfully!")

		// Handle graceful shutdown if not in detached mode
		if !detached {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)

			// Block until we receive a signal
			<-c

			fmt.Println("\nStopping validator node...")
			validatorNode.Stop()
			fmt.Println("Validator node stopped.")
		}
	},
}

func init() {
	startCmd.Flags().Int("block-polling-interval", 10, "Interval in seconds to poll for new blocks")
	startCmd.Flags().Bool("detached", false, "Run the validator in detached mode")
	startCmd.Flags().String("log-file", "", "Log file to write validator logs to")
}
