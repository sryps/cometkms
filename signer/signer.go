package signer

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtencoding "github.com/cometbft/cometbft/crypto/encoding"
	cmtnet "github.com/cometbft/cometbft/libs/net"
	"github.com/cometbft/cometbft/libs/protoio"
	cmtp2pconn "github.com/cometbft/cometbft/p2p/conn"
	pbprivval "github.com/cometbft/cometbft/proto/tendermint/privval"
	"github.com/golang/protobuf/proto"
)

// NewSimpleSigner initializes a signer with address and key.
func NewSigner(addr string, privKey cmted25519.PrivKey, keyFilePath string, stateFilePath string) (*SimpleSigner, error) {
	pubKey, err := cmtencoding.PubKeyToProto(privKey.PubKey())
	if err != nil {
		return nil, err
	}

	return &SimpleSigner{
		addr:          addr,
		privKey:       privKey,
		PubKey:        pubKey,
		keyFilePath:   keyFilePath,
		stateFilePath: stateFilePath,
	}, nil
}

// Run starts the signer and handles one connection at a time.
func (s *SimpleSigner) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := s.connectAndServe(ctx); err != nil {
				fmt.Println("connection error:", err)
				time.Sleep(2 * time.Second)
			}
		}
	}
}

func (s *SimpleSigner) connectAndServe(ctx context.Context) error {
	// Create a new connection to the node
	proto, addr := cmtnet.ProtocolAndAddress(s.addr)
	connRaw, err := (&net.Dialer{}).DialContext(ctx, proto, addr)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer connRaw.Close()

	conn, err := cmtp2pconn.MakeSecretConnection(connRaw, s.privKey)
	if err != nil {
		return fmt.Errorf("secret connection failed: %w", err)
	}

	log.Println("Connected to node:", conn.RemoteAddr())

	// Set up a reader and writer for the connection
	for {
		msg := pbprivval.Message{}
		msg, err = readMsg(conn, 1024*1024)
		if err != nil {
			return fmt.Errorf("read failed: %w", err)
		}

		resp := s.handleRequest(&msg)
		_, err := writeMessage(conn, &resp)
		if err != nil {
			return fmt.Errorf("write failed: %w", err)
		}
	}
}

func (s *SimpleSigner) handleRequest(msg *pbprivval.Message) pbprivval.Message {
	// Main handler for incoming messages from node
	switch req := msg.Sum.(type) {

	// Handle Pubkey Requests
	case *pbprivval.Message_PubKeyRequest:
		return pbprivval.Message{
			Sum: &pbprivval.Message_PubKeyResponse{
				PubKeyResponse: &pbprivval.PubKeyResponse{PubKey: s.PubKey},
			},
		}

	// Handle Vote Signing Requests
	case *pbprivval.Message_SignVoteRequest:
		// Check for double sign attempts before handling the sign vote request
		dsCheck := s.isDoubleSignAttempt(req.SignVoteRequest)
		if !dsCheck {
			return s.handleSignVoteRequest(req.SignVoteRequest)
		} else {
			return pbprivval.Message{}
		}

	// Handle Proposal Signing Requests
	case *pbprivval.Message_SignProposalRequest:
		return s.handleSignProposalRequest(req.SignProposalRequest)

	default:
		return pbprivval.Message{
			Sum: &pbprivval.Message_PingResponse{PingResponse: &pbprivval.PingResponse{}},
		}
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
