package sigclient

import (
	pbprivval "github.com/cometbft/cometbft/api/cometbft/privval/v1"
)

func (s *SimpleSigner) returnSigningProposalError(err error) pbprivval.Message {
	// Return a signed proposal response with an error
	return pbprivval.Message{
		Sum: &pbprivval.Message_SignedProposalResponse{
			SignedProposalResponse: &pbprivval.SignedProposalResponse{
				Error: &pbprivval.RemoteSignerError{Description: err.Error()},
			},
		},
	}
}

func (s *SimpleSigner) returnSigningVoteError(err error) pbprivval.Message {
	// Return a signed vote response with an error
	return pbprivval.Message{
		Sum: &pbprivval.Message_SignedVoteResponse{
			SignedVoteResponse: &pbprivval.SignedVoteResponse{
				Error: &pbprivval.RemoteSignerError{Description: err.Error()},
			},
		},
	}
}
