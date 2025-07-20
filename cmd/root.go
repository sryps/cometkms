package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cometkms",
	Short: "CometKMS is a remote signer for CometBFT",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting main app...")

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		RunApp(ctx)
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
