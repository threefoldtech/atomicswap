package stellar

import (
	"fmt"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/txnbuild"
)

type (
	RedeemOutput struct {
		RedeemTransactionTxHash string `json:"redeemTransaction"`
	}
)

func Redeem(network string, receiverKeyPair *keypair.Full, holdingAccountAddress string, secret []byte, client horizonclient.ClientInterface) (RedeemOutput, error) {
	holdingAccount, err := GetAccount(holdingAccountAddress, client)
	if err != nil {
		return RedeemOutput{}, err
	}
	receiverAddress := receiverKeyPair.Address()
	operations := createRedeemOperations(holdingAccount, receiverAddress)

	redeemTransactionParams := txnbuild.TransactionParams{
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewTimebounds(int64(0), int64(0)),
		},
		Operations:    operations,
		SourceAccount: holdingAccount,
	}

	redeemTransaction, err := txnbuild.NewTransaction(redeemTransactionParams)
	if err != nil {
		return RedeemOutput{}, fmt.Errorf("Unable to build the transaction: %v", err)
	}
	redeemTransaction, err = redeemTransaction.SignHashX(secret)
	if err != nil {
		return RedeemOutput{}, fmt.Errorf("Unable to sign with the secret:%v", err)
	}
	redeemTransaction, err = redeemTransaction.Sign(network, receiverKeyPair)
	if err != nil {
		return RedeemOutput{}, fmt.Errorf("Unable to sign with the receiver keypair:%v", err)
	}

	txSuccess, err := SubmitTransaction(redeemTransaction, client)
	if err != nil {
		return RedeemOutput{}, err
	}

	output := RedeemOutput{
		RedeemTransactionTxHash: fmt.Sprintf("%v", txSuccess.Hash),
	}
	return output, nil
}
