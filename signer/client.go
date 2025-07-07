package signer

import (
	"context"
	"log"
	"sync"
	"time"
)

// Run starts the signer and handles one connection at a time.
func (s *SimpleSigner) Run(ctx context.Context) error {
	var err error
	// Set private key and public key for primary and secondary connections
	signingKey := s.signingKey
	signingPubkey := s.signingPubkey
	secSigningKey := s.secondaryNonsigningPrivkey
	nonsigningPubkey := s.secondaryNonsigningPubkey

	primAddr := s.connectionManager.primaryAddr
	secAddr := s.connectionManager.secondaryAddr

	var wg sync.WaitGroup

	wg.Add(2) // Launching 2 goroutines

	// setup primary connection
	go func() {
		defer wg.Done()
		for s.connectionManager.primaryConn == nil {
			select {
			case <-ctx.Done():
				log.Println("Context done, stopping primary connection attempt")
				return
			default:
				s.connectionManager.primaryConn, err = s.connectAndServe(ctx, primAddr, signingKey, signingPubkey)
				if err != nil {
					log.Printf("failed to connect to primary address %s", primAddr)
					time.Sleep(2 * time.Second) // wait before retrying
				}
			}
		}
	}()

	// setup secondary connection if provided
	if secAddr != "" {
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				log.Println("Context done, stopping secondary connection attempt")
				return
			default:
				for s.connectionManager.secondaryConn == nil {
					s.connectionManager.secondaryConn, err = s.connectAndServe(ctx, secAddr, secSigningKey, nonsigningPubkey)
					if err != nil {
						log.Printf("failed to connect to secondary address %s", secAddr)
						time.Sleep(2 * time.Second) // wait before retrying
					}
				}
			}
		}()
	}
	wg.Wait()
	return nil
}
