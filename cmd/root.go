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

		// Default values for flags
		if homeDir == "" {
			homeDir = os.Getenv("HOME") + "/.cometbft"
		}
		if signerAddress == "" {
			signerAddress = "tcp://127.0.1:12345"
		}
		if signerRpc == "" {
			signerRpc = "http://127.0.1:26657"
		}
		if privKeyFilePath == "" {
			privKeyFilePath = homeDir + "/priv_validator_key.json"
		}
		if err := InitCometBFT(homeDir); err != nil {
			fmt.Printf("Failed to initialize CometBFT home directory: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Using CometBFT home directory: %s\n", homeDir)
		fmt.Printf("Using remote signer address: %s\n", signerAddress)
		fmt.Printf("Using remote signer RPC: %s\n", signerRpc)
		fmt.Printf("Using private key file: %s\n", privKeyFilePath)

		RunApp(ctx, homeDir, signerAddress, signerRpc, privKeyFilePath)
	},
}

var homeDir string
var signerAddress string
var signerRpc string
var privKeyFilePath string

func init() {
	rootCmd.Flags().StringVarP(&homeDir, "cmt-home", "c", "", "Path to the CometBFT config directory (if empty, uses $HOME/.cometbft)")
	rootCmd.Flags().StringVarP(&signerAddress, "signer-addr", "a", "", "Address of the remote signer (example: tcp://127.0.0.1:12345)")
	rootCmd.Flags().StringVarP(&signerRpc, "signer-rpc", "r", "", "RPC address of the remote signer RPC for TX submission (example: http://127.0.0.1:26657)")
	rootCmd.Flags().StringVarP(&privKeyFilePath, "signer-key-file", "k", "", "Path to the private key file for the remote signer (default: $HOME/.cometbft/priv_validator_key.json)")
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
