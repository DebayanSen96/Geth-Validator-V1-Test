package cmd

import (
	"fmt"
	"os"

	"github.com/dexponent/geth-validator/internal/config"
	"github.com/dexponent/geth-validator/internal/validator"
	"github.com/spf13/cobra"
)

// claimCmd represents the claim command
var claimCmd = &cobra.Command{
	Use:   "claim",
	Short: "Claim accumulated rewards for the validator",
	Long:  `Claim accumulated rewards for the validator from successful verifications.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			os.Exit(1)
		}

		// Check pending rewards first
		rewards, err := validator.GetValidatorRewards(cfg)
		if err != nil {
			fmt.Printf("Error getting validator rewards: %v\n", err)
			os.Exit(1)
		}

		if rewards <= 0 {
			fmt.Println("No rewards to claim.")
			return
		}

		fmt.Printf("Claiming %.6f DXP tokens in rewards...\n", rewards)

		// Claim rewards
		txHash, err := validator.ClaimValidatorRewards(cfg)
		if err != nil {
			fmt.Printf("Error claiming rewards: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Transaction sent: %s\n", txHash)
		fmt.Println("Rewards claimed successfully!")
	},
}
