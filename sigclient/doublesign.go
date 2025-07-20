package sigclient

import (
	"fmt"
	pbprivval "github.com/cometbft/cometbft/proto/tendermint/privval"
	"github.com/dgraph-io/badger/v4"
	"log"
)

func (s *SimpleSigner) isDoubleSignAttempt(req *pbprivval.SignVoteRequest) bool {
	heightStr := fmt.Sprintf("%d", req.Vote.Height)
	roundStr := fmt.Sprintf("%d", req.Vote.Round)
	stepStr := req.Vote.Type.String()
	dbKey := []byte(heightStr + ":" + roundStr + ":" + stepStr)
	log.Printf("Checking db key for double sign: %s", dbKey)

	var exists bool

	err := s.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(dbKey)
		if err == nil {
			exists = true
		} else if err == badger.ErrKeyNotFound {
			exists = false
		} else {
			return err
		}
		return nil
	})

	if err != nil {
		// You may want to log this or handle it differently
		fmt.Printf("Badger view error: %v\n", err)
		return false
	}

	return exists
}
