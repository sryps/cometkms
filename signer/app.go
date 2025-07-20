package signer

import (
	//"cometkms/sighandler"
	"cometkms/state"
	"context"
	"errors"
	"log"
	"strconv"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/dgraph-io/badger/v4"
)

func NewSigner(db *badger.DB) *App {
	app := &App{
		db:        db,
		AppHeight: 0,
		AppHash:   []byte{},
		BlockTxs:  [][]byte{},
	}

	var err error
	app.AppHeight, app.AppHash, err = app.LoadMetadata()
	if err != nil {
		log.Printf("Warning: failed to load metadata: %v", err)
	}
	return app
}

func (app *App) Info(_ context.Context, info *abcitypes.RequestInfo) (*abcitypes.ResponseInfo, error) {
	log.Printf("App Info: height=%d, hash=%x", app.AppHeight, app.AppHash)
	return &abcitypes.ResponseInfo{
		LastBlockHeight:  app.AppHeight,
		LastBlockAppHash: app.AppHash,
		Data:             "CometKMS Signer Application",
	}, nil
}

func (app *App) Query(_ context.Context, req *abcitypes.RequestQuery) (*abcitypes.ResponseQuery, error) {
	resp := abcitypes.ResponseQuery{Key: req.Data}

	dbErr := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(req.Data)
		if err != nil {
			if !errors.Is(err, badger.ErrKeyNotFound) {
				return err
			}
			resp.Log = "key does not exist"
			return nil
		}

		return item.Value(func(val []byte) error {
			resp.Log = "exists"
			resp.Value = val
			return nil
		})
	})
	if dbErr != nil {
		log.Panicf("Error reading database, unable to execute query: %v", dbErr)
	}
	return &resp, nil
}

func (app *App) InitChain(_ context.Context, chain *abcitypes.RequestInitChain) (*abcitypes.ResponseInitChain, error) {
	return &abcitypes.ResponseInitChain{}, nil
}

func (app *App) PrepareProposal(_ context.Context, proposal *abcitypes.RequestPrepareProposal) (*abcitypes.ResponsePrepareProposal, error) {
	log.Printf("I am the proposer for height %d with public key %s", proposal.Height, state.ProposerPubKey)
	state.Proposer.Store(state.ProposerStatus{
		IsProposer: true,
		Height:     proposal.Height,
	})
	//go sighandler.TestTx(proposal.Height)
	return &abcitypes.ResponsePrepareProposal{Txs: proposal.Txs}, nil
}

func (app *App) ProcessProposal(_ context.Context, proposal *abcitypes.RequestProcessProposal) (*abcitypes.ResponseProcessProposal, error) {
	return &abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_ACCEPT}, nil
}

func (app *App) ListSnapshots(_ context.Context, snapshots *abcitypes.RequestListSnapshots) (*abcitypes.ResponseListSnapshots, error) {
	return &abcitypes.ResponseListSnapshots{}, nil
}

func (app *App) OfferSnapshot(_ context.Context, snapshot *abcitypes.RequestOfferSnapshot) (*abcitypes.ResponseOfferSnapshot, error) {
	return &abcitypes.ResponseOfferSnapshot{}, nil
}

func (app *App) LoadSnapshotChunk(_ context.Context, chunk *abcitypes.RequestLoadSnapshotChunk) (*abcitypes.ResponseLoadSnapshotChunk, error) {
	return &abcitypes.ResponseLoadSnapshotChunk{}, nil
}

func (app *App) ApplySnapshotChunk(_ context.Context, chunk *abcitypes.RequestApplySnapshotChunk) (*abcitypes.ResponseApplySnapshotChunk, error) {
	return &abcitypes.ResponseApplySnapshotChunk{}, nil
}

func (app *App) VerifyVoteExtension(_ context.Context, verify *abcitypes.RequestVerifyVoteExtension) (*abcitypes.ResponseVerifyVoteExtension, error) {
	return &abcitypes.ResponseVerifyVoteExtension{}, nil
}

func (app *App) ExtendVote(_ context.Context, extend *abcitypes.RequestExtendVote) (*abcitypes.ResponseExtendVote, error) {
	return &abcitypes.ResponseExtendVote{}, nil
}

// LoadMetadata loads the application metadata from the database.
// It retrieves the application appHeight and appHash from the database.
func (app *App) LoadMetadata() (int64, []byte, error) {
	var height int64
	var hash []byte
	err := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("meta:app_height"))
		if err == nil {
			val, _ := item.ValueCopy(nil)
			height, _ = strconv.ParseInt(string(val), 10, 64)
		}

		item, err = txn.Get([]byte("meta:app_hash"))
		if err == nil {
			hash, _ = item.ValueCopy(nil)
		}

		return nil
	})
	return height, hash, err
}
