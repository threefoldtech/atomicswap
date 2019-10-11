package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/stellar/go/txnbuild"

	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/threefoldtech/atomicswap/cmd/stellaratomicswap/stellar"

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
		fmt.Println("  initiate <initiator seed> <participant address> <amount>")
		fmt.Println("  participate <participant seed> <initiator address> <amount> <secret hash>")
		fmt.Println("  redeem <contract> <contract transaction> <secret>")
		fmt.Println("  refund <contract> <contract transaction>")
		fmt.Println("  extractsecret <redemption transaction> <secret hash>")
		fmt.Println("  auditcontract <contract> <contract transaction>")
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
}

type participateCmd struct {
	cp1Addr             string
	participatorKeyPair *keypair.Full
	amount              string
	secretHash          []byte
}

type redeemCmd struct {
	contract   []byte
	contractTx string
	secret     []byte
}

type refundCmd struct {
	contract   []byte
	contractTx string
}

type extractSecretCmd struct {
	redemptionTx string
	secretHash   []byte
}

type auditContractCmd struct {
	contract   []byte
	contractTx string
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
		cmdArgs = 2
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

		cmd = &initiateCmd{InitiatorKeyPair: initiatorFullKeypair, cp2Addr: args[2], amount: args[3]}
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
		cmd = &participateCmd{participatorKeyPair: participatorFullKeypair, cp1Addr: args[2], amount: args[3], secretHash: secretHash}
	}
	err = cmd.runCommand(client)
	return false, err
}

func sha256Hash(x []byte) []byte {
	h := sha256.Sum256(x)
	return h[:]
}
func createRefundTransaction(holdingAccountAddress string, refundAccountAdress string, client horizonclient.ClientInterface) (refundTransaction txnbuild.Transaction, err error) {
	holdingAccount, err := stellar.GetAccount(holdingAccountAddress, client)
	if err != nil {
		return
	}
	_, err = holdingAccount.IncrementSequenceNumber()
	if err != nil {
		return
	}

	mergeAccountOperation := txnbuild.AccountMerge{
		Destination:   refundAccountAdress,
		SourceAccount: holdingAccount,
	}
	refundTransaction = txnbuild.Transaction{
		Timebounds: txnbuild.NewTimebounds(int64(0), int64(0)),
		Operations: []txnbuild.Operation{
			&mergeAccountOperation,
		},
		Network:       targetNetwork,
		SourceAccount: holdingAccount,
	}

	if err = refundTransaction.Build(); err != nil {
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
func createHoldingAccountTransaction(holdingAccountAddress string, xlmAmount string, fundingAccount *horizon.Account, network string) (createAccountTransaction txnbuild.Transaction, err error) {

	accountCreationOperation := txnbuild.CreateAccount{
		Destination:   holdingAccountAddress,
		Amount:        xlmAmount,
		SourceAccount: fundingAccount,
	}

	createAccountTransaction = txnbuild.Transaction{
		SourceAccount: fundingAccount,
		Operations: []txnbuild.Operation{
			&accountCreationOperation,
		},
		Network:    network,
		Timebounds: txnbuild.NewInfiniteTimeout(), //TODO: Use a real timeout
	}

	return
}

//createHoldingAccount creates a new account to hold the atomic swap balance
func createHoldingAccount(holdingAccountAddress string, xlmAmount string, fundingKeyPair *keypair.Full, network string, client horizonclient.ClientInterface) (err error) {

	fundingAccount, err := stellar.GetAccount(fundingKeyPair.Address(), client)
	if err != nil {
		return fmt.Errorf("Failed to get the funding account:%s", err)
	}
	createAccountTransaction, err := createHoldingAccountTransaction(holdingAccountAddress, xlmAmount, fundingAccount, network)
	if err != nil {
		return fmt.Errorf("Failed to create the holding account transaction: %s", err)
	}
	txe, err := createAccountTransaction.BuildSignEncode(fundingKeyPair)
	if err != nil {
		return fmt.Errorf("Failed to sign the holding account transaction: %s", err)
	}
	_, err = stellar.SubmitTransaction(txe, client)
	if err != nil {
		return fmt.Errorf("Failed to publish the holding account creation transaction : %s", err)
	}
	return
}
func createHoldingAccountSigningTransaction(holdingAccount *horizon.Account, counterPartyAddress string, secretHash []byte, refundTxHash []byte, network string) (setOptionsTransaction txnbuild.Transaction, err error) {

	depositorSigningOperation := txnbuild.SetOptions{
		Signer: &txnbuild.Signer{
			Address: counterPartyAddress,
			Weight:  1,
		},
		SourceAccount: holdingAccount,
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
		SourceAccount: holdingAccount,
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
		SourceAccount: holdingAccount,
	}
	setSigingWeightsOperation := txnbuild.SetOptions{
		MasterWeight:    txnbuild.NewThreshold(txnbuild.Threshold(uint8(0))),
		LowThreshold:    txnbuild.NewThreshold(txnbuild.Threshold(2)),
		MediumThreshold: txnbuild.NewThreshold(txnbuild.Threshold(2)),
		HighThreshold:   txnbuild.NewThreshold(txnbuild.Threshold(2)),
		SourceAccount:   holdingAccount,
	}
	setOptionsTransaction = txnbuild.Transaction{
		SourceAccount: holdingAccount, //TODO: check if this can be changed to the fundingaccount
		Operations: []txnbuild.Operation{
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
func setHoldingAccountSigningOptions(holdingAccountKeyPair *keypair.Full, counterPartyAddress string, secretHash []byte, refundTxHash []byte, network string, client horizonclient.ClientInterface) (err error) {

	holdingAccountAddress := holdingAccountKeyPair.Address()
	holdingAccount, err := stellar.GetAccount(holdingAccountAddress, client)
	if err != nil {
		return fmt.Errorf("Failed to get the holding account: %s", err)
	}
	setSigningOptionsTransaction, err := createHoldingAccountSigningTransaction(holdingAccount, counterPartyAddress, secretHash, refundTxHash, targetNetwork)
	if err != nil {
		return fmt.Errorf("Failed to create the signing options transaction: %s", err)
	}
	txe, err := setSigningOptionsTransaction.BuildSignEncode(holdingAccountKeyPair)
	if err != nil {
		return fmt.Errorf("Failed to sign the signing options transaction: %s", err)
	}
	_, err = stellar.SubmitTransaction(txe, client)
	if err != nil {
		return fmt.Errorf("Failed to publish the signing options transaction : %s", err)
	}
	return
}
func createAtomicSwapHoldingAccount(fundingKeyPair *keypair.Full, holdingAccountKeyPair *keypair.Full, counterPartyAddress string, xlmAmount string, secretHash []byte, client horizonclient.ClientInterface) (refundTransaction txnbuild.Transaction, err error) {

	holdingAccountAddress := holdingAccountKeyPair.Address()
	err = createHoldingAccount(holdingAccountAddress, xlmAmount, fundingKeyPair, targetNetwork, client)
	if err != nil {
		return
	}

	refundTransaction, err = createRefundTransaction(holdingAccountAddress, fundingKeyPair.Address(), client)
	if err != nil {
		return
	}
	refundTransactionHash, err := refundTransaction.Hash()
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

	refundTransaction, err := createAtomicSwapHoldingAccount(cmd.InitiatorKeyPair, holdingAccountKeyPair, cmd.cp2Addr, cmd.amount, secretHash, client)
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

	refundTransaction, err := createAtomicSwapHoldingAccount(cmd.participatorKeyPair, holdingAccountKeyPair, cmd.cp1Addr, cmd.amount, cmd.secretHash, client)
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
