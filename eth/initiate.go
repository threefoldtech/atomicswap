package eth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type InitiateOutput struct {
	Secret              [32]byte          `json:"secret"`
	SecretHash          [32]byte          `json:"secretHash"`
	InitiatorAddress    common.Address    `json:"initiatorAddress"`
	ContractTransaction types.Transaction `json:"contractTransaction"`
}

// Initiate an atomic swap
func Initiate(ctx context.Context, sct SwapContractTransactor, cp2Addr common.Address, amount *big.Int) (InitiateOutput, error) {
	secret, secretHash := generateSecretHashPair()
	tx, err := sct.initiateTx(ctx, amount, secretHash, cp2Addr)
	if err != nil {
		return InitiateOutput{}, fmt.Errorf("failed to create initiate TX: %v", err)
	}

	err = tx.Send(ctx)
	if err != nil {
		return InitiateOutput{}, err
	}

	return InitiateOutput{
		Secret:              secret,
		SecretHash:          secretHash,
		InitiatorAddress:    sct.FromAddr,
		ContractTransaction: *tx.Transaction,
	}, nil
}

func generateSecretHashPair() (secret, secretHash [sha256.Size]byte) {
	rand.Read(secret[:])
	secretHash = sha256Hash(secret[:])
	return
}
