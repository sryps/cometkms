package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cometkms/signer"
)

func main() {
	log.Println("Starting remote signer CometKMS...")

	// Define command line flags
	var addr string
	var keyFilePath string
	var stateFilePath string
	var help string
	flag.StringVar(&addr, "addr", "", "validator address to connect to (example: tcp://127.0.0.1:12345)")
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

	// Initialize the signer with the address, private key, and file paths
	s, err := signer.NewSigner(addr, privkey, keyFilePath, stateFilePath)
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
			if err := s.Run(ctx); err != nil {
				log.Fatal(err)
			}
		}
	}
}
