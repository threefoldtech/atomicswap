package eth

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func Refund(ctx context.Context, sct SwapContractTransactor, contractTx *types.Transaction) (common.Hash, error) {
	params, err := unpackContractInputParams(sct.Abi, contractTx)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := sct.refundTx(ctx, params.SecretHash)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to create refund TX: %v", err)
	}

	err = tx.Send(ctx)
	if err != nil {
		return common.Hash{}, err
	}
	return tx.Hash(), nil

}
