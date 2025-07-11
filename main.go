package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cometkms/signer"

	pbcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
)

func main() {
	log.Println("Starting remote signer CometKMS...")

	// Define command line flags
	var addr string
	var secondaryAddr string
	var keyFilePath string
	var stateFilePath string
	var help string
	flag.StringVar(&addr, "addr", "", "validator address to connect to (example: tcp://127.0.0.1:12345)")
	flag.StringVar(&secondaryAddr, "addr-backup", "", "backup validator address to connect to (example: tcp://127.0.0.1:54321)")
	flag.StringVar(&keyFilePath, "privkey", "priv_validator_key.json", "path to private key file")
	flag.StringVar(&stateFilePath, "statefile", "priv_validator_state.json", "path to signing state file")
	flag.StringVar(&help, "help", "", "show help message")
	flag.Parse()

	// If help is requested, show usage and exit
	if help != "" {
		flag.Usage()
		return
	}

	// Validate required flags
	if addr == "" {
		log.Fatal("Node address is required - use -addr flag (example: tcp://127.0.0.1:12345)")
	}
	if secondaryAddr == "" {
		log.Printf("No backup address provided, using primary address %s", addr)
	}

	// Load the private key from the specified file
	log.Printf("Loading private key from %s", keyFilePath)
	privkey, _, err := signer.LoadKeyFromFile(keyFilePath)
	if err != nil {
		log.Fatalf("Failed to load key: %v", err)
	}

	// Create the state file if it does not exist
	if err := signer.CreateStateFileIfNoneExists(stateFilePath); err != nil {
		log.Fatalf("Failed to create state file: %v", err)
	}

	// Create nonsigning public key
	var emptyPubkey pbcrypto.PublicKey
	nonsigningPrivkey, nonsigningPubkey := signer.SetSecondaryKeys()
	if nonsigningPrivkey == nil {
		log.Fatal("Failed to set secondary private key")
	}
	if nonsigningPubkey == emptyPubkey {
		log.Fatal("Failed to set secondary public key")
	}

	// Initialize the signer with the address, private key, and file paths
	signer, err := signer.NewSigner(addr, secondaryAddr, privkey, nonsigningPrivkey, nonsigningPubkey, keyFilePath, stateFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// Start the remote signer
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			log.Println("Received termination signal, shutting down signer...")
			return
		default:
			log.Println("Starting main signer loop...")
			if err := signer.Run(ctx); err != nil {
				log.Fatal(err)
			}
		}
	}
}
