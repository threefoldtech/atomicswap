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
		fmt.Println("  initiate <participant address> <amount>")
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
	runCommand() error
}

// offline commands don't require wallet RPC.
type offlineCommand interface {
	command
	runOfflineCommand() error
}

type initiateCmd struct {
	cp2Addr string
	amount  string
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
		cmdArgs = 2
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

	var cmd command
	switch args[0] {
	case "initiate":
		_, err := keypair.Parse(args[1])
		if err != nil {
			return true, fmt.Errorf("invalid participant address: %v", err)
		}

		_, err = strconv.ParseFloat(args[2], 64)
		if err != nil {
			return true, fmt.Errorf("failed to decode amount: %v", err)
		}

		cmd = &initiateCmd{cp2Addr: args[1], amount: args[2]}
	}
	err = cmd.runCommand()
	return false, err
}

func sha256Hash(x []byte) []byte {
	h := sha256.Sum256(x)
	return h[:]
}
func (cmd *initiateCmd) runCommand() error {
	var secret [secretSize]byte
	_, err := rand.Read(secret[:])
	if err != nil {
		return err
	}
	secretHash := sha256Hash(secret[:])
	if !*automatedFlag {
		fmt.Printf("Secret:      %x\n", secret)
		fmt.Printf("Secret hash: %x\n\n", secretHash)
	} else {
		output := struct {
			Secret     string `json:"secret"`
			SecretHash string `json:"hash"`
		}{fmt.Sprintf("%x", secret),
			fmt.Sprintf("%x", secretHash),
		}
		jsonoutput, _ := json.Marshal(output)
		fmt.Println(string(jsonoutput))
	}
	return errors.New("Not implemented yet")
}
