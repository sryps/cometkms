package sigclient

import (
	"cometkms/types"
	"context"
	"encoding/json"
	"fmt"
	cmthttp "github.com/cometbft/cometbft/rpc/client/http"
	"log"
)

func (s *SimpleSigner) SaveState(vote *types.SigningState) error {
	// 1. Set up client to local CometBFT node
	client, err := cmthttp.New("http://localhost:2657", "/websocket")
	if err != nil {
		log.Printf("failed to connect to RPC: %x", err)
	}

	// Key format: "height:round:type(step)"
	key := fmt.Sprintf("%d:%d:%v", vote.Height, vote.Round, vote.Type)
	tx := types.DBEntry{
		Key: key,
		Value: types.Entry{
			RequestedHeight: vote.Height,
			PubKey:          []byte("TODO"),
			LastBlockSigner: "TODO",
			ChainID:         "TODO",
			BlockHash:       vote.BlockID.BlockHash,
			SignedState: types.SignedState{
				SignedHeight:  vote.Height,
				SignedRound:   vote.Round,
				SignedStep:    vote.Type,
				SignedStepStr: vote.TypeStr,
				VoteSignature: vote.Signature,
			},
		},
	}

	// 3. Marshal to JSON
	txBytes, err := json.Marshal(tx)
	if err != nil {
		log.Printf("failed to marshal tx: %x", err)
	}

	log.Printf("Prepared transaction: Height=%d, Round=%d, Type=%d, Key=%s",
		vote.Height, vote.Round, vote.Type, key)
	// 4. Broadcast the transaction
	res, err := client.BroadcastTxCommit(context.Background(), txBytes)
	if err != nil {
		log.Printf("broadcast failed: %x", err)
		return fmt.Errorf("broadcast failed: %w", err)
	}
	log.Printf("Broadcasted TX, hash: %x, height: %d, code: %d, log: %s",
		res.Hash, res.Height, res.CheckTx.Code, res.CheckTx.Log)
	return nil
}

// readState reads the signer state from the file.
func (s *SimpleSigner) ReadState() (*types.SigningState, error) {
	// Read the signer state from the file
	var state *types.SigningState
	return state, nil
}
