package signer

import (
	"fmt"
)

func (s *SimpleSigner) String(x any) string {
	return fmt.Sprintf("%s", x)
}
