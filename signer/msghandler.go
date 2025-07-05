package signer

import (
	"log"

	pbprivval "github.com/cometbft/cometbft/proto/tendermint/privval"
	"github.com/cometbft/cometbft/types"
)

func (s *SimpleSigner) handleRequest(msg *pbprivval.Message, activeConn bool) pbprivval.Message {
	// Main handler for incoming messages from node
	switch req := msg.Sum.(type) {

	// Handle Pubkey Requests
	case *pbprivval.Message_PubKeyRequest:
		log.Printf("Received PubKeyRequest for chain ID: %s", req.PubKeyRequest.ChainId)
		return pbprivval.Message{
			Sum: &pbprivval.Message_PubKeyResponse{
				PubKeyResponse: &pbprivval.PubKeyResponse{PubKey: s.PubKey},
			},
		}

	// Handle Vote Signing Requests
	case *pbprivval.Message_SignVoteRequest:
		log.Printf("Received SignVoteRequest for chain ID: %s", req.SignVoteRequest.ChainId)
		// Check for double sign attempts before handling the sign vote request
		if !activeConn {
			log.Printf("Ignoring sign vote request from inactive connection: %s", req.SignVoteRequest.ChainId)
			return pbprivval.Message{}
		} else {
			dsCheck := s.isDoubleSignAttempt(req.SignVoteRequest)
			if !dsCheck {
				return s.handleSignVoteRequest(req.SignVoteRequest)
			} else {
				return pbprivval.Message{}
			}
		}

	// Handle Proposal Signing Requests
	case *pbprivval.Message_SignProposalRequest:
		log.Printf("Received SignProposalRequest for chain ID: %s", req.SignProposalRequest.ChainId)
		if !activeConn {
			log.Printf("Ignoring sign proposal request from inactive connection: %s", req.SignProposalRequest.ChainId)
			return pbprivval.Message{}
		} else {
			return s.handleSignProposalRequest(req.SignProposalRequest)
		}

	default:
		log.Printf("Received Ping Request: %T", msg.Sum)
		return pbprivval.Message{
			Sum: &pbprivval.Message_PingResponse{PingResponse: &pbprivval.PingResponse{}},
		}
	}
}

func (s *SimpleSigner) handleSignVoteRequest(req *pbprivval.SignVoteRequest) pbprivval.Message {

	// Sign the vote request body
	var err error
	signBytes := types.VoteSignBytes(req.ChainId, req.Vote)
	req.Vote.Signature, err = s.privKey.Sign(signBytes)
	if err != nil {
		return s.returnSigningVoteError(err)
	}

	// Sign the vote extension (if present)
	if len(req.Vote.Extension) > 0 {
		extSignBytes := types.VoteExtensionSignBytes(req.ChainId, req.Vote)
		req.Vote.ExtensionSignature, err = s.privKey.Sign(extSignBytes)
		if err != nil {
			return s.returnSigningVoteError(err)
		}
	}

	// Assign state struct with requested vote information
	state := &SigningState{
		Height:  req.Vote.Height,
		Round:   req.Vote.Round,
		TypeStr: req.Vote.Type.String(),
		Type:    req.Vote.Type,
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
	req.Proposal.Signature, err = s.privKey.Sign(signBytes)
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
