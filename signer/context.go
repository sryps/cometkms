package signer

import (
	"log"
)

func (s *SimpleSigner) contextHandler() error {
	keyOwner := s.getSigningKeyOwner()

	if s.connectionManager.primaryConn == nil && keyOwner == Primary {
		log.Printf("Primary connection is nil, but signing key owner is primary. Attempting to reconnect.")
		if s.connectionManager.secondaryConn != nil {
			log.Printf("Secondary connection is available, switching signing key owner to secondary.")
			s.setSigningKeyOwner(Secondary)
			s.connectionManager.secondaryConn.Close()
		}
	}

	if s.connectionManager.secondaryConn == nil && keyOwner == Secondary {
		log.Printf("Secondary connection is nil, but signing key owner is secondary. Attempting to reconnect.")
		if s.connectionManager.primaryConn != nil {
			s.setSigningKeyOwner(Primary)
			s.connectionManager.primaryConn.Close()
		}
	}

	return nil
}
