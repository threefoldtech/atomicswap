package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/stellar/go/xdr"

	"github.com/pkg/errors"

	"github.com/stellar/go/strkey"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/protocols/horizon"
	hprotocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/threefoldtech/atomicswap/stellar"
	"github.com/threefoldtech/atomicswap/timings"

	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
)

const verify = true

const secretSize = 32

var (
	targetNetwork = network.PublicNetworkPassphrase
)
var (
	flagset       = flag.NewFlagSet("", flag.ExitOnError)
	testnetFlag   = flagset.Bool("testnet", false, "use testnet network")
	automatedFlag = flagset.Bool("automated", false, "Use automated/unattended version with json output")
	assetParam    = flagset.String("asset", "", "The asset to transfer in case of non native XLM, format: `code:issuer`")
)

// There are two directions that the atomic swap can be performed, as the
// initiator can be on either chain.  This tool only deals with creating the
// Stellar transactions for these swaps.  A second tool should be used for the
// transaction on the other chain.  Any chain can be used so long as it supports
// OP_SHA256 and OP_CHECKLOCKTIMEVERIFY.
//
// Example scenarios using bitcoin as the second chain:
//
// Scenerio 1:
//   cp1 initiates (dcr)
//   cp2 participates with cp1 H(S) (xlm)
//   cp1 redeems xlm revealing S
//     - must verify H(S) in contract is hash of known secret
//   cp2 redeems dcr with S
//
// Scenerio 2:
//   cp1 initiates (xlm)
//   cp2 participates with cp1 H(S) (dcr)
//   cp1 redeems dcr revealing S
//     - must verify H(S) in contract is hash of known secret
//   cp2 redeems xlm with S

func init() {
	flagset.Usage = func() {
		fmt.Println("Usage: stellaratomicswap [flags] cmd [cmd args]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  initiate [-asset code:issuer] <initiator seed> <participant address> <amount>")
		fmt.Println("  participate [-asset code:issuer]  <participant seed> <initiator address> <amount> <secret hash>")
		fmt.Println("  redeem <receiver seed> <holdingAccountAdress> <secret>")
		fmt.Println("  refund <refund transaction>")
		fmt.Println("  extractsecret <holdingAccountAdress> <secret hash>")
		fmt.Println("  auditcontract <holdingAccountAdress> < refund transaction>")
		fmt.Println()
		fmt.Println("Flags:")
		flagset.PrintDefaults()
	}
}

type command interface {
	runCommand(client horizonclient.ClientInterface) error
}

// offline commands don't require wallet RPC.
type offlineCommand interface {
	command
	runOfflineCommand() error
}

type initiateCmd struct {
	InitiatorKeyPair *keypair.Full
	cp2Addr          string
	amount           string
	asset            txnbuild.Asset
}

type participateCmd struct {
	cp1Addr             string
	participatorKeyPair *keypair.Full
	amount              string
	secretHash          []byte
	asset               txnbuild.Asset
}

type redeemCmd struct {
	ReceiverKeyPair       *keypair.Full
	holdingAccountAddress string
	secret                []byte
}

type refundCmd struct {
	refundTx txnbuild.Transaction
}

type extractSecretCmd struct {
	holdingAccountAdress string
	secretHash           string
}

type auditContractCmd struct {
	refundTx             txnbuild.Transaction
	holdingAccountAdress string
}

func main() {
	showUsage, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if showUsage {
		flagset.Usage()
	}
	if err != nil || showUsage {
		os.Exit(1)
	}
}

func checkCmdArgLength(args []string, required int) (nArgs int) {
	if len(args) < required {
		return 0
	}
	for i, arg := range args[:required] {
		if len(arg) != 1 && strings.HasPrefix(arg, "-") {
			return i
		}
	}
	return required
}
func run() (showUsage bool, err error) {

	flagset.Parse(os.Args[1:])
	args := flagset.Args()
	var asset txnbuild.Asset
	if *assetParam != "" {
		assetparts := strings.SplitN(*assetParam, ":", 2)
		if len(assetparts) != 2 {
			return true, errors.New("Invalid asset format")
		}
		asset = txnbuild.CreditAsset{
			Code:   assetparts[0],
			Issuer: assetparts[1],
		}
	} else {
		asset = txnbuild.NativeAsset{}
	}
	if len(args) == 0 {
		return true, nil
	}
	cmdArgs := 0
	switch args[0] {
	case "initiate":
		cmdArgs = 3
	case "participate":
		cmdArgs = 4
	case "redeem":
		cmdArgs = 3
	case "refund":
		cmdArgs = 1
	case "extractsecret":
		cmdArgs = 2
	case "auditcontract":
		cmdArgs = 2
	default:
		return true, fmt.Errorf("unknown command %v", args[0])
	}
	nArgs := checkCmdArgLength(args[1:], cmdArgs)
	flagset.Parse(args[1+nArgs:])
	if nArgs < cmdArgs {
		return true, fmt.Errorf("%s: too few arguments", args[0])
	}
	if flagset.NArg() != 0 {
		return true, fmt.Errorf("unexpected argument: %s", flagset.Arg(0))
	}

	if *testnetFlag {
		targetNetwork = network.TestNetworkPassphrase
	}

	var client horizonclient.ClientInterface
	switch targetNetwork {
	case network.PublicNetworkPassphrase:
		client = horizonclient.DefaultPublicNetClient
	case network.TestNetworkPassphrase:
		client = horizonclient.DefaultTestNetClient

	}

	var cmd command
	switch args[0] {
	case "initiate":
		initiatorKeypair, err := keypair.Parse(args[1])
		if err != nil {
			return true, fmt.Errorf("invalid initiator seed: %v", err)
		}
		initiatorFullKeypair, ok := initiatorKeypair.(*keypair.Full)
		if !ok {
			return true, errors.New("invalid initiator seed")
		}

		_, err = keypair.Parse(args[2])
		if err != nil {
			return true, fmt.Errorf("invalid participant address: %v", err)
		}

		_, err = strconv.ParseFloat(args[3], 64)
		if err != nil {
			return true, fmt.Errorf("failed to decode amount: %v", err)
		}

		cmd = &initiateCmd{InitiatorKeyPair: initiatorFullKeypair, cp2Addr: args[2], amount: args[3], asset: asset}
	case "participate":
		participatorKeypair, err := keypair.Parse(args[1])
		if err != nil {
			return true, fmt.Errorf("invalid participator seed: %v", err)
		}
		participatorFullKeypair, ok := participatorKeypair.(*keypair.Full)
		if !ok {
			return true, errors.New("invalid participator seed")
		}

		_, err = keypair.Parse(args[2])
		if err != nil {
			return true, fmt.Errorf("invalid initiator address: %v", err)
		}

		_, err = strconv.ParseFloat(args[3], 64)
		if err != nil {
			return true, fmt.Errorf("failed to decode amount: %v", err)
		}

		secretHash, err := hex.DecodeString(args[4])
		if err != nil {
			return true, errors.New("secret hash must be hex encoded")
		}
		if len(secretHash) != sha256.Size {
			return true, errors.New("secret hash has wrong size")
		}
		cmd = &participateCmd{participatorKeyPair: participatorFullKeypair, cp1Addr: args[2], amount: args[3], secretHash: secretHash, asset: asset}
	case "auditcontract":
		_, err = keypair.Parse(args[1])
		if err != nil {
			return true, fmt.Errorf("invalid holding account address: %v", err)
		}
		genericTransaction, err := txnbuild.TransactionFromXDR(args[2])
		if err != nil {
			return true, fmt.Errorf("failed to decode refund transaction: %v", err)
		}
		refundTransaction, ok := genericTransaction.Transaction()
		if !ok {
			return true, errors.New("transaction XDR does not contain an actual transaction")
		}
		cmd = &auditContractCmd{holdingAccountAdress: args[1], refundTx: *refundTransaction}
	case "refund":

		genericTransaction, err := txnbuild.TransactionFromXDR(args[1])
		if err != nil {
			return true, fmt.Errorf("failed to decode refund transaction: %v", err)
		}
		refundTransaction, ok := genericTransaction.Transaction()
		if !ok {
			return true, errors.New("transaction XDR does not contain an actual transaction")
		}
		cmd = &refundCmd{refundTx: *refundTransaction}
	case "redeem":

		receiverKeypair, err := keypair.Parse(args[1])
		if err != nil {
			return true, fmt.Errorf("invalid receiver seed: %v", err)
		}
		receiverFullKeypair, ok := receiverKeypair.(*keypair.Full)
		if !ok {
			return true, errors.New("invalid receiver seed")
		}
		_, err = keypair.Parse(args[2])
		if err != nil {
			return true, fmt.Errorf("invalid holding account address: %v", err)
		}
		secret, err := hex.DecodeString(args[3])
		if err != nil {
			return true, fmt.Errorf("failed to decode secret: %v", err)
		}
		if len(secret) != secretSize {
			return true, fmt.Errorf("The secret should be %d bytes instead of %d", secretSize, len(secret))
		}
		cmd = &redeemCmd{ReceiverKeyPair: receiverFullKeypair, holdingAccountAddress: args[2], secret: secret}

	case "extractsecret":

		_, err = keypair.Parse(args[1])
		if err != nil {
			return true, fmt.Errorf("invalid holding account address: %v", err)
		}
		cmd = &extractSecretCmd{holdingAccountAdress: args[1], secretHash: args[2]}
	}
	err = cmd.runCommand(client)
	return false, err
}

func sha256Hash(x []byte) []byte {
	h := sha256.Sum256(x)
	return h[:]
}
func createRefundTransaction(holdingAccountAddress string, refundAccountAdress string, locktime time.Time, client horizonclient.ClientInterface) (refundTransaction *txnbuild.Transaction, err error) {
	holdingAccount, err := stellar.GetAccount(holdingAccountAddress, client)
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
		Operations:    operations,
		SourceAccount: holdingAccount,
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

//createHoldingAccountTransaction creates a new account to hold the atomic swap balance
//with the signers modified to the atomic swap rules:
//- signature of the destinee and the secret
//- hash of a specific transaction that is present on the chain
//    that merges the escrow account to the account that needs to withdraw
//    and that can only be published in the future ( timeout mechanism)

// createHoldingAccount creates a new account to hold the atomic swap balance
func createHoldingAccount(holdingAccountAddress string, amount string, fundingKeyPair *keypair.Full, network string, asset txnbuild.Asset, client horizonclient.ClientInterface) (err error) {
	fundingAccount, err := stellar.GetAccount(fundingKeyPair.Address(), client)
	if err != nil {
		return
	}
	createAccountTransaction, err := stellar.CreateAccountTransaction(holdingAccountAddress, amount, fundingAccount, network)
	if err != nil {
		return fmt.Errorf("Failed to create the holding account transaction: %s", err)
	}
	tx, err := createAccountTransaction.Sign(network, fundingKeyPair)
	if err != nil {
		return fmt.Errorf("Failed to sign the holding account transaction: %s", err)
	}
	_, err = stellar.SubmitTransaction(tx, client)
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
	secretHashAddress, err := stellar.CreateHashxAddress(secretHash)
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
	refundTxHashAdddress, err := stellar.CreateHashTxAddress(refundTxHash)
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
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewInfiniteTimeout(), //TODO: Use a real timeout
		},
	}

	setOptionsTransaction, err = txnbuild.NewTransaction(setOptionsTransactionParams)

	return
}
func setHoldingAccountSigningOptions(holdingAccountKeyPair *keypair.Full, counterPartyAddress string, secretHash []byte, refundTxHash []byte, network string, client horizonclient.ClientInterface) (err error) {

	holdingAccountAddress := holdingAccountKeyPair.Address()
	holdingAccount, err := stellar.GetAccount(holdingAccountAddress, client)
	if err != nil {
		return
	}
	setSigningOptionsTransaction, err := createHoldingAccountSigningTransaction(holdingAccount, counterPartyAddress, secretHash, refundTxHash, targetNetwork)
	if err != nil {
		return fmt.Errorf("Failed to create the signing options transaction: %s", err)
	}
	tx, err := setSigningOptionsTransaction.Sign(network, holdingAccountKeyPair)
	if err != nil {
		return fmt.Errorf("Failed to sign the signing options transaction: %s", err)
	}
	_, err = stellar.SubmitTransaction(tx, client)
	if err != nil {
		return fmt.Errorf("Failed to publish the signing options transaction : %s", err)
	}
	return
}
func fundHoldingAccount(fundingKeyPair *keypair.Full, holdingAccountKeyPair *keypair.Full, amount string, asset txnbuild.Asset, client horizonclient.ClientInterface) (err error) {
	holdingAccount, err := stellar.GetAccount(holdingAccountKeyPair.Address(), client)
	if err != nil {
		return
	}

	changetrust := txnbuild.ChangeTrust{
		Line:          txnbuild.ChangeTrustAssetWrapper{Asset: txnbuild.CreditAsset{Code: asset.GetCode(), Issuer: asset.GetIssuer()}},
		Limit:         amount,
		SourceAccount: holdingAccount.GetAccountID(),
	}
	fundingAccount, err := stellar.GetAccount(fundingKeyPair.Address(), client)
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
		SourceAccount: fundingAccount,
		Operations:    []txnbuild.Operation{&changetrust, &payment},
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewInfiniteTimeout(), // Use a real timeout in production!
		},
	}
	tx, err := txnbuild.NewTransaction(txParams)
	if err != nil {
		err = errors.Wrap(err, "failed to build funding transaction")
		return
	}
	txe, err := tx.Sign(targetNetwork, holdingAccountKeyPair, fundingKeyPair)
	if err != nil {
		err = fmt.Errorf("Failed to build,sign and encode the funding transaction: %v", err)
		return
	}
	_, err = stellar.SubmitTransaction(txe, client)
	if err != nil {
		transactionID, _ := tx.HashHex(targetNetwork)
		err = fmt.Errorf("Failed to publish the funding transaction : %s\n%s", transactionID, err)
		return
	}
	return
}
func createAtomicSwapHoldingAccount(fundingKeyPair *keypair.Full, holdingAccountKeyPair *keypair.Full, counterPartyAddress string, amount string, secretHash []byte, locktime time.Time, asset txnbuild.Asset, client horizonclient.ClientInterface) (refundTransaction *txnbuild.Transaction, err error) {

	holdingAccountAddress := holdingAccountKeyPair.Address()

	xlmAmount := "10"
	if asset.IsNative() {
		xlmAmount = amount
	}
	err = createHoldingAccount(holdingAccountAddress, xlmAmount, fundingKeyPair, targetNetwork, asset, client)
	if err != nil {
		return
	}

	if !asset.IsNative() {
		err = fundHoldingAccount(fundingKeyPair, holdingAccountKeyPair, amount, asset, client)
		if err != nil {
			return
		}
	}

	refundTransaction, err = createRefundTransaction(holdingAccountAddress, fundingKeyPair.Address(), locktime, client)
	if err != nil {
		return
	}
	refundTransactionHash, err := refundTransaction.Hash(targetNetwork)
	if err != nil {
		err = fmt.Errorf("Failed to Hash the refund transaction: %s", err)
		return
	}
	err = setHoldingAccountSigningOptions(holdingAccountKeyPair, counterPartyAddress, secretHash, refundTransactionHash[:], targetNetwork, client)

	return
}
func (cmd *initiateCmd) runCommand(client horizonclient.ClientInterface) error {
	var secret [secretSize]byte
	_, err := rand.Read(secret[:])
	if err != nil {
		return err
	}
	secretHash := sha256Hash(secret[:])
	fundingAccountAddress := cmd.InitiatorKeyPair.Address()
	holdingAccountKeyPair, err := stellar.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("Failed to create holding account keypair: %s", err)
	}
	holdingAccountAddress := holdingAccountKeyPair.Address()
	//TODO: print the holding account private key in case of an error further down this function
	//to recover the funds

	locktime := time.Now().Add(timings.LockTime)
	refundTransaction, err := createAtomicSwapHoldingAccount(cmd.InitiatorKeyPair, holdingAccountKeyPair, cmd.cp2Addr, cmd.amount, secretHash, locktime, cmd.asset, client)
	if err != nil {
		return err
	}

	serializedRefundTx, err := refundTransaction.Base64()
	if err != nil {
		return err
	}
	if !*automatedFlag {
		fmt.Printf("Secret:      %x\n", secret)
		fmt.Printf("Secret hash: %x\n\n", secretHash)
		fmt.Printf("initiator address: %s\n", fundingAccountAddress)
		fmt.Printf("holding account address: %s\n", holdingAccountAddress)
		fmt.Printf("refund transaction:\n%s\n", serializedRefundTx)
	} else {
		output := struct {
			Secret                string `json:"secret"`
			SecretHash            string `json:"hash"`
			InitiatorAddress      string `json:"initiator"`
			HoldingAccountAddress string `json:"holdingaccount"`
			RefundTransaction     string `json:"refundtransaction"`
		}{fmt.Sprintf("%x", secret),
			fmt.Sprintf("%x", secretHash),
			fundingAccountAddress,
			holdingAccountAddress,
			serializedRefundTx,
		}
		jsonoutput, _ := json.Marshal(output)
		fmt.Println(string(jsonoutput))
	}
	return nil
}

func (cmd *participateCmd) runCommand(client horizonclient.ClientInterface) error {

	fundingAccountAddress := cmd.participatorKeyPair.Address()
	holdingAccountKeyPair, err := stellar.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("Failed to create holding account keypair: %s", err)
	}
	holdingAccountAddress := holdingAccountKeyPair.Address()
	//TODO: print the holding account private key in case of an error further down this function
	//to recover the funds

	locktime := time.Now().Add(timings.LockTime / 2)
	refundTransaction, err := createAtomicSwapHoldingAccount(cmd.participatorKeyPair, holdingAccountKeyPair, cmd.cp1Addr, cmd.amount, cmd.secretHash, locktime, cmd.asset, client)
	if err != nil {
		return err
	}

	serializedRefundTx, err := refundTransaction.Base64()
	if err != nil {
		return err
	}
	if !*automatedFlag {
		fmt.Printf("participant address: %s\n", fundingAccountAddress)
		fmt.Printf("holding account address: %s\n", holdingAccountAddress)
		fmt.Printf("refund transaction:\n%s\n", serializedRefundTx)
	} else {

		output := struct {
			InitiatorAddress      string `json:"partcipant"`
			HoldingAccountAddress string `json:"holdingaccount"`
			RefundTransaction     string `json:"refundtransaction"`
		}{
			fundingAccountAddress,
			holdingAccountAddress,
			serializedRefundTx,
		}
		jsonoutput, _ := json.Marshal(output)
		fmt.Println(string(jsonoutput))
	}
	return nil
}

func (cmd *auditContractCmd) runCommand(client horizonclient.ClientInterface) error {
	holdingAccount, err := client.AccountDetail(horizonclient.AccountRequest{AccountID: cmd.holdingAccountAdress})
	if err != nil {
		return fmt.Errorf("Error getting the holding account details: %v", err)
	}
	if err != nil {
		return err
	}
	//Check if the signing tresholds are correct
	if holdingAccount.Thresholds.HighThreshold != 2 || holdingAccount.Thresholds.MedThreshold != 2 || holdingAccount.Thresholds.LowThreshold != 2 {
		return fmt.Errorf("Holding account signing tresholds are wrong.\nTresholds: High: %d, Medium: %d, Low: %d", holdingAccount.Thresholds.HighThreshold, holdingAccount.Thresholds.MedThreshold, holdingAccount.Thresholds.LowThreshold)
	}
	//Get the signing conditions
	var refundTxHashFromSigningConditions []byte
	recipientAddress := ""
	var secretHash []byte
	for _, signer := range holdingAccount.Signers {
		if signer.Weight == 0 { //The original keypair's signing weight is set to 0
			continue
		}
		switch signer.Type {
		case hprotocol.KeyTypeNames[strkey.VersionByteAccountID]:
			if recipientAddress != "" {
				return fmt.Errorf("Multiple recipients as signer: %s and %s", recipientAddress, signer.Key)
			}
			recipientAddress = signer.Key
			if signer.Weight != 1 {
				return fmt.Errorf("Signing weight of the recipient is wrong. Recipient: %s Weight: %d", signer.Key, signer.Weight)
			}
		case hprotocol.KeyTypeNames[strkey.VersionByteHashTx]:
			if refundTxHashFromSigningConditions != nil {
				return errors.New("Multiple refund transaction hashes as signer")
			}

			refundTxHashFromSigningConditions, err = strkey.Decode(strkey.VersionByteHashTx, signer.Key)
			if err != nil {
				return fmt.Errorf("Faulty encoded refund transaction hash: %s", err)
			}
			if signer.Weight != 2 {
				return fmt.Errorf("Signing weight of the refund transaction is wrong. Weight: %d", signer.Weight)
			}

		case hprotocol.KeyTypeNames[strkey.VersionByteHashX]:
			if secretHash != nil {
				return fmt.Errorf("Multiple secret hashes  transaction hashes as signer: %s and %s", secretHash, signer.Key)
			}
			secretHash, err = strkey.Decode(strkey.VersionByteHashX, signer.Key)
			if err != nil {
				return fmt.Errorf("Faulty encoded secret hash: %s", err)
			}
			if signer.Weight != 1 {
				return fmt.Errorf("Signing weight of the secret hash is wrong. Weight: %d", signer.Weight)
			}
		default:
			return fmt.Errorf("Unexpected signer type: %s", signer.Type)
		}
	}
	//Make sure all signing conditions are present
	if refundTxHashFromSigningConditions == nil {
		return errors.New("Missing refund transaction hash as signer")
	}
	if secretHash == nil {
		return errors.New("Missing secret as signer")
	}
	if recipientAddress == "" {
		return errors.New("Missing recipient as signer")
	}
	//Compare the refund transaction hash in the signing condition to the one of the passed refund transaction
	//cmd.refundTx.Network = targetNetwork TODO: Still required with new library or implied?
	refundTxHash, err := cmd.refundTx.Hash(targetNetwork)
	if err != nil {
		return fmt.Errorf("Unable to hash the passed refund transaction: %v", err)
	}
	if !bytes.Equal(refundTxHashFromSigningConditions, refundTxHash[:]) {
		return errors.New("Refund transaction hash in the signing condition is not equal to the one of the passed refund transaction")
	}
	//and finally get the locktime and refund address
	lockTime := cmd.refundTx.Timebounds().MinTime
	if len(cmd.refundTx.Operations()) != 1 {
		return fmt.Errorf("Refund transaction is expected to have 1 operation instead of %d", len(cmd.refundTx.Operations()))
	}
	refundoperation := cmd.refundTx.Operations()[0]
	accountMergeOperation, ok := cmd.refundTx.Operations()[0].(*txnbuild.AccountMerge)
	if !ok {
		return fmt.Errorf("Expecting an accountmerge operation in the refund transaction but got a %v", reflect.TypeOf(refundoperation))
	}
	if accountMergeOperation.SourceAccount != cmd.holdingAccountAdress {
		return fmt.Errorf("The refund transaction does not refund from the holding account but from %v", accountMergeOperation.SourceAccount)
	}
	refundAddress := accountMergeOperation.Destination
	if !*automatedFlag {
		fmt.Printf("Contract address:        %v\n", cmd.holdingAccountAdress)
		fmt.Println("Contract value:")
		for _, balance := range holdingAccount.Balances {
			if balance.Asset.Type == stellar.NativeAssetType {
				fmt.Printf("Amount: %s XLM\n", balance.Balance)
			} else {
				fmt.Printf("Amount: %s Code: %s Issuer: %s\n", balance.Balance, balance.Code, balance.Issuer)
			}
		}
		fmt.Printf("Recipient address:       %v\n", recipientAddress)
		fmt.Printf("Refund address: %v\n\n", refundAddress)

		fmt.Printf("Secret hash: %x\n\n", secretHash)

		t := time.Unix(lockTime, 0)
		fmt.Printf("Locktime: %v\n", t.UTC())
		reachedAt := time.Until(t).Truncate(time.Second)
		if reachedAt > 0 {
			fmt.Printf("Locktime reached in %v\n", reachedAt)
		} else {
			fmt.Printf("Refund time lock has expired\n")
		}
	} else {
		output := struct {
			ContractAddress  string `json:"contractAddress"`
			ContractValue    string `json:"contractValue"`
			RecipientAddress string `json:"recipientAddress"`
			RefundAddress    string `json:"refundAddress"`
			SecretHash       string `json:"secretHash"`
			Locktime         string `json:"Locktime"`
		}{
			fmt.Sprintf("%v", cmd.holdingAccountAdress),
			"", //TODO: json output for balances
			recipientAddress,
			refundAddress,
			fmt.Sprintf("%x", secretHash),
			"",
		}
		t := time.Unix(lockTime, 0)
		output.Locktime = fmt.Sprintf("%v", t.UTC())
		jsonoutput, _ := json.Marshal(output)
		fmt.Println(string(jsonoutput))
	}
	return nil
}

func (cmd *refundCmd) runCommand(client horizonclient.ClientInterface) error {
	txe := cmd.refundTx
	result, err := stellar.SubmitTransaction(&txe, client)
	if err != nil {
		return err
	}
	if !*automatedFlag {
		fmt.Println(result.ID) // FIXME: this was "result.TransactionSuccessToString()"
	}
	return nil
}

func createRedeemOperations(holdingAccount *horizon.Account, receiverAddress string) (redeemOperations []txnbuild.Operation) {
	redeemOperations = make([]txnbuild.Operation, 0, len(holdingAccount.Balances))
	for _, balance := range holdingAccount.Balances {
		if balance.Asset.Type == stellar.NativeAssetType {
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

func (cmd *redeemCmd) runCommand(client horizonclient.ClientInterface) error {
	holdingAccount, err := stellar.GetAccount(cmd.holdingAccountAddress, client)
	if err != nil {
		return err
	}
	receiverAddress := cmd.ReceiverKeyPair.Address()
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
		return fmt.Errorf("Unable to build the transaction: %v", err)
	}
	redeemTransaction, err = redeemTransaction.SignHashX(cmd.secret)
	if err != nil {
		return fmt.Errorf("Unable to sign with the secret:%v", err)
	}
	redeemTransaction, err = redeemTransaction.Sign(targetNetwork, cmd.ReceiverKeyPair)
	if err != nil {
		return fmt.Errorf("Unable to sign with the receiver keypair:%v", err)
	}

	txSuccess, err := stellar.SubmitTransaction(redeemTransaction, client)
	if err != nil {
		return err
	}

	if !*automatedFlag {
		fmt.Println(txSuccess.ID) // FIXME: this was txSuccess.TransactionSuccessToString()
	} else {
		output := struct {
			RedeemTransactionTxHash string `json:"redeemTransaction"`
		}{
			fmt.Sprintf("%v", txSuccess.Hash),
		}
		jsonoutput, _ := json.Marshal(output)
		fmt.Println(string(jsonoutput))
	}
	return nil
}

func (cmd *extractSecretCmd) runCommand(client horizonclient.ClientInterface) error {
	transactions, err := stellar.GetAccountDebitediTransactions(cmd.holdingAccountAdress, client)
	if err != nil {
		return fmt.Errorf("Error getting the transaction that debited the holdingAccount: %v", err)
	}
	if len(transactions) == 0 {
		return errors.New("The holdingaccount has not been redeemed yet")
	}
	var extractedSecret []byte
transactionsLoop:
	for _, transaction := range transactions {

		for _, rawSignature := range transaction.Signatures {

			decodedSignature, err := base64.StdEncoding.DecodeString(rawSignature)
			if err != nil {
				return fmt.Errorf("Error base64 decoding signature :%v", err)
			}
			if len(decodedSignature) > xdr.Signature(decodedSignature).XDRMaxSize() {
				continue // this is certainly not the secret we are looking for
			}
			signatureHash := sha256.Sum256(decodedSignature)
			hexSignatureHash := fmt.Sprintf("%x", signatureHash)
			if hexSignatureHash == cmd.secretHash {
				extractedSecret = decodedSignature
				break transactionsLoop
			}
		}
	}

	if extractedSecret == nil {
		return errors.New("Unable to find the matching secret")
	}
	fmt.Printf("Extracted secret: %x\n", extractedSecret)
	return nil
}
