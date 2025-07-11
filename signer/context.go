package signer

import (
	"context"
	"log"
)

func (s *SimpleSigner) contextHandler(ctx context.Context) error {
	keyOwner := s.getSigningKeyOwner()

	if s.connectionManager.primaryConn == nil && s.connectionManager.secondaryConn == nil {
		log.Printf("Both primary and secondary connections are nil. Restarting connections...")
		s.setSigningKeyOwner(Primary) // Default to primary if both are nil
		return nil
	}

	if s.connectionManager.primaryConn == nil && keyOwner == Primary {
		log.Printf("Primary connection is nil, but signing key owner is primary...checking if secondary connection is up...")
		if s.connectionManager.secondaryConn != nil {
			log.Printf("Secondary connection is available, switching signing key owner to secondary.")
			s.setSigningKeyOwner(Secondary)
			s.connectionManager.secondaryConn.Close()
			<-ctx.Done()
		} else {
			log.Printf("No secondary connection available, retrying primary connection...")
			return nil
		}
	}

	if s.connectionManager.secondaryConn == nil && keyOwner == Secondary {
		log.Printf("Secondary connection is nil, but signing key owner is secondary...checking if primary connection is up...")
		if s.connectionManager.primaryConn != nil {
			log.Printf("Primary connection is available, switching signing key owner to primary.")
			s.setSigningKeyOwner(Primary)
			s.connectionManager.primaryConn.Close()
			<-ctx.Done()
		}

	}

	return nil
}
