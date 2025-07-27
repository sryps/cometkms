package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the signing state",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing CometBFT signer chain...")

		// Validate flags
		if initHomeDir == "" {
			initHomeDir = fmt.Sprintf("%s/.cometbft/chain-init", os.Getenv("HOME"))
			log.Fatalf("Must set CometKMS home init directory with --init-path (-p) example: %s\n", initHomeDir)
		}
		if numNodes == "" {
			log.Fatal("Number of nodes must be specified with --num-nodes (-n) flag")
		}

		// ensure cometbft binary is available
		if _, err := exec.Command("cometbft", "version").Output(); err != nil {
			log.Fatal("CometBFT binary not found. Please ensure CometBFT v0.38.x is installed and available in your PATH.")
		}

		if err := InitSignerCometBft(initHomeDir, numNodes); err != nil {
			fmt.Printf("Failed to initialize CometBFT signer: %v\n", err)
			return
		}
		fmt.Printf("Number of nodes initialized: %s\n", numNodes)
		for i := 0; i < len(numNodes); i++ {
			fmt.Printf("Node config %d initialized successfully at %s/node%d\n", i+1, initHomeDir, i+1)
		}
		fmt.Printf("You can now copy %s/nodeX to your CometBFT home directory where you are running the remote signer.", initHomeDir)
	},
}

var initHomeDir string
var numNodes string

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&initHomeDir, "init-path", "p", "", "Path to the CometBFT home directory (default: $HOME/.cometbft/chain-init)")
	initCmd.Flags().StringVarP(&numNodes, "num-nodes", "n", "", "Number of nodes to initialize in the CometKMS network")
}

func InitSignerCometBft(initHomeDir string, numNodes string) error {
	// Initialize the CometBFT home directory
	fmt.Println("Initializing CometBFT home directory:", initHomeDir)
	result, err := exec.Command(
		"cometbft", "testnet",
		"--o", initHomeDir,
		"--v", numNodes,
	).Output()
	if err != nil {
		return fmt.Errorf("failed to initialize CometBFT home directory: %v, output: %s", err, result)
	}
	return nil
}
