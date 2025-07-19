package signer

import (
	"bytes"
	"cometkms/types"
	"context"
	"encoding/json"
	"fmt"
	"log"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/dgraph-io/badger/v4"
)

func (app *App) validateTx(tx []byte) (uint32, string) {
	var code uint32
	// Check if the transaction is empty
	if len(tx) == 0 {
		log.Println("Transaction is empty")
		code = 1
	}
	// Check if the transaction already exists
	if app.txExists(tx) {
		log.Printf("Transaction already exists: %s", tx)
		code = 2
	}

	// Check if the transaction has a valid DBEntry struct format
	_, err := UnmarshalDBEntry(tx)
	if err != nil {
		code = 3
	}

	log := app.TxErrorLog(tx, code)

	return code, log
}

func (app *App) CheckTx(_ context.Context, check *abcitypes.CheckTxRequest) (*abcitypes.CheckTxResponse, error) {
	var log string
	code, log := app.validateTx(check.Tx)
	return &abcitypes.CheckTxResponse{Code: code, Log: log}, nil
}

// Check if Key-Value pair already exists
func (app *App) txExists(tx []byte) bool {
	key := bytes.Split(tx, []byte("="))[0]
	exists := false
	_ = app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == nil {
			log.Printf("Key %s already exists", key)
			exists = item != nil
		}
		return nil
	})
	return exists
}

func UnmarshalDBEntry(data []byte) (*types.DBEntry, error) {
	var entry types.DBEntry
	err := json.Unmarshal(data, &entry)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (app *App) TxErrorLog(tx []byte, code uint32) string {
	var log string
	switch code {
	case 0:
		log = "Transaction is valid"
	case ERRORCODEInvalidFormat:
		log = fmt.Sprintf("Invalid format for transaction: %s", tx)
	case ERRORCODEAlreadyExists:
		log = fmt.Sprintf("Transaction already exists: %s", tx)
	case ERRORCODEInvalidDBEntry:
		log = fmt.Sprintf("Invalid DBEntry format for transaction: %s", tx)
	default:
		log = fmt.Sprintf("Unknown error for transaction: %s, code: %d", tx, code)
	}
	return log
}
