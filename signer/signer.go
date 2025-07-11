package signer

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtencoding "github.com/cometbft/cometbft/crypto/encoding"
	cmtnet "github.com/cometbft/cometbft/libs/net"
	"github.com/cometbft/cometbft/libs/protoio"
	cmtp2pconn "github.com/cometbft/cometbft/p2p/conn"
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

func (s *SimpleSigner) connectAndServe(ctx context.Context, addr string, privateKey cmted25519.PrivKey, pubkey pbcrypto.PublicKey, role Role) error {
	for {
		select {
		case <-ctx.Done():
			log.Printf("Context done, stopping key runner for role %s", role)
			return nil
		default:
			// Create a new connection to the node
			proto, addr := cmtnet.ProtocolAndAddress(addr)
			connRaw, err := (&net.Dialer{}).DialContext(ctx, proto, addr)
			if err != nil {
				return fmt.Errorf("dial failed: %w", err)
			}
			defer connRaw.Close()

			conn, err := cmtp2pconn.MakeSecretConnection(connRaw, privateKey)
			if err != nil {
				return fmt.Errorf("secret connection failed: %w", err)
			}

			log.Println("Connected to node:", conn.RemoteAddr(), "as", role)

			if addr == s.connectionManager.primaryAddr {
				s.connectionManager.primaryConn = conn
			}
			if addr == s.connectionManager.secondaryAddr {
				s.connectionManager.secondaryConn = conn
			}

			select {
			case <-ctx.Done():
				return nil
			case <-time.After(2 * time.Second):
			}

			// Set up a reader and writer for the connection
			for {
				msg := pbprivval.Message{}
				msg, err = readMsg(conn, 1024*1024)
				if err != nil {
					return fmt.Errorf("read failed: %w", err)
				}

				resp := s.handleRequest(&msg, pubkey)
				_, err := writeMessage(conn, &resp)
				if err != nil {
					return fmt.Errorf("write failed: %w", err)
				}
			}
		}
	}
}

func (s *SimpleSigner) handleRequest(msg *pbprivval.Message, pubkey pbcrypto.PublicKey) pbprivval.Message {
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
			return s.handleSignVoteRequest(req.SignVoteRequest)
		}

	// Handle Proposal Signing Requests
	case *pbprivval.Message_SignProposalRequest:
		return s.handleSignProposalRequest(req.SignProposalRequest)

	default:
		return pbprivval.Message{
			Sum: &pbprivval.Message_PingResponse{PingResponse: &pbprivval.PingResponse{}},
		}
	}
	return pbprivval.Message{}
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
