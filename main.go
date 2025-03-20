package main

import (
	"fmt"
	"os"

	"github.com/dexponent/geth-validator/cmd"
)

func main() {
	// Execute the root command
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
