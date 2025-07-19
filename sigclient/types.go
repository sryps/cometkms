package sigclient

import (
	pbcrypto "github.com/cometbft/cometbft/api/cometbft/crypto/v1"
	pbtypes "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	"time"
)

// SimpleSigner is a struct that holds the configuration for a remote signer.
type SimpleSigner struct {
	addr          string
	privKey       cmted25519.PrivKey
	PubKey        PubKey
	stateFilePath string
	keyFilePath   string
}

type PubKey struct {
	PubKeyType    pbcrypto.PublicKey
	PubKeyBytes   []byte
	PubKeyTypeStr string
}

// SigningState is a struct that holds the state of the last signed state.
type SigningState struct {
	Type               pbtypes.SignedMsgType `json:"type"`     // 0: unknown, 1: prevote, 2: precommit, 32: proposal
	TypeStr            string                `json:"type_str"` // human-readable type
	Height             int64                 `json:"height"`
	Round              int32                 `json:"round"`
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
