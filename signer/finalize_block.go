package signer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"log"
)

func (app *App) FinalizeBlock(_ context.Context, req *abcitypes.FinalizeBlockRequest) (*abcitypes.FinalizeBlockResponse, error) {
	var txs = make([]*abcitypes.ExecTxResult, len(req.Txs))

	app.onGoingBlock = app.db.NewTransaction(true)
	for i, tx := range req.Txs {
		if code, logData := app.validateTx(tx); code != 0 {
			log.Printf("Error: invalid transaction index %v", i)
			txs[i] = &abcitypes.ExecTxResult{Code: code, Log: logData}
		} else {

			unmarshaledTx, err := UnmarshalDBEntry(tx)
			if err != nil {
				log.Printf("Error unmarshaling transaction: %v", err)
				txs[i] = &abcitypes.ExecTxResult{Code: 4, Log: "Failed to unmarshal transaction"}
				continue
			}
			key := unmarshaledTx.Key
			value := unmarshaledTx.Value
			keyBytes := []byte(fmt.Sprintf("%x", key))
			valueBytes, err := json.Marshal(value)
			if err != nil {
				log.Fatalf("failed to marshal value: %v", err)
			}
			valueBytes = bytes.TrimSpace(valueBytes)

			if err := app.onGoingBlock.Set(keyBytes, valueBytes); err != nil {
				log.Panicf("Error writing to database, unable to execute tx: %v", err)
			}

			// Add an event for the transaction execution.
			// Multiple events can be emitted for a transaction, but we are adding only one event
			txs[i] = &abcitypes.ExecTxResult{
				Code: 0,
				Events: []abcitypes.Event{
					{
						Type: "app",
						Attributes: []abcitypes.EventAttribute{
							{Key: "key", Value: string(keyBytes), Index: true},
							{Key: "value", Value: string(valueBytes), Index: true},
						},
					},
				},
			}
		}
	}
	app.AppHash = req.GetHash()
	log.Printf("FinalizeBlock: computed AppHash = %X", app.AppHash)

	return &abcitypes.FinalizeBlockResponse{
		TxResults: txs,
		AppHash:   app.AppHash,
	}, nil
}
