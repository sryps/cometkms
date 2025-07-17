package signer

const (
	// Invalid Format
	ERRORCODEInvalidFormat = 1
	// TX Already Exists
	ERRORCODEAlreadyExists = 2
	// Invalid DBEntry
	ERRORCODEInvalidDBEntry = 3
)

type DBEntry struct {
	Key   []byte `json:"key"`
	Value Entry
}

type Entry struct {
	RequestedHeight int64       `json:"requested_height"`
	PubKey          []byte      `json:"pubkey"`
	ChainID         string      `json:"chain_id"`
	BlockHash       []byte      `json:"block_hash"`
	SignedState     SignedState `json:"signed_state"`
}

type SignedState struct {
	SignedHeight  int64  `json:"signed_height"`
	SignedRound   int64  `json:"signed_round"`
	SignedStep    string `json:"signed_step"`
	VoteSignature []byte `json:"signature"`
}
