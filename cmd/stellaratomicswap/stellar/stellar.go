package stellar

//Package stellar provides simple stellar specific functions for the stellar atomic swap

import (
	"errors"
	"fmt"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/strkey"

	"github.com/stellar/go/protocols/horizon"
)

//GenerateKeyPair creates a new stellar full keypair
func GenerateKeyPair() (pair *keypair.Full, err error) {

	pair, err = keypair.Random()
	return
}

//CreateHashxAddress creates the stellar address for a Hashx signer
func CreateHashxAddress(hash []byte) (address string, err error) {
	return strkey.Encode(strkey.VersionByteHashX, hash)
}

//CreateHashTxAddress creates the stellar address for a HashTx signer
func CreateHashTxAddress(hash []byte) (address string, err error) {
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
