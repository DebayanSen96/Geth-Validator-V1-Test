package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dxp-validator",
	Short: "Dexponent GETH-based validator node",
	Long: `A GETH-based validator node for the Dexponent protocol.

This validator allows verifiers to register, start verification processes,
run a consensus mechanism, and submit proofs to smart contracts deployed
on the Base chain.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, print help
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Load environment variables from .env file if it exists
	envFile := filepath.Join(".env")
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err != nil {
			fmt.Printf("Warning: Error loading .env file: %v\n", err)
		}
	}

	// Add persistent flags that will be global for all subcommands
	RootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is .env)")
	RootCmd.PersistentFlags().StringP("log-level", "l", "info", "log level (debug, info, warn, error)")

	// Initialize subcommands
	RootCmd.AddCommand(startCmd)
	RootCmd.AddCommand(stopCmd)
	RootCmd.AddCommand(statusCmd)
	RootCmd.AddCommand(rewardsCmd)
	RootCmd.AddCommand(claimCmd)
	// Note: Contract commands are added in the contract.go file
}
