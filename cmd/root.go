package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cometkms",
	Short: "CometKMS is a remote signer for CometBFT",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting main app...")
		RunApp()
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
