package signer

import (
	"time"

	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtp2pconn "github.com/cometbft/cometbft/p2p/conn"
	pbcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	pbtypes "github.com/cometbft/cometbft/proto/tendermint/types"
)

// SimpleSigner is a struct that holds the configuration for a remote signer.
type SimpleSigner struct {
	address           string             // Primary address of the validator
	addressBackup     string             // Backup address for high availability
	privKey           cmted25519.PrivKey // Private key of the validator
	PubKey            pbcrypto.PublicKey // Public key of the validator
	stateFilePath     string             // Path to the state file for storing signing state
	keyFilePath       string             // Path to the file of the private key ie. "priv_validator_key.json"
	connectionManager *ConnectionManager // Connection manager for handling connections
	secondaryPubkey   pbcrypto.PublicKey // Public key to use for the secondary connection that isnt signing
}

type ConnectionManager struct {
	isPrimaryConnActiveSigner   bool // Flag to indicate if the primary connection is active signer
	isSecondaryConnActiveSigner bool // Flag to indicate if the secondary connection is active signer
	primaryConn                 *cmtp2pconn.SecretConnection
	secondaryConn               *cmtp2pconn.SecretConnection
}

// SigningState is a struct that holds the state of the last signed state.
type SigningState struct {
	Height             int64                 `json:"height"`
	Round              int32                 `json:"round"`
	TypeStr            string                `json:"type_str"` // human-readable type
	Type               pbtypes.SignedMsgType `json:"type"`     // 0: unknown, 1: prevote, 2: precommit, 32: proposal
	BlockID            BlockID               `json:"block_id"`
	ValidatorAddress   []byte                `json:"validator_address"`
	Timestamp          time.Time             `json:"timestamp"`
	Signature          []byte                `json:"signature"`
	ExtensionSignature []byte                `json:"extension_signature"`
	ChainId            string                `json:"chain_id"`
}
type BlockID struct {
	BlockHash     []byte        `json:"block_hash"`
	PartSetHeader PartSetHeader `json:"part_set_header"`
}
type PartSetHeader struct {
	Hash  []byte `json:"hash"`
	Total uint32 `json:"total"`
}
