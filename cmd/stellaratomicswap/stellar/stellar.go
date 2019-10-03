package stellar

//Package stellar provides simple stellar specific functions for the stellar atomic swap

import (
	"github.com/stellar/go/keypair"
)

//GenerateKeyPair creates a new stellar full keypair
func GenerateKeyPair() (pair *keypair.Full, err error) {

	pair, err = keypair.Random()
	return
}
