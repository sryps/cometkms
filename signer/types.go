package signer

import (
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/dgraph-io/badger/v4"
)

const (
	// Invalid Format
	ERRORCODEInvalidFormat = 1
	// TX Already Exists
	ERRORCODEAlreadyExists = 2
	// Invalid DBEntry
	ERRORCODEInvalidDBEntry = 3
)

var _ abcitypes.Application = (*App)(nil)

type App struct {
	db           *badger.DB
	onGoingBlock *badger.Txn
	AppHeight    int64
	AppHash      []byte
	BlockTxs     [][]byte
}
