package eth

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/threefoldtech/atomicswap/eth/contract"
)

// ExtractSecret from a redeem call to the contract
func ExtractSecret(ctx context.Context, sct SwapContractTransactor, redemptionTx *types.Transaction, secretHash [sha256.Size]byte) ([]byte, error) {
	abi, err := abi.JSON(strings.NewReader(contract.ContractABI))
	if err != nil {
		return nil, fmt.Errorf("failed to read (smart) contract ABI: %v", err)
	}

	txData := redemptionTx.Data()

	// first 4 bytes contain the id, so let's get method using that ID
	method, err := abi.MethodById(txData[:4])
	if err != nil {
		return nil, fmt.Errorf("failed to get method using its parsed id: %v", err)
	}
	if method.Name != "redeem" {
		return nil, fmt.Errorf("unexpected name for unpacked method ID: %s", method.Name)
	}

	// unpack the params
	rawParams, err := method.Inputs.Unpack(txData[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack method's input params: %v", err)
	}

	if len(rawParams) != 2 {
		return nil, errors.New("unexpected redeem call argument count")
	}

	secret, ok := rawParams[0].([32]byte)
	if !ok {
		return nil, errors.New("could not decode secret in redeem call")
	}
	contractSecretHash, ok := rawParams[0].([sha256.Size]byte)
	if !ok {
		return nil, errors.New("could not decode secret hash in redeem call")
	}

	// ensure secret hash is the same as the given one
	if secretHash != contractSecretHash {
		return nil, fmt.Errorf("unexpected secret hash found: %x", contractSecretHash)

	}
	computedSecretHash := sha256Hash(secret[:])
	if contractSecretHash != computedSecretHash {
		return nil, fmt.Errorf("unexpected secret found: %x", secret)
	}

	return secret[:], nil
}
