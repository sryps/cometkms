package sigclient

import (
	"io"
	"net"

	pbprivval "github.com/cometbft/cometbft/api/cometbft/privval/v1"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtencoding "github.com/cometbft/cometbft/crypto/encoding"
	"github.com/cometbft/cometbft/libs/protoio"
	"github.com/golang/protobuf/proto"
)

// NewSimpleSigner initializes a signer with address and key.
func SigClient(addr string, privKey cmted25519.PrivKey, keyFilePath string) (*SimpleSigner, error) {
	pubKey, err := cmtencoding.PubKeyToProto(privKey.PubKey())
	if err != nil {
		return nil, err
	}

	pubKeyStr, err := PublicKeyToString(pubKey)
	return &SimpleSigner{
		addr:    addr,
		privKey: privKey,
		PubKey: PubKey{
			PubKeyType:    pubKey,
			PubKeyBytes:   pubKey.GetEd25519(),
			PubKeyTypeStr: pubKeyStr,
		},
		keyFilePath: keyFilePath,
	}, nil
}

func (s *SimpleSigner) handleRequest(msg *pbprivval.Message) pbprivval.Message {
	// Main handler for incoming messages from node
	switch req := msg.Sum.(type) {

	// Handle Pubkey Requests
	case *pbprivval.Message_PubKeyRequest:
		return pbprivval.Message{
			Sum: &pbprivval.Message_PubKeyResponse{
				PubKeyResponse: &pbprivval.PubKeyResponse{
					PubKeyType:  s.PubKey.PubKeyTypeStr,
					PubKeyBytes: s.PubKey.PubKeyBytes,
				},
			},
		}

	// Handle Vote Signing Requests
	case *pbprivval.Message_SignVoteRequest:
		// Check for double sign attempts before handling the sign vote request
		dsCheck := s.isDoubleSignAttempt(req.SignVoteRequest)
		if !dsCheck {
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
