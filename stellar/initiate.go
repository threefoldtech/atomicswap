package stellar

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/threefoldtech/atomicswap/timings"
)

type (
	// InitiateOutput is the result of the Initiate call
	InitiateOutput struct {
		// Secret is the hex encoded secret
		Secret [secretSize]byte `json:"secret"`
		// SecretHash is the hex encoded SHA256 hash of the secret
		SecretHash []byte `json:"hash"`
		// InitiatorAddress is the address of the initiator keypair
		InitiatorAddress string `json:"initiator"`
		// HoldingAccountAddress is the address of the holding account
		HoldingAccountAddress string `json:"holdingaccount"`
		// RefundTransaction is the base64 encoded refund transaction
		RefundTransaction string `json:"refundtransaction"`
	}
)

const (
	secretSize = 32
)

func Initiate(network string, initiatorKeyPair *keypair.Full, destination string, amount string, asset txnbuild.Asset, client horizonclient.ClientInterface) (InitiateOutput, error) {
	if _, err := keypair.ParseAddress(destination); err != nil {
		return InitiateOutput{}, errors.Wrap(err, "could not decode destination address")
	}

	var secret [secretSize]byte
	_, err := rand.Read(secret[:])
	if err != nil {
		return InitiateOutput{}, err
	}
	secretHash := sha256Hash(secret[:])
	fundingAccountAddress := initiatorKeyPair.Address()
	holdingAccountKeyPair, err := GenerateKeyPair()
	if err != nil {
		return InitiateOutput{}, errors.Wrap(err, "failed to create holding account keypair")
	}
	holdingAccountAddress := holdingAccountKeyPair.Address()
	//TODO: print the holding account private key in case of an error further down this function
	//to recover the funds

	locktime := time.Now().Add(timings.LockTime)
	refundTransaction, err := createAtomicSwapHoldingAccount(network, initiatorKeyPair, holdingAccountKeyPair, destination, amount, secretHash, locktime, asset, client)
	if err != nil {
		return InitiateOutput{}, err
	}

	serializedRefundTx, err := refundTransaction.Base64()
	if err != nil {
		return InitiateOutput{}, err
	}

	output := InitiateOutput{
		Secret:                secret,
		SecretHash:            secretHash,
		InitiatorAddress:      fundingAccountAddress,
		HoldingAccountAddress: holdingAccountAddress,
		RefundTransaction:     serializedRefundTx,
	}
	return output, nil
}

func sha256Hash(x []byte) []byte {
	h := sha256.Sum256(x)
	return h[:]
}

func createAtomicSwapHoldingAccount(network string, fundingKeyPair *keypair.Full, holdingAccountKeyPair *keypair.Full, counterPartyAddress string, amount string, secretHash []byte, locktime time.Time, asset txnbuild.Asset, client horizonclient.ClientInterface) (refundTransaction *txnbuild.Transaction, err error) {
	holdingAccountAddress := holdingAccountKeyPair.Address()

	xlmAmount := "10"
	if asset.IsNative() {
		xlmAmount = amount
	}
	err = createHoldingAccount(holdingAccountAddress, xlmAmount, fundingKeyPair, network, asset, client)
	if err != nil {
		err = errors.Wrap(err, "could not create holding account")
		return
	}

	if !asset.IsNative() {
		err = fundHoldingAccount(network, fundingKeyPair, holdingAccountKeyPair, amount, asset, client)
		if err != nil {
			err = errors.Wrap(err, "could not fund holding account")
			return
		}
	}

	refundTransaction, err = createRefundTransaction(holdingAccountAddress, fundingKeyPair.Address(), locktime, client)
	if err != nil {
		err = errors.Wrap(err, "could not create refund transaction")
		return
	}
	refundTransactionHash, err := refundTransaction.Hash(network)
	if err != nil {
		err = fmt.Errorf("Failed to Hash the refund transaction: %s", err)
		return
	}
	err = setHoldingAccountSigningOptions(holdingAccountKeyPair, counterPartyAddress, secretHash, refundTransactionHash[:], network, client)

	return
}

// createHoldingAccount creates a new account to hold the atomic swap balance
func createHoldingAccount(holdingAccountAddress string, amount string, fundingKeyPair *keypair.Full, network string, asset txnbuild.Asset, client horizonclient.ClientInterface) (err error) {
	fundingAccount, err := GetAccount(fundingKeyPair.Address(), client)
	if err != nil {
		return
	}
	createAccountTransaction, err := CreateAccountTransaction(holdingAccountAddress, amount, fundingAccount, network)
	if err != nil {
		return fmt.Errorf("Failed to create the holding account transaction: %s", err)
	}
	tx, err := createAccountTransaction.Sign(network, fundingKeyPair)
	if err != nil {
		return fmt.Errorf("Failed to sign the holding account transaction: %s", err)
	}
	_, err = SubmitTransaction(tx, client)
	if err != nil {
		accountID, err2 := createAccountTransaction.HashHex(network)
		if err2 != nil {
			panic(err2)
		}
		return fmt.Errorf("Failed to publish the holding account creation transaction : %s\n%s", accountID, err)
	}
	return
}

func createHoldingAccountSigningTransaction(holdingAccount *horizon.Account, counterPartyAddress string, secretHash []byte, refundTxHash []byte, network string) (setOptionsTransaction *txnbuild.Transaction, err error) {
	depositorSigningOperation := txnbuild.SetOptions{
		Signer: &txnbuild.Signer{
			Address: counterPartyAddress,
			Weight:  1,
		},
		SourceAccount: holdingAccount.GetAccountID(),
	}
	secretHashAddress, err := CreateHashxAddress(secretHash)
	if err != nil {
		return
	}
	secretSigningOperation := txnbuild.SetOptions{
		Signer: &txnbuild.Signer{
			Address: secretHashAddress,
			Weight:  1,
		},
		SourceAccount: holdingAccount.GetAccountID(),
	}
	refundTxHashAdddress, err := CreateHashTxAddress(refundTxHash)
	if err != nil {
		return
	}
	refundSigningOperation := txnbuild.SetOptions{
		Signer: &txnbuild.Signer{
			Address: refundTxHashAdddress,
			Weight:  2,
		},
		SourceAccount: holdingAccount.GetAccountID(),
	}
	setSigningWeightsOperation := txnbuild.SetOptions{
		MasterWeight:    txnbuild.NewThreshold(txnbuild.Threshold(uint8(0))),
		LowThreshold:    txnbuild.NewThreshold(txnbuild.Threshold(2)),
		MediumThreshold: txnbuild.NewThreshold(txnbuild.Threshold(2)),
		HighThreshold:   txnbuild.NewThreshold(txnbuild.Threshold(2)),
		SourceAccount:   holdingAccount.GetAccountID(),
	}
	setOptionsTransactionParams := txnbuild.TransactionParams{
		SourceAccount: holdingAccount, //TODO: check if this can be changed to the fundingaccount
		Operations: []txnbuild.Operation{
			&depositorSigningOperation,
			&secretSigningOperation,
			&refundSigningOperation,
			&setSigningWeightsOperation,
		},
		IncrementSequenceNum: true,
		BaseFee:              200000,
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewInfiniteTimeout(), //TODO: Use a real timeout
		},
	}

	setOptionsTransaction, err = txnbuild.NewTransaction(setOptionsTransactionParams)

	return
}

func setHoldingAccountSigningOptions(holdingAccountKeyPair *keypair.Full, counterPartyAddress string, secretHash []byte, refundTxHash []byte, network string, client horizonclient.ClientInterface) (err error) {
	holdingAccountAddress := holdingAccountKeyPair.Address()
	holdingAccount, err := GetAccount(holdingAccountAddress, client)
	if err != nil {
		return
	}
	setSigningOptionsTransaction, err := createHoldingAccountSigningTransaction(holdingAccount, counterPartyAddress, secretHash, refundTxHash, network)
	if err != nil {
		return fmt.Errorf("Failed to create the signing options transaction: %s", err)
	}
	tx, err := setSigningOptionsTransaction.Sign(network, holdingAccountKeyPair)
	if err != nil {
		return fmt.Errorf("Failed to sign the signing options transaction: %s", err)
	}
	_, err = SubmitTransaction(tx, client)
	if err != nil {
		return fmt.Errorf("Failed to publish the signing options transaction : %s", err)
	}
	return
}

func fundHoldingAccount(network string, fundingKeyPair *keypair.Full, holdingAccountKeyPair *keypair.Full, amount string, asset txnbuild.Asset, client horizonclient.ClientInterface) (err error) {
	holdingAccount, err := GetAccount(holdingAccountKeyPair.Address(), client)
	if err != nil {
		return
	}

	changetrust := txnbuild.ChangeTrust{
		Line:          txnbuild.ChangeTrustAssetWrapper{Asset: txnbuild.CreditAsset{Code: asset.GetCode(), Issuer: asset.GetIssuer()}},
		Limit:         amount,
		SourceAccount: holdingAccount.GetAccountID(),
	}
	fundingAccount, err := GetAccount(fundingKeyPair.Address(), client)
	if err != nil {
		return
	}
	payment := txnbuild.Payment{
		Destination:   holdingAccount.AccountID,
		Amount:        amount,
		Asset:         asset,
		SourceAccount: fundingAccount.GetAccountID(),
	}

	txParams := txnbuild.TransactionParams{
		SourceAccount:        fundingAccount,
		Operations:           []txnbuild.Operation{&changetrust, &payment},
		IncrementSequenceNum: true,
		BaseFee:              200000,
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewInfiniteTimeout(), // Use a real timeout in production!
		},
	}
	tx, err := txnbuild.NewTransaction(txParams)
	if err != nil {
		err = errors.Wrap(err, "failed to build funding transaction")
		return
	}
	txe, err := tx.Sign(network, holdingAccountKeyPair, fundingKeyPair)
	if err != nil {
		err = fmt.Errorf("Failed to build,sign and encode the funding transaction: %v", err)
		return
	}
	_, err = SubmitTransaction(txe, client)
	if err != nil {
		transactionID, _ := tx.HashHex(network)
		err = fmt.Errorf("Failed to publish the funding transaction : %s\n%s", transactionID, err)
		return
	}
	return
}

func createRefundTransaction(holdingAccountAddress string, refundAccountAdress string, locktime time.Time, client horizonclient.ClientInterface) (refundTransaction *txnbuild.Transaction, err error) {
	holdingAccount, err := GetAccount(holdingAccountAddress, client)
	if err != nil {
		return
	}
	_, err = holdingAccount.IncrementSequenceNumber()
	if err != nil {
		err = fmt.Errorf("Unable to increment the sequence number of the holding account:%v", err)
		return
	}

	operations := createRedeemOperations(holdingAccount, refundAccountAdress)

	refundTransactionParams := txnbuild.TransactionParams{
		Operations:           operations,
		SourceAccount:        holdingAccount,
		IncrementSequenceNum: true,
		BaseFee:              200000,
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewTimebounds(locktime.Unix(), int64(0)),
		},
	}

	refundTransaction, err = txnbuild.NewTransaction(refundTransactionParams)
	if err != nil {
		err = fmt.Errorf("Failed to build the refund transaction: %s", err)
		return
	}
	return
}

func createRedeemOperations(holdingAccount *horizon.Account, receiverAddress string) (redeemOperations []txnbuild.Operation) {
	redeemOperations = make([]txnbuild.Operation, 0, len(holdingAccount.Balances))
	for _, balance := range holdingAccount.Balances {
		if balance.Asset.Type == NativeAssetType {
			continue
		}
		payment := txnbuild.Payment{
			Destination: receiverAddress,
			Amount:      balance.Balance,
			Asset: txnbuild.CreditAsset{
				Code:   balance.Code,
				Issuer: balance.Issuer,
			}}
		redeemOperations = append(redeemOperations, &payment)

		removetrust := txnbuild.ChangeTrust{
			Line:          txnbuild.ChangeTrustAssetWrapper{Asset: txnbuild.CreditAsset{Code: balance.Code, Issuer: balance.Issuer}},
			Limit:         "0",
			SourceAccount: holdingAccount.GetAccountID(),
		}
		redeemOperations = append(redeemOperations, &removetrust)
	}

	mergeAccountOperation := txnbuild.AccountMerge{
		Destination:   receiverAddress,
		SourceAccount: holdingAccount.GetAccountID(),
	}
	redeemOperations = append(redeemOperations, &mergeAccountOperation)

	return
}
