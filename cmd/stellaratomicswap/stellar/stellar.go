package stellar

//Package stellar provides simple stellar specific functions for the stellar atomic swap

import (
	"errors"
	"fmt"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/strkey"
	"github.com/stellar/go/txnbuild"

	"github.com/stellar/go/protocols/horizon"
)

//GenerateKeyPair creates a new stellar full keypair
func GenerateKeyPair() (pair *keypair.Full, err error) {

	pair, err = keypair.Random()
	return
}

func createHashxAddress(hash []byte) (address string, err error) {
	return strkey.Encode(strkey.VersionByteHashX, hash)
}

func createHashTxAddress(hash []byte) (address string, err error) {
	return strkey.Encode(strkey.VersionByteHashTx, hash)
}

//GetAccount returns information for a single account
func GetAccount(address string, client horizonclient.ClientInterface) (account *horizon.Account, err error) {
	ar := horizonclient.AccountRequest{AccountID: address}
	accountStruct, err := client.AccountDetail(ar)
	if err != nil {
		return
	}
	account = &accountStruct
	return
}

//CreateHoldingAccount creates a new account to hold the atomic swap balance
//with the signers modified to the atomic swap rules:
//- signature of the destinee and the secret
//- hash of a specific transaction that is present on the chain
//    that merges the escrow account to the account that needs to withdraw
//    and that can only be published in the future ( timeout mechanism)
func CreateHoldingAccount(holdingAccountAddress string, xlmAmount string, withdrawalAccount *horizon.Account, counterPartyAddress string, secretHash []byte, refundTxHash []byte, network string) (createAccountTransaction txnbuild.Transaction, err error) {

	accountCreationOperation := txnbuild.CreateAccount{
		Destination:   holdingAccountAddress,
		Amount:        xlmAmount,
		SourceAccount: withdrawalAccount,
	}

	depositorSigningOperation := txnbuild.SetOptions{
		Signer: &txnbuild.Signer{
			Address: counterPartyAddress,
			Weight:  1,
		},
	}
	secretHashAddress, err := createHashxAddress(secretHash)
	if err != nil {
		return
	}
	secretSigningOperation := txnbuild.SetOptions{
		Signer: &txnbuild.Signer{
			Address: secretHashAddress,
			Weight:  1,
		},
	}
	refundTxHashAdddress, err := createHashTxAddress(refundTxHash)
	if err != nil {
		return
	}
	refundSigningOperation := txnbuild.SetOptions{
		Signer: &txnbuild.Signer{
			Address: refundTxHashAdddress,
			Weight:  2,
		},
	}
	setSigingWeightsOperation := txnbuild.SetOptions{
		MasterWeight:    txnbuild.NewThreshold(txnbuild.Threshold(uint8(0))),
		LowThreshold:    txnbuild.NewThreshold(txnbuild.Threshold(2)),
		MediumThreshold: txnbuild.NewThreshold(txnbuild.Threshold(2)),
		HighThreshold:   txnbuild.NewThreshold(txnbuild.Threshold(2)),
	}
	createAccountTransaction = txnbuild.Transaction{
		SourceAccount: withdrawalAccount,
		Operations: []txnbuild.Operation{
			&accountCreationOperation,
			&depositorSigningOperation,
			&secretSigningOperation,
			&refundSigningOperation,
			&setSigingWeightsOperation,
		},
		Network:    network,
		Timebounds: txnbuild.NewInfiniteTimeout(), //TODO: Use a real timeout
	}

	return
}

//SubmitTransaction submits the transactio and provides a better formatted error on failure
func SubmitTransaction(tx string, client horizonclient.ClientInterface) (txSuccess horizon.TransactionSuccess, err error) {

	txSuccess, err = client.SubmitTransactionXDR(tx)
	if err != nil {
		he := err.(*horizonclient.Error)
		errordetail := (he.Problem.Detail)
		if resultcodes, err2 := he.ResultCodes(); err2 == nil {
			errordetail = fmt.Sprintf("%s\nResultcodes:\n%s\n", errordetail, resultcodes)
		}

		errordetail = fmt.Sprintf("%sExtras:\n", errordetail)
		for _, ex := range he.Problem.Extras {
			errordetail = fmt.Sprintf("%s%s\n", errordetail, ex)
		}

		err = errors.New(errordetail)
	}
	return
}
