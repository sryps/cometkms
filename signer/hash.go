package signer

import (
	"crypto/sha256"
)

func (app *App) ComputeBlockHash(txs [][]byte) []byte {
	h := sha256.New()
	for _, tx := range txs {
		h.Write(tx)
	}
	return h.Sum(nil)
}
