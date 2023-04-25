package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/txnbuild"
	"github.com/threefoldtech/atomicswap/stellar"

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
	asset                txnbuild.Asset
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
		cmd = &auditContractCmd{holdingAccountAdress: args[1], refundTx: *refundTransaction, asset: asset}
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

func (cmd *initiateCmd) runCommand(client horizonclient.ClientInterface) error {
	output, err := stellar.Initiate(targetNetwork, cmd.InitiatorKeyPair, cmd.cp2Addr, cmd.amount, cmd.asset, client)
	if err != nil {
		return err
	}

	if !*automatedFlag {
		fmt.Printf("Secret:      %x\n", output.Secret)
		fmt.Printf("Secret hash: %x\n\n", output.SecretHash)
		fmt.Printf("initiator address: %s\n", output.InitiatorAddress)
		fmt.Printf("holding account address: %s\n", output.HoldingAccountAddress)
		fmt.Printf("refund transaction:\n%s\n", output.RefundTransaction)
	} else {
		jsonoutput, _ := json.Marshal(output)
		fmt.Println(string(jsonoutput))
	}
	return nil
}

func (cmd *participateCmd) runCommand(client horizonclient.ClientInterface) error {
	output, err := stellar.Participate(targetNetwork, cmd.participatorKeyPair, cmd.cp1Addr, cmd.amount, cmd.secretHash, cmd.asset, client)
	if err != nil {
		return err
	}
	if !*automatedFlag {
		fmt.Printf("participant address: %s\n", output.InitiatorAddress)
		fmt.Printf("holding account address: %s\n", output.HoldingAccountAddress)
		fmt.Printf("refund transaction:\n%s\n", output.RefundTransaction)
	} else {
		jsonoutput, _ := json.Marshal(output)
		fmt.Println(string(jsonoutput))
	}
	return nil
}

func (cmd *auditContractCmd) runCommand(client horizonclient.ClientInterface) error {
	output, err := stellar.AuditContract(targetNetwork, cmd.refundTx, cmd.holdingAccountAdress, cmd.asset, client)
	if err != nil {
		return err
	}
	if !*automatedFlag {
		fmt.Printf("Contract address:        %v\n", cmd.holdingAccountAdress)
		fmt.Println("Contract value:")
		fmt.Printf("Amount: %s Code: %s Issuer: %s\n", output.ContractValue, cmd.asset.GetCode(), cmd.asset.GetIssuer())
		fmt.Printf("Recipient address:       %v\n", output.RecipientAddress)
		fmt.Printf("Refund address: %v\n\n", output.RefundAddress)

		fmt.Printf("Secret hash: %x\n\n", output.SecretHash)

		t := time.Unix(output.Locktime, 0)
		fmt.Printf("Locktime: %v\n", t.UTC())
		reachedAt := time.Until(t).Truncate(time.Second)
		if reachedAt > 0 {
			fmt.Printf("Locktime reached in %v\n", reachedAt)
		} else {
			fmt.Printf("Refund time lock has expired\n")
		}
	} else {
		jsonoutput, _ := json.Marshal(output)
		fmt.Println(string(jsonoutput))
	}
	return nil
}

func (cmd *refundCmd) runCommand(client horizonclient.ClientInterface) error {
	result, err := stellar.Refund(targetNetwork, cmd.refundTx, client)
	if err != nil {
		return err
	}
	if !*automatedFlag {
		fmt.Println(result)
	}
	return nil
}

func (cmd *redeemCmd) runCommand(client horizonclient.ClientInterface) error {
	output, err := stellar.Redeem(targetNetwork, cmd.ReceiverKeyPair, cmd.holdingAccountAddress, cmd.secret, client)
	if err != nil {
		return err
	}

	if !*automatedFlag {
		fmt.Println(output.RedeemTransactionTxHash) // FIXME: this was txSuccess.TransactionSuccessToString()
	} else {
		jsonoutput, _ := json.Marshal(output)
		fmt.Println(string(jsonoutput))
	}
	return nil
}

func (cmd *extractSecretCmd) runCommand(client horizonclient.ClientInterface) error {
	secret, err := stellar.ExtractSecret(targetNetwork, cmd.holdingAccountAdress, cmd.secretHash, client)
	if err != nil {
		return err
	}

	fmt.Printf("Extracted secret: %x\n", secret)
	return nil
}
