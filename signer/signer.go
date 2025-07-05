package signer

import (
	"context"
	"io"
	"log"
	"net"
	"time"

	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtencoding "github.com/cometbft/cometbft/crypto/encoding"
	"github.com/cometbft/cometbft/libs/protoio"
	pbcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	pbprivval "github.com/cometbft/cometbft/proto/tendermint/privval"
	"github.com/golang/protobuf/proto"
)

// NewSimpleSigner initializes a signer with address and key.
func NewSigner(address string, addressBackup string, privKey cmted25519.PrivKey, keyFilePath string, stateFilePath string, secPubKey pbcrypto.PublicKey) (*SimpleSigner, error) {
	pubKey, err := cmtencoding.PubKeyToProto(privKey.PubKey())
	if err != nil {
		return nil, err
	}

	return &SimpleSigner{
		address:       address,
		addressBackup: addressBackup,
		privKey:       privKey,
		PubKey:        pubKey,
		keyFilePath:   keyFilePath,
		stateFilePath: stateFilePath,
		connectionManager: &ConnectionManager{
			isPrimaryConnActiveSigner:   true,
			isSecondaryConnActiveSigner: false,
		},
		secondaryPubkey: secPubKey,
	}, nil
}

// Run starts the signer and handles one connection at a time.
func (s *SimpleSigner) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := s.clientConnections(ctx); err != nil {
				log.Println("connection error:", err)
				time.Sleep(2 * time.Second)
			}
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
