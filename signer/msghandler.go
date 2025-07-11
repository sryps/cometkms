package signer

import (
	pbprivval "github.com/cometbft/cometbft/proto/tendermint/privval"
	"github.com/cometbft/cometbft/types"
	"log"
)

func (s *SimpleSigner) handleSignVoteRequest(req *pbprivval.SignVoteRequest) pbprivval.Message {

	// Sign the vote request body
	var err error
	signBytes := types.VoteSignBytes(req.ChainId, req.Vote)
	req.Vote.Signature, err = s.signingKey.Sign(signBytes)
	if err != nil {
		return s.returnSigningVoteError(err)
	}

	// Sign the vote extension (if present)
	if len(req.Vote.Extension) > 0 {
		extSignBytes := types.VoteExtensionSignBytes(req.ChainId, req.Vote)
		req.Vote.ExtensionSignature, err = s.signingKey.Sign(extSignBytes)
		if err != nil {
			return s.returnSigningVoteError(err)
		}
	}

	// Assign state struct with requested vote information
	state := &SigningState{
		Type:    req.Vote.Type,
		TypeStr: req.Vote.Type.String(),
		Height:  req.Vote.Height,
		Round:   req.Vote.Round,
		BlockID: BlockID{
			BlockHash: req.Vote.BlockID.Hash,
			PartSetHeader: PartSetHeader{
				Hash:  req.Vote.BlockID.PartSetHeader.Hash,
				Total: req.Vote.BlockID.PartSetHeader.Total,
			},
		},
		ValidatorAddress:   req.Vote.ValidatorAddress,
		Timestamp:          req.Vote.Timestamp,
		Signature:          req.Vote.Signature,
		ExtensionSignature: req.Vote.ExtensionSignature,
		ChainId:            req.ChainId,
	}

	// Write the vote to the state file
	if err := s.saveState(state); err != nil {
		log.Fatalf("Failed to save signer state: %v", err)
	}

	// Log the signed vote
	log.Printf(
		"Signed vote:  height=%d  round=%d  type=%s  hash=%X\n",
		req.Vote.Height,
		req.Vote.Round,
		req.Vote.Type,
		req.Vote.BlockID.Hash,
	)

	// Return the signed vote response
	return pbprivval.Message{
		Sum: &pbprivval.Message_SignedVoteResponse{
			SignedVoteResponse: &pbprivval.SignedVoteResponse{
				Vote: *req.Vote,
			},
		},
	}
}

func (s *SimpleSigner) handleSignProposalRequest(req *pbprivval.SignProposalRequest) pbprivval.Message {

	// Sign the proposal
	var err error
	signBytes := types.ProposalSignBytes(req.ChainId, req.Proposal)
	req.Proposal.Signature, err = s.signingKey.Sign(signBytes)
	if err != nil {
		return s.returnSigningProposalError(err)
	}

	// Return the signed proposal response
	return pbprivval.Message{
		Sum: &pbprivval.Message_SignedProposalResponse{
			SignedProposalResponse: &pbprivval.SignedProposalResponse{
				Proposal: *req.Proposal,
			},
		},
	}
}
