package cmd

import (
	"cometkms/sigclient"
	"cometkms/signer"
	"cometkms/state"
	"context"
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

func RunApp(ctx context.Context, homeDir string, signerAddress string, signerRpc string, privKeyFilePath string) {

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

	///////// Setting configurations for the CometBFT node
	// Each tx is the last_signed_state, and should trigger a new block with create_empty_blocks=false
	config.Consensus.CreateEmptyBlocks = false
	// Since we want blocks created on each signRequest (which generates a TX submission) it should not wait for a timeout
	config.Consensus.TimeoutCommit = time.Millisecond * 50
	config.Consensus.TimeoutPrevote = time.Millisecond * 50
	config.Consensus.TimeoutPrevoteDelta = time.Hour * 24
	config.Consensus.TimeoutPrecommit = time.Millisecond * 50
	config.Consensus.TimeoutPrecommitDelta = time.Hour * 24
	config.Consensus.TimeoutPropose = time.Hour * 24
	// Since we should only every have one transaction in the mempool, we can set the size to 1
	// It is better to fail then have FIFO mempool not include signing state TX in the correct order and block.
	config.Mempool.Size = 1
	// Required for local testing.
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

	// Load the private key from the specified file
	privkey, _, err := sigclient.LoadKeyFromFile(privKeyFilePath)
	if err != nil {
		log.Fatalf("Failed to load key: %v", err)
	}
	log.Printf("Loaded private key from %s", privKeyFilePath)

	s, err := sigclient.SigClient(signerAddress, config.RPC.ListenAddress, privkey, privKeyFilePath, db)
	if err != nil {
		log.Fatal(err)
	}

	// Start the remote signer application
	log.Printf("Starting remote signer client at %s", signerAddress)
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
