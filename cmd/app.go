package cmd

import (
	"cometkms/sigclient"
	"cometkms/signer"
	"cometkms/state"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"

	cfg "github.com/cometbft/cometbft/config"
	cmtflags "github.com/cometbft/cometbft/libs/cli/flags"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	nm "github.com/cometbft/cometbft/node"
	"github.com/dgraph-io/badger/v4"
	"github.com/spf13/viper"
)

var homeDir string

func init() {
	flag.StringVar(&homeDir, "cmt-home", "", "Path to the CometBFT config directory (if empty, uses $HOME/.cometbft)")
}

func RunApp(ctx context.Context) {
	// Parse command line flags
	flag.Parse()
	if homeDir == "" {
		homeDir = os.ExpandEnv("$HOME/.cometbft")
	}
	if err := InitCometBFT(homeDir); err != nil {
		log.Fatalf("Failed to initialize CometBFT: %v", err)
	}

	// Set up the CometBFT configuration
	config := cfg.DefaultConfig()

	config.SetRoot(homeDir)
	viper.SetConfigFile(fmt.Sprintf("%s/%s", homeDir, "config/config.toml"))

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Reading config: %v", err)
	}
	if err := viper.Unmarshal(config); err != nil {
		log.Fatalf("Decoding config: %v", err)
	}
	if err := config.ValidateBasic(); err != nil {
		log.Fatalf("Invalid configuration data: %v", err)
	}

	//// Setting configurations for the CometBFT node
	// Should only allow one transaction in the mempool.
	// Each tx is the last_signed_state, and should trigger a new block with create_empty_blocks=false
	config.Mempool.Size = 1
	config.Consensus.CreateEmptyBlocks = false
	config.Consensus.TimeoutCommit = time.Millisecond * 1000
	config.P2P.AllowDuplicateIP = true

	// Set the home directory for the BadgerDB AppState database
	dbPath := filepath.Join(homeDir, "badger")
	db, err := badger.Open(badger.DefaultOptions(dbPath))

	if err != nil {
		log.Fatalf("Opening database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Closing database: %v", err)
		}
	}()

	// Initialize the CometBFT application
	app := signer.NewSigner(db)
	if err != nil {
		log.Fatalf("failed to load metadata: %v", err)
	}
	log.Printf("Loaded metadata: height=%d, hash=%x", app.AppHeight, app.AppHash)

	pv := privval.LoadFilePV(
		config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(),
	)
	pubkey, err := pv.GetPubKey()
	state.ProposerPubKey = pubkey.Bytes()

	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		log.Fatalf("failed to load node's key: %v", err)
	}

	logger := cmtlog.NewTMLogger(cmtlog.NewSyncWriter(os.Stdout))
	logger, err = cmtflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel)

	if err != nil {
		log.Fatalf("failed to parse log level: %v", err)
	}

	// Create the CometBFT node
	node, err := nm.NewNode(
		config,
		pv,
		nodeKey,
		proxy.NewLocalClientCreator(app),
		nm.DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		nm.DefaultMetricsProvider(config.Instrumentation),
		logger,
	)

	if err != nil {
		log.Fatalf("Creating node: %v", err)
	}

	// Setup the remote signer client
	var addr string
	var keyFilePath string
	var help string
	addr = "tcp://127.0.0.1:12345" // Default address
	keyFilePath = "priv_validator_key.json"
	if os.Getenv("SIGNER_ADDR") != "" {
		addr = os.Getenv("SIGNER_ADDR")
	}
	if os.Getenv("SIGNER_KEY_FILE") != "" {
		keyFilePath = os.Getenv("SIGNER_KEY_FILE")
	}

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
	privkey, _, err := sigclient.LoadKeyFromFile(keyFilePath)
	if err != nil {
		log.Fatalf("Failed to load key: %v", err)
	}

	s, err := sigclient.SigClient(addr, config.RPC.ListenAddress, privkey, keyFilePath, db)
	if err != nil {
		log.Fatal(err)
	}

	// Start the remote signer application
	log.Printf("Starting remote signer client at %s", addr)
	go s.Run(ctx)
	log.Printf("Remote signer client started successfully")

	// Start the CometBFT node
	if err := node.Start(); err != nil {
		log.Fatalf("Starting node: %v", err)
	}
	defer func() {
		node.Stop()
		node.Wait()
	}()

	<-ctx.Done()
	log.Println("Received shutdown signal, stopping all services...")
}

func InitCometBFT(homedir string) error {
	// Run cometbft init temp command if the home directory does not exist
	if _, err := os.Stat(homedir); os.IsNotExist(err) {
		cmd := exec.Command("cometbft", "init", "temp", "--home", homedir)
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Command failed: %v", err)
		}
		log.Printf("Initialized CometBFT home directory at %s", homedir)
	}
	return nil
}
