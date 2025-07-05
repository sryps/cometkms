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

func (s *SimpleSigner) clientConnections(ctx context.Context) error {

	var err error

	s.connectionManager.isPrimaryConnActiveSigner = true
	s.connectionManager.isSecondaryConnActiveSigner = false

	// Setup primary connection
	s.connectionManager.primaryConn, err = s.initConnection(ctx, s.address)
	if err != nil {
		s.connectionManager.isPrimaryConnActiveSigner = false
		s.connectionManager.isSecondaryConnActiveSigner = true
		logmsg := fmt.Errorf("failed to connect to primary node: %w", err)
		log.Printf("%s", logmsg)

		// while loop to retry connection until s.connectionManager.primaryConn is established
		for s.connectionManager.primaryConn == nil {
			s.connectionManager.primaryConn, err = s.initConnection(ctx, s.address)
			if err != nil {
				log.Printf("Retry failed: %s", err)
				time.Sleep(2 * time.Second) // Wait before retrying
			}
		}
		s.connectionManager.isPrimaryConnActiveSigner = true
		s.connectionManager.isSecondaryConnActiveSigner = false
	}
	defer s.connectionManager.primaryConn.Close()
	log.Println("Connected to primary node:", s.address)

	// Setup secondary connection
	if s.addressBackup == "" {
		log.Println("No secondary address provided, using primary connection only.")
	} else {
		// Attempt to connect to the secondary node
		s.connectionManager.secondaryConn, err = s.initConnection(ctx, s.addressBackup)
		if err != nil {
			s.connectionManager.isPrimaryConnActiveSigner = true
			s.connectionManager.isSecondaryConnActiveSigner = false
			logmsg := fmt.Errorf("failed to connect to secondary node: %w", err)
			log.Printf("%s", logmsg)

		}
		defer s.connectionManager.secondaryConn.Close()
		log.Println("Connected to secondary node:", s.addressBackup)
	}

	// Set up a reader and writer for both connections
	for {
		// Assign separate vars for primary and secondary connections before entering the loop so they cant change in the middle of the loop.
		// There is scenario where primary = true and then connection fails, sets secondary to true
		// and then it submits signature to secondary connection after already submitting to primary connection
		isPrimaryConnActive := s.connectionManager.isPrimaryConnActiveSigner
		isSecondaryConnActive := s.connectionManager.isSecondaryConnActiveSigner

		// Check to make sure both connections aren't active signers at the same time.
		if isPrimaryConnActive && isSecondaryConnActive {
			return fmt.Errorf("both primary and secondary connections are active, cannot handle messages")
		} else {

			// Handle messages from the primary connection
			if isPrimaryConnActive && s.connectionManager.primaryConn != nil {
				err := s.haMsgHandler(ctx, s.connectionManager.primaryConn, isPrimaryConnActive)
				if err != nil {
					s.connectionManager.isPrimaryConnActiveSigner = false
					s.connectionManager.isSecondaryConnActiveSigner = true
					return fmt.Errorf("failed to handle primary connection message, switching to secondary connection: %w", err)
				}
			}

			// Handle messages from the secondary connection
			if isSecondaryConnActive && s.connectionManager.secondaryConn != nil {
				err = s.haMsgHandler(ctx, s.connectionManager.secondaryConn, isSecondaryConnActive)
				if err != nil {
					s.connectionManager.isPrimaryConnActiveSigner = true
					s.connectionManager.isSecondaryConnActiveSigner = false
					return fmt.Errorf("failed to handle secondary connection message, switching to primary connection: %w", err)
				}
			}
		}
	}
}

// HA (High Availability) connection handler
func (s *SimpleSigner) initConnection(ctx context.Context, address string) (*cmtp2pconn.SecretConnection, error) {
	// Create a new connection to the node
	proto, addr := cmtnet.ProtocolAndAddress(address)
	connRaw, err := (&net.Dialer{}).DialContext(ctx, proto, addr)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %w", err)
	}

	conn, err := cmtp2pconn.MakeSecretConnection(connRaw, s.privKey)
	if err != nil {
		return nil, fmt.Errorf("secret connection failed: %w", err)
	}

	return conn, nil
}

func (s *SimpleSigner) haMsgHandler(ctx context.Context, conn *cmtp2pconn.SecretConnection, activeConn bool) error {
	msg := pbprivval.Message{}
	msg, err := readMsg(conn, 1024*1024)
	if err != nil {
		return fmt.Errorf("readMsg failed: %w", err)
	}

	resp := s.handleRequest(&msg, activeConn)
	_, err = writeMessage(conn, &resp)
	if err != nil {
		return fmt.Errorf("writeMsg failed: %w", err)
	}
	return nil
}
