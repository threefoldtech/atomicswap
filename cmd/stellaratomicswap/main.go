package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/stellar/go/txnbuild"

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
		fmt.Println("  participate <initiator address> <amount> <secret hash>")
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
	cp1Addr    string
	amount     string
	secretHash []byte
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
		cmdArgs = 3
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
	}
	err = cmd.runCommand(client)
	return false, err
}

func sha256Hash(x []byte) []byte {
	h := sha256.Sum256(x)
	return h[:]
}
func createRefundTransaction(holdingAccount txnbuild.Account, refundAccountAdress string) (refundTransaction txnbuild.Transaction) {

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
	return
}

func (cmd *initiateCmd) runCommand(client horizonclient.ClientInterface) error {
	var secret [secretSize]byte
	_, err := rand.Read(secret[:])
	if err != nil {
		return err
	}
	secretHash := sha256Hash(secret[:])

	initiatoraccount, err := stellar.GetAccount(cmd.InitiatorKeyPair.Address(), client)
	if err != nil {
		return err
	}
	holdingAccountKeyPair, err := stellar.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("Failed to create holding account keypair: %s", err)
	}
	holdingAccount := initiatoraccount //Just for testing
	_, err = holdingAccount.IncrementSequenceNumber()
	if err != nil {
		return err
	}
	initiatoraccount, err = stellar.GetAccount(cmd.InitiatorKeyPair.Address(), client)
	if err != nil {
		return err
	}
	refundTransaction := createRefundTransaction(holdingAccount, initiatoraccount.GetAccountID())
	if err != nil {
		return fmt.Errorf("Failed to create the refund transaction: %s", err)
	}
	if err = refundTransaction.Build(); err != nil {
		return fmt.Errorf("Failed to build the refund transaction: %s", err)
	}

	refundTransactionHash, err := refundTransaction.Hash()
	if err != nil {
		return fmt.Errorf("Failed to Hash the refund transaction: %s", err)
	}
	createAccountTransaction, err := stellar.CreateHoldingAccount(holdingAccountKeyPair.Address(), cmd.amount, initiatoraccount, cmd.cp2Addr, secretHash, refundTransactionHash[:], targetNetwork)
	if err != nil {
		return fmt.Errorf("Failed to create the holding account transaction: %s", err)
	}
	txe, err := createAccountTransaction.BuildSignEncode(cmd.InitiatorKeyPair)
	if err != nil {
		return fmt.Errorf("Failed to sign the holding account transaction: %s", err)
	}
	txSuccess, err := stellar.SubmitTransaction(txe, client)
	if err != nil {
		return fmt.Errorf("Failed to publish the holding account creation transaction : %s", err)
	}
	serializedRefundTx, err := refundTransaction.Base64()
	if err != nil {
		return err
	}
	if !*automatedFlag {
		fmt.Println(txSuccess.TransactionSuccessToString())
		fmt.Printf("Secret:      %x\n", secret)
		fmt.Printf("Secret hash: %x\n\n", secretHash)
		fmt.Printf("initiator address: %s\n", initiatoraccount.GetAccountID())
		fmt.Printf("holding account address: %s\n", holdingAccountKeyPair.Address())
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
			initiatoraccount.GetAccountID(),
			holdingAccountKeyPair.Address(),
			serializedRefundTx,
		}
		jsonoutput, _ := json.Marshal(output)
		fmt.Println(string(jsonoutput))
	}
	return errors.New("Not implemented yet")
}
