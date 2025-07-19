package sigclient

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"encoding/base64"
	pbcrypto "github.com/cometbft/cometbft/api/cometbft/crypto/v1"
	"github.com/cometbft/cometbft/crypto/ed25519"
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

func PublicKeyToString(pk pbcrypto.PublicKey) (string, error) {
	switch pk := pk.Sum.(type) {
	case *pbcrypto.PublicKey_Ed25519:
		return fmt.Sprintf("ed25519:%s", base64.StdEncoding.EncodeToString(pk.Ed25519)), nil
	case *pbcrypto.PublicKey_Secp256K1:
		return fmt.Sprintf("secp256k1:%s", base64.StdEncoding.EncodeToString(pk.Secp256K1)), nil
	default:
		return "", fmt.Errorf("unsupported public key type %T", pk)
	}
}
