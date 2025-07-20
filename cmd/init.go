package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the signing state",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running InitChain...")
		// InitChain logic
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
