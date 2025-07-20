package sigclient

import (
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	pbcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	"github.com/dgraph-io/badger/v4"
)

// SimpleSigner is a struct that holds the configuration for a remote signer.
type SimpleSigner struct {
	addr        string
	RPCaddr     string
	privKey     cmted25519.PrivKey
	PubKey      pbcrypto.PublicKey
	keyFilePath string
	db          *badger.DB
}
