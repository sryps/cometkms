package main

import (
	"context"
	"flag"
	"log"

	"cometkms/signer"
)

func main() {
	log.Println("Starting remote signer CometKMS...")

	// Define command line flags
	var address string
	var addressBackup string
	var keyFilePath string
	var stateFilePath string
	var help string
	flag.StringVar(&address, "address", "", "primary validator address to connect to (example: tcp://127.0.0.1:12345)")
	flag.StringVar(&addressBackup, "address-backup", "", "secondary validator address to connect to for HA (example: tcp://127.0.0.1:54321)")
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
	if address == "" {
		log.Fatal("-address : Node address is required - use -address flag (example: tcp://127.0.0.1:12345)")
	}
	if addressBackup == "" {
		log.Printf("-address-backup : Node address is recommended for high availability, but not required. If not provided, the signer will only use the primary address.")
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
	secPubKey := signer.SetSecondaryPubkey()
	signer, err := signer.NewSigner(address, addressBackup, privkey, keyFilePath, stateFilePath, secPubKey)
	if err != nil {
		log.Fatal(err)
	}

	// Start the remote signer
	ctx := context.Background()
	if err := signer.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
