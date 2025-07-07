package signer

import (
	"log"
)

func (s *SimpleSigner) contextHandler() error {
	keyRole := s.getSigningKeyOwner()

	if s.connectionManager.primaryConn == nil && keyRole == Primary {
		log.Printf("Primary connection is nil, but signing key owner is primary. Attempting to reconnect.")
		if s.connectionManager.secondaryConn != nil {
			log.Printf("Secondary connection is available, switching signing key owner to secondary.")
			s.setSigningKeyOwner(Secondary)
			s.connectionManager.primaryConn.Close()
			s.connectionManager.secondaryConn.Close()
		}
	}

	if s.connectionManager.secondaryConn == nil && keyRole == Secondary {
		log.Printf("Secondary connection is nil, but signing key owner is secondary. Attempting to reconnect.")
		if s.connectionManager.primaryConn != nil {
			s.setSigningKeyOwner(Primary)
			s.connectionManager.secondaryConn.Close()
			s.connectionManager.primaryConn.Close()
		}
	}

	return nil
}
