package eth

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type (
	AuditContractOutput struct {
		ContractAddress  common.Address    `json:"contractAddress"`
		ContractValue    *big.Int          `json:"contractValue"`
		RecipientAddress common.Address    `json:"recipientAddress"`
		RefundAddress    common.Address    `json:"refundAddress"`
		SecretHash       [sha256.Size]byte `json:"secretHash"`
		Locktime         int64             `json:"locktime"`
	}
)

func AuditContract(ctx context.Context, sct SwapContractTransactor, contractTx *types.Transaction) (AuditContractOutput, error) {
	// unpack input params from contract tx
	params, err := unpackContractInputParams(sct.abi, contractTx)
	if err != nil {
		return AuditContractOutput{}, err
	}

	rpcTransaction := struct {
		tx          *types.Transaction
		BlockNumber *string
		BlockHash   *common.Hash
		From        *common.Address
	}{}

	// get transaction by hash
	contractHash := contractTx.Hash()
	err = sct.client.rpcClient.CallContext(ctx,
		&rpcTransaction, "eth_getTransactionByHash", contractHash)
	if err != nil {
		return AuditContractOutput{}, fmt.Errorf(
			"failed to find transaction (%x): %v", contractHash, err)
	}
	if rpcTransaction.BlockNumber == nil || *rpcTransaction.BlockNumber == "" || *rpcTransaction.BlockNumber == "0" {
		return AuditContractOutput{}, fmt.Errorf("transaction (%x) is pending", contractHash)
	}

	// get block in order to know the timestamp of the txn
	block, err := sct.client.BlockByHash(ctx, *rpcTransaction.BlockHash)
	if err != nil {
		return AuditContractOutput{}, fmt.Errorf(
			"failed to find block (%x): %v", rpcTransaction.BlockHash, err)
	}

	// compute the locktime
	lockTime := time.Unix(int64(block.Time())+params.LockDuration.Int64(), 0)

	// NOTE:
	// the reason we require th node for this method,
	// is because we need to be able to know the transaction's timestamp

	return AuditContractOutput{
			ContractAddress:  *contractTx.To(),
			ContractValue:    contractTx.Value(),
			RecipientAddress: params.ToAddress,
			RefundAddress:    *rpcTransaction.From,
			SecretHash:       params.SecretHash,
			Locktime:         lockTime.Unix(),
		},
		nil
}
