package stellar

//Package stellar provides simple stellar specific functions for the stellar atomic swap

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/protocols/horizon/effects"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/strkey"
	"github.com/stellar/go/txnbuild"
)

// NativeAssetType is the value rturned by the horizon client for a the native asset
const NativeAssetType = "native"

// GenerateKeyPair creates a new stellar full keypair
func GenerateKeyPair() (pair *keypair.Full, err error) {

	pair, err = keypair.Random()
	return
}

// CreateHashxAddress creates the stellar address for a Hashx signer
func CreateHashxAddress(hash []byte) (address string, err error) {
	return strkey.Encode(strkey.VersionByteHashX, hash)
}

// CreateHashTxAddress creates the stellar address for a HashTx signer
func CreateHashTxAddress(hash []byte) (address string, err error) {
	return strkey.Encode(strkey.VersionByteHashTx, hash)
}

// GetAccount returns information for a single account
func GetAccount(address string, client horizonclient.ClientInterface) (account *horizon.Account, err error) {
	ar := horizonclient.AccountRequest{AccountID: address}
	accountStruct, err := client.AccountDetail(ar)
	if err != nil {
		err = fmt.Errorf("Failed to get account details for account %s: %v", address, err)
		return
	}
	account = &accountStruct
	return
}

func getIDFromLink(href string) string {
	splittedHref := strings.Split(href, "/")
	return splittedHref[len(splittedHref)-1]
}

// GetAccountDebitediTransactions returns the transactions that debited the account
func GetAccountDebitediTransactions(accountAddress string, client horizonclient.ClientInterface) (transactions []horizon.Transaction, err error) {
	effectRequest := horizonclient.EffectRequest{ForAccount: accountAddress, Limit: 100}
	effect, err := client.Effects(effectRequest)
	if err != nil {
		return
	}
	transactions = make([]horizon.Transaction, 0, 1)
	for _, effectRecord := range effect.Embedded.Records {
		if effectRecord.GetType() != effects.EffectTypeNames[effects.EffectAccountDebited] {
			continue
		}
		realEffect, ok := effectRecord.(effects.AccountDebited)
		if !ok {
			return nil, fmt.Errorf("effect is not a horizon protocol AccountDebited effect but a %v", reflect.TypeOf(effectRecord))
		}
		operationID := getIDFromLink(realEffect.Links.Operation.Href)

		operation, err := client.OperationDetail(operationID)
		if err != nil {
			return nil, fmt.Errorf("Failed to get the operation with ID %v", operationID)
		}
		transactionHash := operation.GetTransactionHash()
		transaction, err := client.TransactionDetail(transactionHash)
		if err != nil {
			return nil, fmt.Errorf("Failed to get the transaction with hash %v", transactionHash)
		}
		transactions = append(transactions, transaction)
	}
	return
}

// GetNetworkPassPhrase fetches the networkPassphrase from a client
func GetNetworkPassPhrase(client horizonclient.Client) (networkpassphrase string, err error) {
	r, err := client.Root()
	if err != nil {
		err = fmt.Errorf("Failed to get the root from the client: %v", err)
		return
	}
	networkpassphrase = r.NetworkPassphrase
	return
}

// CreateAccountTransaction creates the transactio for creating a new account
func CreateAccountTransaction(newccountAddress string, xlmAmount string, fundingAccount *horizon.Account, network string) (createAccountTransaction *txnbuild.Transaction, err error) {

	accountCreationOperation := txnbuild.CreateAccount{
		Destination:   newccountAddress,
		Amount:        xlmAmount,
		SourceAccount: fundingAccount.GetAccountID(),
	}

	createAccountTransactionParams := txnbuild.TransactionParams{
		SourceAccount: fundingAccount,
		Operations: []txnbuild.Operation{
			&accountCreationOperation,
		},
		IncrementSequenceNum: true,
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewInfiniteTimeout(), //TODO: Use a real timeout
		},
		BaseFee: 200000,
	}

	createAccountTransaction, err = txnbuild.NewTransaction(createAccountTransactionParams)

	return
}

// SubmitTransaction submits the transactio and provides a better formatted error on failure
func SubmitTransaction(tx *txnbuild.Transaction, client horizonclient.ClientInterface) (txSuccess horizon.Transaction, err error) {

	txSuccess, err = client.SubmitTransaction(tx)
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
