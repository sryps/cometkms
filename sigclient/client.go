package sigclient

import (
	"context"
	"errors"
	"fmt"
	cmtp2pconn "github.com/cometbft/cometbft/p2p/conn"
	pbprivval "github.com/cometbft/cometbft/proto/tendermint/privval"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

func (s *SimpleSigner) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Context done, stopping remote signer")
			return
		default:
			if err := s.ConnectAndServe(ctx); err != nil {
				fmt.Println("connection error:", err)
				time.Sleep(2 * time.Second)
			}
		}

		log.Println("Remote signer stopped")
		<-ctx.Done()
	}
}

func (s *SimpleSigner) ConnectAndServe(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Create a new connection to the node
			proto, addr := ProtocolAndAddress(s.addr)
			connRaw, err := (&net.Dialer{}).DialContext(ctx, proto, addr)
			if err != nil {
				log.Printf("dial failed: %v", err)
				time.Sleep(2 * time.Second) // Wait before retrying
				continue
			}
			defer connRaw.Close()

			conn, err := cmtp2pconn.MakeSecretConnection(connRaw, s.privKey)
			if err != nil {
				log.Printf("secret connection failed: %v", err)
			}
			defer conn.Close()

			log.Println("Connected to node:", conn.RemoteAddr())

			// Set up a reader and writer for the connection
			for {
				select {
				case <-ctx.Done():
					log.Println("Context done, closing connection")
					return nil
				default:
					msg := pbprivval.Message{}
					msg, err = readMsg(conn, 1024*1024)
					if err != nil {
						if errors.Is(err, io.EOF) {
							log.Printf("connection closed by peer (EOF), reconnecting...")
							return nil
						} else {
							log.Printf("read failed: %v", err)
						}
					}

					if msg.Sum != nil {
						resp := s.handleRequest(&msg)
						_, err := writeMessage(conn, &resp)
						if err != nil {
							log.Printf("write failed: %v", err)
						}
						continue // continue to read next message
					}
				}
			}
		}
	}
}

func ProtocolAndAddress(addr string) (string, string) {
	if strings.Contains(addr, "://") {
		parts := strings.SplitN(addr, "://", 2)
		return parts[0], parts[1]
	}
	return "tcp", addr
}
