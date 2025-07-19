package sigclient

import (
	"context"
	"fmt"
	pbprivval "github.com/cometbft/cometbft/api/cometbft/privval/v1"
	cmtp2pconn "github.com/cometbft/cometbft/p2p/conn"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

func (s *SimpleSigner) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
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
		}
	}()

	wg.Wait()
	log.Println("Remote signer stopped")
	<-ctx.Done()
	cancel()
	return nil
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
				return fmt.Errorf("dial failed: %w", err)
			}
			defer connRaw.Close()

			conn, err := cmtp2pconn.MakeSecretConnection(connRaw, s.privKey)
			if err != nil {
				return fmt.Errorf("secret connection failed: %w", err)
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
						return fmt.Errorf("read failed: %w", err)
					}

					resp := s.handleRequest(&msg)
					_, err := writeMessage(conn, &resp)
					if err != nil {
						return fmt.Errorf("write failed: %w", err)
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
