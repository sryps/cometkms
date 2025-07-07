package signer

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/cometbft/cometbft/crypto/ed25519"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtencoding "github.com/cometbft/cometbft/crypto/encoding"
	pbcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
)

type KeyPair struct {
	Address string `json:"address"`
	Pubkey  struct {
		Type  string `json:"type"`
		Value []byte `json:"value"`
	} `json:"pub_key"`
	Privkey struct {
		Type  string `json:"type"`
		Value []byte `json:"value"`
	} `json:"priv_key"`
}

func LoadKeyFromFile(path string) (ed25519.PrivKey, ed25519.PubKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read key file: %w", err)
	}

	// unmarshall JSON from file
	var keyPair KeyPair
	if err := json.Unmarshal(keyData, &keyPair); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal key data: %w", err)
	}

	// if keyPair.Privkey.Type != "tendermint/PrivKeyEd25519" {
	if keyPair.Privkey.Type != "tendermint/PrivKeyEd25519" {
		return nil, nil, fmt.Errorf("unsupported private key type: %s", keyPair.Privkey.Type)
	}
	// if keyPair.Pubkey.Type != "tendermint/PubKeyEd25519" {
	if keyPair.Pubkey.Type != "tendermint/PubKeyEd25519" {
		return nil, nil, fmt.Errorf("unsupported public key type: %s", keyPair.Pubkey.Type)
	}
	keyData = keyPair.Privkey.Value
	if len(keyData) == 0 {
		return nil, nil, fmt.Errorf("private key data is empty")
	}

	if len(keyData) != ed25519.PrivateKeySize {
		return nil, nil, fmt.Errorf("invalid key length: expected %d, got %d", ed25519.PrivateKeySize, len(keyData))
	}

	privKey := ed25519.PrivKey(keyData)
	pubKey := privKey.PubKey().(ed25519.PubKey)
	log.Printf("Loaded private key for address %s", keyPair.Address)
	return privKey, pubKey, nil
}

func SetSecondaryKeys() (ed25519.PrivKey, pbcrypto.PublicKey) {
	tmpPrivKey := cmted25519.GenPrivKey()
	tmpPubKey, _ := cmtencoding.PubKeyToProto(tmpPrivKey.PubKey())
	// convert pub key to hex string for logging
	tmpPubKeyHex := fmt.Sprintf("%X", tmpPubKey.String()[:24])
	log.Printf("Generated new secondary public key: %s", tmpPubKeyHex)
	return tmpPrivKey, tmpPubKey
}
