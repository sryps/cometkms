package signer

import (
	"io"
	"log"
	"net"
	"sync"

	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtencoding "github.com/cometbft/cometbft/crypto/encoding"
	"github.com/cometbft/cometbft/libs/protoio"
	pbcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	pbprivval "github.com/cometbft/cometbft/proto/tendermint/privval"
	"github.com/golang/protobuf/proto"
)

// NewSimpleSigner initializes a signer with address and key.
func NewSigner(addr string, secondaryAddr string, privKey cmted25519.PrivKey, nonsigningPrivkey cmted25519.PrivKey, nonsigningPubkey pbcrypto.PublicKey, keyFilePath string, stateFilePath string) (*SimpleSigner, error) {
	pubKey, err := cmtencoding.PubKeyToProto(privKey.PubKey())
	if err != nil {
		return nil, err
	}

	return &SimpleSigner{
		connectionManager: &ConnectionManager{
			primaryAddr:   addr,
			secondaryAddr: secondaryAddr,
		},
		signingKey:                 privKey,
		signingPubkey:              pubKey,
		secondaryNonsigningPrivkey: nonsigningPrivkey,
		secondaryNonsigningPubkey:  nonsigningPubkey,
		signingKeyOwner:            Primary, // default to primary
		mu:                         sync.Mutex{},
		keyFilePath:                keyFilePath,
		stateFilePath:              stateFilePath,
	}, nil
}

func (s *SimpleSigner) handleRequest(msg *pbprivval.Message, pubkey pbcrypto.PublicKey, role Role) pbprivval.Message {
	// Main handler for incoming messages from node
	switch req := msg.Sum.(type) {

	// Handle Pubkey Requests
	case *pbprivval.Message_PubKeyRequest:
		if pubkey == s.secondaryNonsigningPubkey {
			log.Println("Received PubKeyRequest for secondary nonsigning key, not signing block")
		}
		return pbprivval.Message{
			Sum: &pbprivval.Message_PubKeyResponse{
				PubKeyResponse: &pbprivval.PubKeyResponse{PubKey: pubkey},
			},
		}

	// Handle Vote Signing Requests
	case *pbprivval.Message_SignVoteRequest:
		// Check for double sign attempts before handling the sign vote request
		dsCheck := s.isDoubleSignAttempt(req.SignVoteRequest)
		if !dsCheck {
			log.Printf("Signing block for role: %s", role)
			return s.handleSignVoteRequest(req.SignVoteRequest)
		} else {
			return pbprivval.Message{
				Sum: &pbprivval.Message_SignedVoteResponse{
					SignedVoteResponse: &pbprivval.SignedVoteResponse{
						Error: &pbprivval.RemoteSignerError{
							Description: "Double sign attempt detected",
						},
					},
				},
			}

		}

	// Handle Proposal Signing Requests
	case *pbprivval.Message_SignProposalRequest:
		return s.handleSignProposalRequest(req.SignProposalRequest)

	case *pbprivval.Message_PingRequest:
		return pbprivval.Message{
			Sum: &pbprivval.Message_PingResponse{PingResponse: &pbprivval.PingResponse{}},
		}

	default:
		return pbprivval.Message{}
	}
}

func readMsg(reader io.Reader, maxReadSize int) (msg pbprivval.Message, err error) {
	// Read a protobuf message from the reader with a maximum size limit
	if maxReadSize <= 0 {
		maxReadSize = 1024 * 1024 // 1MB
	}
	protoReader := protoio.NewDelimitedReader(reader, maxReadSize)
	_, err = protoReader.ReadMsg(&msg)
	return msg, err
}

func writeMessage(conn net.Conn, msg proto.Message) (int, error) {
	// Write a protobuf message to the connection
	writer := protoio.NewDelimitedWriter(conn)
	return writer.WriteMsg(msg)
}
