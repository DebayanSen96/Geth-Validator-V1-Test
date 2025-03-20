package cmd

import (
	"fmt"
	"os"

	"github.com/dexponent/geth-validator/internal/config"
	"github.com/dexponent/geth-validator/internal/validator"
	"github.com/spf13/cobra"
)

// rewardsCmd represents the rewards command
var rewardsCmd = &cobra.Command{
	Use:   "rewards",
	Short: "Check accumulated rewards for the validator",
	Long:  `Check accumulated rewards for the validator from successful verifications.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			os.Exit(1)
		}

		// Get validator rewards
		rewards, err := validator.GetValidatorRewards(cfg)
		if err != nil {
			fmt.Printf("Error getting validator rewards: %v\n", err)
			os.Exit(1)
		}

		// Print rewards information
		fmt.Printf("Pending rewards: %.6f DXP\n", rewards)
	},
}
