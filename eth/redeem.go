package eth

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type (
	RedeemOutput struct {
		RedeemTxHash common.Hash `json:"redeemTransactionHash"`
	}

	params struct {
		LockDuration *big.Int
		SecretHash   [sha256.Size]byte
		ToAddress    common.Address
	}
)

// Redeem an atomic swap
func Redeem(ctx context.Context, sct SwapContractTransactor, contractTx *types.Transaction, secret [32]byte) (RedeemOutput, error) {
	params, err := unpackContractInputParams(sct.abi, contractTx)
	if err != nil {
		return RedeemOutput{}, err
	}
	tx, err := sct.redeemTx(ctx, params.SecretHash, secret)
	if err != nil {
		return RedeemOutput{}, fmt.Errorf("failed to create redeem TX: %v", err)
	}

	err = tx.Send(ctx)
	if err != nil {
		return RedeemOutput{}, err
	}
	return RedeemOutput{
		RedeemTxHash: tx.Hash(),
	}, nil
}

func unpackContractInputParams(abi abi.ABI, tx *types.Transaction) (params params, err error) {
	txData := tx.Data()

	// first 4 bytes contain the id, so let's get method using that ID
	method, err := abi.MethodById(txData[:4])
	if err != nil {
		err = fmt.Errorf("failed to get method using its parsed id: %v", err)
		return
	}

	rawParams, err := method.Inputs.Unpack(txData[4:])
	if err != nil {
		err = fmt.Errorf("failed to unpack method's input params: %v", err)
	}

	if len(rawParams) != 3 {
		err = errors.New("unexpected argument count")
		return
	}
	lockDuration, ok := rawParams[0].(*big.Int)
	if !ok {
		err = errors.New("could not parse lock duration")
		return
	}
	secretHash, ok := rawParams[1].([sha256.Size]byte)
	if !ok {
		err = errors.New("could not parse secret hash")
		return
	}
	toAddress, ok := rawParams[2].(common.Address)
	if !ok {
		err = errors.New("could not parse to address")
		return
	}
	params.LockDuration = lockDuration
	params.SecretHash = secretHash
	params.ToAddress = toAddress
	return
}
