package signer

import (
	"context"
	"log"
	"sync"
	"time"

	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	pbcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
)

// Run starts the signer and handles one connection at a time.
func (s *SimpleSigner) Run(ctx context.Context) error {

	s.setSigningKeyOwner(Primary)

	var wg sync.WaitGroup
	wg.Add(1)

	// Start primary connection
	go func() {
		log.Println("Starting primary connection...")
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping primary key runner go routine...")
				return
			default:
				err := s.primaryConnection(ctx, s.connectionManager.primaryAddr, Primary)
				if err != nil {
					log.Printf("Failed to connect for %s role: %v", Primary, s.connectionManager.primaryAddr)
					time.Sleep(2 * time.Second)

					s.connectionManager.primaryConn = nil

					// Contest handler checks if primary or secondary connection is nil and swaps signing key owner if necessary
					contextHandlerErr := s.contextHandler(ctx)
					if contextHandlerErr != nil {
						log.Printf("Context handler error for role %s: %v", Primary, contextHandlerErr)
					}
				}
			}
		}
	}()

	// setup secondary connection if provided
	if s.connectionManager.secondaryAddr != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.secondaryConnection(ctx, s.connectionManager.secondaryAddr, Secondary)
			if err != nil {
				log.Printf("Failed to connect for %s role: %v", Secondary, s.connectionManager.secondaryAddr)
				time.Sleep(2 * time.Second)

				s.connectionManager.secondaryConn = nil

				// Contest handler checks if primary or secondary connection is nil and swaps signing key owner if necessary
				contextHandlerErr := s.contextHandler(ctx)
				if contextHandlerErr != nil {
					log.Printf("Context handler error for role %s: %v", Secondary, contextHandlerErr)
				}
			}
		}()
	}

	wg.Wait()
	log.Println("All key runners have finished. Exiting signer.")
	<-ctx.Done()
	return nil
}

func (s *SimpleSigner) getKeysForRole(role Role) (cmted25519.PrivKey, pbcrypto.PublicKey) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.signingKeyOwner == role {
		return s.signingKey, s.signingPubkey
	}
	return s.secondaryNonsigningPrivkey, s.secondaryNonsigningPubkey
}

func (s *SimpleSigner) setSigningKeyOwner(role Role) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.signingKeyOwner = role
}

func (s *SimpleSigner) getSigningKeyOwner() Role {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.signingKeyOwner
}

func (s *SimpleSigner) swapSigningKeyOwner() Role {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.signingKeyOwner == Primary {
		s.signingKeyOwner = Secondary
	} else {
		s.signingKeyOwner = Primary
	}

	log.Printf("Swapped signing key owner to: %s", s.signingKeyOwner)
	return s.signingKeyOwner
}
