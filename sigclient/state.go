package sigclient

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

func (s *SimpleSigner) saveState(vote *SigningState) error {
	// Write the signer state to the file
	data, err := json.MarshalIndent(vote, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}
	return os.WriteFile(s.stateFilePath, data, 0644)
}

// readState reads the signer state from the file.
func (s *SimpleSigner) ReadState() (*SigningState, error) {
	// Read the signer state from the file
	data, err := os.ReadFile(s.stateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &SigningState{}, fmt.Errorf("state file does not exist: %w", err)
		}
		log.Fatalf("Failed to read signer state: %v", err)
		return &SigningState{}, fmt.Errorf("failed to read state file: %w", err)
	}

	// Unmarshal the JSON data into a SigningState struct
	var state *SigningState
	if err := json.Unmarshal(data, &state); err != nil {
		return &SigningState{}, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return state, nil
}

func CreateStateFileIfNoneExists(statefilepath string) error {
	// Create the state file if it does not exist
	if _, err := os.Stat(statefilepath); os.IsNotExist(err) {
		initialState := SigningState{
			Type:    0,
			TypeStr: "unknown",
			Height:  0,
			Round:   0,
			BlockID: BlockID{
				BlockHash: nil,
				PartSetHeader: PartSetHeader{
					Hash:  nil,
					Total: 0,
				},
			},
			ValidatorAddress:   nil,
			Timestamp:          time.Time{},
			Signature:          nil,
			ExtensionSignature: nil,
			ChainId:            "",
		}

		// Marshal the initial state to JSON and write it to the file
		data, err := json.MarshalIndent(initialState, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal initial state: %w", err)
		}
		if err := os.WriteFile(statefilepath, data, 0644); err != nil {
			return fmt.Errorf("failed to create state file: %w", err)
		}
	}
	return nil
}
