package stellar

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/txnbuild"
	"github.com/threefoldtech/atomicswap/timings"
)

type (
	ParticipateOutput struct {
		InitiatorAddress      string `json:"partcipant"`
		HoldingAccountAddress string `json:"holdingaccount"`
		RefundTransaction     string `json:"refundtransaction"`
	}
)

// Participate as the second party in an atomic swap
func Participate(network string, participatorKeyPair *keypair.Full, cp1Addr string, amount string, secretHash []byte, asset txnbuild.Asset, client horizonclient.ClientInterface) (ParticipateOutput, error) {
	fundingAccountAddress := participatorKeyPair.Address()
	holdingAccountKeyPair, err := GenerateKeyPair()
	if err != nil {
		return ParticipateOutput{}, fmt.Errorf("Failed to create holding account keypair: %s", err)
	}
	holdingAccountAddress := holdingAccountKeyPair.Address()
	//TODO: print the holding account private key in case of an error further down this function
	//to recover the funds

	locktime := time.Now().Add(timings.LockTime / 2)
	refundTransaction, err := createAtomicSwapHoldingAccount(network, participatorKeyPair, holdingAccountKeyPair, cp1Addr, amount, secretHash, locktime, asset, client)
	if err != nil {
		return ParticipateOutput{}, errors.Wrap(err, "could not create holding account")
	}

	serializedRefundTx, err := refundTransaction.Base64()
	if err != nil {
		return ParticipateOutput{}, errors.Wrap(err, "can't encode refund transaction")
	}

	output := ParticipateOutput{
		InitiatorAddress:      fundingAccountAddress,
		HoldingAccountAddress: holdingAccountAddress,
		RefundTransaction:     serializedRefundTx,
	}
	return output, nil

}
