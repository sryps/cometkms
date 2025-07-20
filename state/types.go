package state

import (
	"sync/atomic"
	"time"

	pbtypes "github.com/cometbft/cometbft/proto/tendermint/types"
)

var ProposerPubKey []byte

type ProposerStatus struct {
	IsProposer bool
	Height     int64
}

var Proposer atomic.Value

func init() {
	Proposer.Store(ProposerStatus{
		IsProposer: false,
		Height:     0,
	})
}

type DBEntry struct {
	Key   string `json:"key"`
	Value Entry  `json:"value"`
}

type Entry struct {
	RequestedHeight int64       `json:"requested_height"`
	ProposerPubKey  []byte      `json:"pubkey"`
	ChainID         string      `json:"chain_id"`
	BlockHash       []byte      `json:"block_hash"`
	SignedState     SignedState `json:"signed_state"`
}

type SignedState struct {
	ValidatorAddress []byte                `json:"validator_address"`
	SignedHeight     int64                 `json:"signed_height"`
	SignedRound      int32                 `json:"signed_round"`
	SignedStep       pbtypes.SignedMsgType `json:"signed_step"`     // 0: unknown, 1: prevote, 2: precommit, 32: proposal
	SignedStepStr    string                `json:"signed_step_str"` // human-readable type
	VoteSignature    []byte                `json:"signature"`
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
