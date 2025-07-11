package signer

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	cmtnet "github.com/cometbft/cometbft/libs/net"
	cmtp2pconn "github.com/cometbft/cometbft/p2p/conn"
	pbprivval "github.com/cometbft/cometbft/proto/tendermint/privval"
)

func (s *SimpleSigner) primaryConnection(ctx context.Context, addr string, role Role) error {
	for {
		select {
		case <-ctx.Done():
			log.Printf("Context done, stopping key runner for role %s", role)
			return nil
		default:
			// Create a new connection to the node
			privateKey, _ := s.getKeysForRole(role)

			var err error
			proto, addr := cmtnet.ProtocolAndAddress(addr)
			s.connectionManager.primaryTcpConn, err = (&net.Dialer{}).DialContext(ctx, proto, addr)
			if err != nil {
				log.Printf("dial failed: %v", err)
				time.Sleep(2 * time.Second) // Wait before retrying connection
				continue
			}

			s.connectionManager.primaryConn, err = cmtp2pconn.MakeSecretConnection(s.connectionManager.primaryTcpConn, privateKey)
			if err != nil {
				log.Printf("secret connection failed: %v", err)
				s.connectionManager.primaryTcpConn.Close()
				time.Sleep(2 * time.Second) // Wait before retrying connection
				continue
			}

			log.Println("Connected to node:", s.connectionManager.primaryConn.RemoteAddr(), "as", role)

			// Set up a reader and writer for the connection
			for {
				select {
				case <-ctx.Done():
					s.connectionManager.primaryTcpConn.Close()
					s.connectionManager.primaryConn.Close()
					log.Printf("Context done inside message loop for role %s", role)
					return nil
				default:
					msg := pbprivval.Message{}
					msg, err = readMsg(s.connectionManager.primaryConn, 1024*1024)
					if err != nil {
						return fmt.Errorf("read failed: %w", err)
					}

					resp := s.handleRequest(&msg, s.signingPubkey)
					_, err := writeMessage(s.connectionManager.primaryConn, &resp)
					if err != nil {
						return fmt.Errorf("write failed: %w", err)
					}
				}
			}
		}
	}
}
