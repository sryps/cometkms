package main

import (
	"cometkms/sigclient"
	"cometkms/signer"
	"context"
	"flag"
	"fmt"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

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

func main() {
	flag.Parse()
	if homeDir == "" {
		homeDir = os.ExpandEnv("$HOME/.cometbft")
	}

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
	config.Consensus.TimeoutCommit = time.Millisecond * 1000
	config.Consensus.CreateEmptyBlocks = false
	//config.Consensus.TimeoutPropose = time.Hour * 24
	config.Mempool.Size = 1

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

	app := signer.NewSigner(db)
	if err != nil {
		log.Fatalf("failed to load metadata: %v", err)
	}
	log.Printf("Loaded metadata: height=%d, hash=%x", app.AppHeight, app.AppHash)

	pv := privval.LoadFilePV(
		config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(),
	)

	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		log.Fatalf("failed to load node's key: %v", err)
	}

	logger := cmtlog.NewTMLogger(cmtlog.NewSyncWriter(os.Stdout))
	logger, err = cmtflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel)

	if err != nil {
		log.Fatalf("failed to parse log level: %v", err)
	}

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
	var RPCaddr string
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
	if os.Getenv("SIGNER_RPC") != "" {
		RPCaddr = os.Getenv("SIGNER_RPC")
	} else {
		log.Fatal("SIGNER_RPC environment variable is required, this address is for broadcasting Txs")
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

	s, err := sigclient.SigClient(addr, RPCaddr, privkey, keyFilePath, db)
	if err != nil {
		log.Fatal(err)
	}
	// Start the remote signer
	log.Printf("Starting remote signer client at %s", addr)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	go s.Run(ctx)

	log.Printf("Remote signer client started successfully")

	node.Start()
	defer func() {
		node.Stop()
		node.Wait()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

}
