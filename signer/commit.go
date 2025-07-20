package signer

import (
	"cometkms/state"
	"context"
	"fmt"
	"log"
	"strconv"

	abcitypes "github.com/cometbft/cometbft/abci/types"
)

func (app *App) Commit(_ context.Context, commit *abcitypes.RequestCommit) (*abcitypes.ResponseCommit, error) {
	app.AppHeight++

	state.Proposer.Store(state.ProposerStatus{
		IsProposer: false, // Reset proposer status
		Height:     app.AppHeight,
	})

	txn := app.db.NewTransaction(true)
	if err := txn.Set([]byte("meta:app_height"), []byte(strconv.FormatInt(app.AppHeight, 10))); err != nil {
		log.Fatalf("failed to store height: %v", err)
	}
	if err := txn.Set([]byte("meta:app_hash"), app.AppHash); err != nil {
		log.Fatalf("failed to store hash: %v", err)
	}
	if err := txn.Commit(); err != nil {
		log.Fatalf("failed to commit height: %v", err)
	}
	// Commit staged txs from FinalizeBlock
	log.Printf("Commit: AppHeight=%d, computed AppHash=%X", app.AppHeight, app.AppHash)
	if err := app.onGoingBlock.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}
	// Generate app hash — e.g., hash of height or root key
	return &abcitypes.ResponseCommit{}, nil
}
