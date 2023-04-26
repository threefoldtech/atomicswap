package eth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type (
	ParticipateOutput struct {
		InitiatorAddress        common.Address `json:"initiatorAddress"`
		ContractTransactionHash common.Hash    `json:"contractTransactionHash"`
	}
)

// Participate in an atomic swap
func Participate(ctx context.Context, sct SwapContractTransactor, cp1Addr common.Address, amount *big.Int, secretHash [32]byte) (ParticipateOutput, error) {
	tx, err := sct.participateTx(ctx, amount, secretHash, cp1Addr)
	if err != nil {
		return ParticipateOutput{}, fmt.Errorf("failed to create participate TX: %v", err)
	}

	err = tx.Send(ctx)
	if err != nil {
		return ParticipateOutput{}, err
	}

	return ParticipateOutput{
		InitiatorAddress:        sct.fromAddr,
		ContractTransactionHash: tx.Hash(),
	}, nil
}
