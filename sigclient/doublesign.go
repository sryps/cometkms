package sigclient

import (
	pbprivval "github.com/cometbft/cometbft/api/cometbft/privval/v1"
	"log"
)

func (s *SimpleSigner) isDoubleSignAttempt(req *pbprivval.SignVoteRequest) bool {

	// Load the last state from the file
	var lastState *SigningState
	lastState, err := s.ReadState()
	if err != nil {
		log.Fatalf("Failed to read signer state: %v", err)
	}

	// check if height,round and type of the last vote are greater than or equal to the current request
	if lastState.Height >= req.Vote.Height &&
		lastState.Round >= req.Vote.Round &&
		lastState.Type >= req.Vote.Type {
		log.Printf("DOUBLE SIGN ATTEMPT for vote at height %d, round %d, block ID %X\n",
			req.Vote.Height,
			req.Vote.Round,
			req.Vote.BlockID.Hash,
		)
		return true
	}
	return false
}
