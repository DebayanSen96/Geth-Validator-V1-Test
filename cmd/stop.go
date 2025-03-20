package cmd

import (
	"fmt"
	"os"

	"github.com/dexponent/geth-validator/internal/config"
	"github.com/dexponent/geth-validator/internal/validator"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the validator node",
	Long:  `Stop a running validator node.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			os.Exit(1)
		}

		// Stop the validator
		if err := validator.StopValidator(cfg); err != nil {
			fmt.Printf("Error stopping validator: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Validator node stopped.")
	},
}
