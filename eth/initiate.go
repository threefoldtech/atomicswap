package eth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

type InitiateOutput struct {
	Secret                  [32]byte       `json:"secret"`
	SecretHash              [32]byte       `json:"secretHash"`
	InitiatorAddress        common.Address `json:"initiatorAddress"`
	ContractTransactionHash common.Hash    `json:"contractTransactionHash"`
}

// Initiate an atomic swap
func Initiate(ctx context.Context, sct swapContractTransactor, cp2Addr common.Address, amount *big.Int) (InitiateOutput, error) {
	secret, secretHash := generateSecretHashPair()
	tx, err := sct.initiateTx(ctx, amount, secretHash, cp2Addr)
	if err != nil {
		return InitiateOutput{}, fmt.Errorf("failed to create initiate TX: %v", err)
	}

	txBytes, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return InitiateOutput{}, fmt.Errorf("failed to encode contract TX: %v", err)
	}
	fmt.Printf("%x\n\n", txBytes)

	err = tx.Send(ctx)
	if err != nil {
		return InitiateOutput{}, err
	}
	fmt.Printf("Published contract transaction (%x)\n", tx.Hash())

	return InitiateOutput{
		Secret:                  secret,
		SecretHash:              secretHash,
		InitiatorAddress:        sct.fromAddr,
		ContractTransactionHash: tx.Hash(),
	}, nil
}

func generateSecretHashPair() (secret, secretHash [sha256.Size]byte) {
	rand.Read(secret[:])
	secretHash = sha256Hash(secret[:])
	return
}
