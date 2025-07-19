package sighandler

import (
	"cometkms/types"
	"context"
	"encoding/json"
	"log"

	cmthttp "github.com/cometbft/cometbft/rpc/client/http"
)

func TestTx(height int64) {
	// 1. Set up client to local CometBFT node
	client, err := cmthttp.New("http://localhost:26657")
	if err != nil {
		log.Printf("failed to connect to RPC: %x", err)
	}

	tx := types.DBEntry{
		Key: height,
		Value: types.Entry{
			RequestedHeight: height,
			PubKey:          []byte("pubkey"),
			ChainID:         "test-chain",
			BlockHash:       []byte("blockhash"),
			SignedState: types.SignedState{
				SignedHeight:  100,
				SignedRound:   0,
				SignedStep:    "prevote",
				VoteSignature: []byte("signature"),
			},
		},
	}

	// 3. Marshal to JSON
	txBytes, err := json.Marshal(tx)
	if err != nil {
		log.Printf("failed to marshal tx: %x", err)
	}

	// 4. Broadcast the transaction
	res, err := client.BroadcastTxCommit(context.Background(), txBytes)
	if err != nil {
		log.Printf("broadcast failed: %x", err)
		return
	}
	log.Printf("Broadcasted TX, hash: %x, height: %d, code: %d, log: %s",
		res.Hash, res.Height, res.CheckTx.Code, res.CheckTx.Log)
}
