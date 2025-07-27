package sigclient

import (
	"cometkms/state"
	"context"
	"encoding/json"
	"fmt"
	"log"

	cmthttp "github.com/cometbft/cometbft/rpc/client/http"
)

func (s *SimpleSigner) SaveState(vote *state.SigningState, chainID string) error {
	// 1. Set up client to local CometBFT node
	client, err := cmthttp.New(s.RPCaddr, "/websocket")
	if err != nil {
		log.Printf("failed to connect to RPC: %x", err)
	}

	// Key format: "height:round:type(step)"
	key := fmt.Sprintf("%d:%d:%v", vote.Height, vote.Round, vote.Type)
	tx := state.DBEntry{
		Key: key,
		Value: state.Entry{
			RequestedHeight: vote.Height,
			ProposerPubKey:  state.ProposerPubKey,
			ChainID:         chainID,
			BlockHash:       vote.BlockID.BlockHash,
			SignedState: state.SignedState{
				ValidatorAddress: vote.ValidatorAddress,
				SignedHeight:     vote.Height,
				SignedRound:      vote.Round,
				SignedStep:       vote.Type,
				SignedStepStr:    vote.TypeStr,
				VoteSignature:    vote.Signature,
			},
		},
	}

	// 3. Marshal to JSON
	txBytes, err := json.Marshal(tx)
	if err != nil {
		log.Printf("failed to marshal tx: %x", err)
	}

	// 4. Broadcast the transaction
	log.Printf("Prepared transaction: signHeight=%d, Round=%d, Type=%d, Key=%s",
		vote.Height, vote.Round, vote.Type, key)

	// Broadcast the transaction
	res, err := client.BroadcastTxCommit(context.Background(), txBytes)
	if err != nil {
		log.Printf("broadcast failed: %x", err)
		return fmt.Errorf("broadcast failed: %w", err)
	}
	log.Printf("Broadcasted TX, hash: %x, cometKMS-height: %d, code: %d, signRequest-height: %s",
		res.Hash, res.Height, res.CheckTx.Code, vote.Height)
	return nil
}

// readState reads the signer state from the file.
func (s *SimpleSigner) ReadState() (*state.SigningState, error) {
	// Read the signer state from the file
	var state *state.SigningState
	return state, nil
}
