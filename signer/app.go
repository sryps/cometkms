package signer

import (
	//"cometkms/sighandler"
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

func (app *App) Info(_ context.Context, info *abcitypes.InfoRequest) (*abcitypes.InfoResponse, error) {
	log.Printf("App Info: height=%d, hash=%x", app.AppHeight, app.AppHash)
	return &abcitypes.InfoResponse{
		LastBlockHeight:  app.AppHeight,
		LastBlockAppHash: app.AppHash,
		Data:             "CometKMS Signer Application",
	}, nil
}

func (app *App) Query(_ context.Context, req *abcitypes.QueryRequest) (*abcitypes.QueryResponse, error) {
	resp := abcitypes.QueryResponse{Key: req.Data}

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

func (app *App) InitChain(_ context.Context, chain *abcitypes.InitChainRequest) (*abcitypes.InitChainResponse, error) {
	return &abcitypes.InitChainResponse{}, nil
}

func (app *App) PrepareProposal(_ context.Context, proposal *abcitypes.PrepareProposalRequest) (*abcitypes.PrepareProposalResponse, error) {
	//go sighandler.TestTx(proposal.Height)
	return &abcitypes.PrepareProposalResponse{Txs: proposal.Txs}, nil
}

func (app *App) ProcessProposal(_ context.Context, proposal *abcitypes.ProcessProposalRequest) (*abcitypes.ProcessProposalResponse, error) {
	return &abcitypes.ProcessProposalResponse{Status: abcitypes.PROCESS_PROPOSAL_STATUS_ACCEPT}, nil
}

func (app *App) ListSnapshots(_ context.Context, snapshots *abcitypes.ListSnapshotsRequest) (*abcitypes.ListSnapshotsResponse, error) {
	return &abcitypes.ListSnapshotsResponse{}, nil
}

func (app *App) OfferSnapshot(_ context.Context, snapshot *abcitypes.OfferSnapshotRequest) (*abcitypes.OfferSnapshotResponse, error) {
	return &abcitypes.OfferSnapshotResponse{}, nil
}

func (app *App) LoadSnapshotChunk(_ context.Context, chunk *abcitypes.LoadSnapshotChunkRequest) (*abcitypes.LoadSnapshotChunkResponse, error) {
	return &abcitypes.LoadSnapshotChunkResponse{}, nil
}

func (app *App) ApplySnapshotChunk(_ context.Context, chunk *abcitypes.ApplySnapshotChunkRequest) (*abcitypes.ApplySnapshotChunkResponse, error) {
	return &abcitypes.ApplySnapshotChunkResponse{Result: abcitypes.APPLY_SNAPSHOT_CHUNK_RESULT_ACCEPT}, nil
}

func (app App) ExtendVote(_ context.Context, extend *abcitypes.ExtendVoteRequest) (*abcitypes.ExtendVoteResponse, error) {
	return &abcitypes.ExtendVoteResponse{}, nil
}

func (app *App) VerifyVoteExtension(_ context.Context, verify *abcitypes.VerifyVoteExtensionRequest) (*abcitypes.VerifyVoteExtensionResponse, error) {
	return &abcitypes.VerifyVoteExtensionResponse{}, nil
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
