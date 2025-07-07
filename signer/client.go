package signer

import (
	"context"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	pbcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	"log"
	"sync"
	"time"
)

// Run starts the signer and handles one connection at a time.
func (s *SimpleSigner) Run() error {

	s.setSigningKeyOwner(Primary)
	var wg sync.WaitGroup
	wg.Add(2)

	// setup primary connection
	s.connectionManager.primaryCtx, s.connectionManager.primaryCancel = context.WithCancel(context.Background())
	s.connectionManager.secondaryCtx, s.connectionManager.secondaryCancel = context.WithCancel(context.Background())

	go func() {
		defer wg.Done()
		err := s.keyRunner(s.connectionManager.primaryCtx, Primary, s.connectionManager.primaryAddr)
		if err != nil {
			log.Printf("Failed to start primary key runner: %v", err)
		}
	}()

	// setup secondary connection if provided
	if s.connectionManager.secondaryAddr != "" {
		go func() {
			defer wg.Done()
			err := s.keyRunner(s.connectionManager.secondaryCtx, Secondary, s.connectionManager.secondaryAddr)
			if err != nil {
				log.Printf("Failed to start secondary key runner: %v", err)
			}
		}()
	}
	wg.Wait()
	return nil
}

func (s *SimpleSigner) keyRunner(ctx context.Context, role Role, addr string) error {
	for {
		select {
		case <-ctx.Done():
			log.Printf("Context done, stopping key runner for role %s", role)
			return nil
		default:
			privKey, pubKey := s.getKeysForRole(role)

			err := s.connectAndServe(ctx, addr, privKey, pubKey, role)
			if err != nil {
				log.Printf("Failed to connect for role %s: %v", role, addr)
				time.Sleep(2 * time.Second)

				// If the connection fails, update the connection manager
				if addr == s.connectionManager.primaryAddr {
					s.connectionManager.primaryConn = nil
				} else if addr == s.connectionManager.secondaryAddr {
					s.connectionManager.secondaryConn = nil
				}

				// If both connections are nil, continue to retry
				if s.connectionManager.primaryConn == nil && s.connectionManager.secondaryConn == nil {
					continue
				}

				contextHandlerErr := s.contextHandler()
				if contextHandlerErr != nil {
					log.Printf("Context handler error for role %s: %v", role, contextHandlerErr)
				}
			}
			log.Printf("Connection for role %s at %s closed", role, addr)
		}
	}
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
