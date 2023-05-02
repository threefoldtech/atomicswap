// Copyright (c) 2018 The Decred developers and Contributors
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/bgentry/speakeasy"
	"github.com/pkg/errors"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/threefoldtech/atomicswap/eth"
	"github.com/threefoldtech/atomicswap/eth/contract"
)

var (
	chainConfig = params.MainnetChainConfig
)

const (
	initiateLockPeriodInSeconds    = 48 * 60 * 60
	participateLockPeriodInSeconds = 24 * 60 * 60

	maxGasLimit = 210000
)

var (
	flagset      = flag.NewFlagSet("", flag.ExitOnError)
	connectFlag  = flagset.String("s", "http://localhost:8545", "endpoint of Ethereum RPC server")
	contractFlag = flagset.String("c", "", "hex-enoded address of the deployed contract")
	accountFlag  = flagset.String("account", "", "account file, account address or nothing for the daemon's first account")
	timeoutFlag  = flagset.Duration("t", 0, "optional timeout of any call made")
	testnetFlag  = flagset.Bool("testnet", false, "use testnet (Rinkeby) network")
)

// There are two directions that the atomic swap can be performed, as the
// initiator can be on either chain.  This tool only deals with creating the
// Bitcoin transactions for these swaps.  A second tool should be used for the
// transaction on the other chain.  Any chain can be used so long as it supports
// OP_SHA256 and OP_CHECKLOCKTIMEVERIFY.
//
// Example scenerios using bitcoin as the second chain:
//
// Scenerio 1:
//   cp1 initiates (dcr)
//   cp2 participates with cp1 H(S) (eth)
//   cp1 redeems eth revealing S
//     - must verify H(S) in contract is hash of known secret
//   cp2 redeems dcr with S
//
// Scenerio 2:
//   cp1 initiates (eth)
//   cp2 participates with cp1 H(S) (dcr)
//   cp1 redeems dcr revealing S
//     - must verify H(S) in contract is hash of known secret
//   cp2 redeems eth with S

func init() {
	flagset.Usage = func() {
		fmt.Println("Usage: ethatomicswap [flags] cmd [cmd args]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  initiate <participant address> <amount>")
		fmt.Println("  participate <initiator address> <amount> <secret hash>")
		fmt.Println("  redeem <contract transaction> <secret>")
		fmt.Println("  refund <contract transaction>")
		fmt.Println("  extractsecret <redemption transaction> <secret hash>")
		fmt.Println("  auditcontract <contract transaction>")
		fmt.Println()
		fmt.Println("Extra Commands:")
		fmt.Println("  deploycontract")
		fmt.Println("  validatedeployedcontract <deploy transaction>")
		fmt.Println()
		fmt.Println("Flags:")
		flagset.PrintDefaults()
	}
}

type command interface {
	runCommand(eth.SwapContractTransactor) error
}

// offline commands don't require wallet RPC.
type offlineCommand interface {
	command
	runOfflineCommand() error
}

type initiateCmd struct {
	cp2Addr common.Address
	amount  *big.Int // in wei
}

type participateCmd struct {
	cp1Addr    common.Address
	amount     *big.Int // in wei
	secretHash [32]byte
}

type redeemCmd struct {
	contractTx *types.Transaction
	secret     [32]byte
}

type refundCmd struct {
	contractTx *types.Transaction
}

type extractSecretCmd struct {
	redemptionTx *types.Transaction
	secretHash   [32]byte
}

type auditContractCmd struct {
	contractTx *types.Transaction
}

type deployContractCmd struct{}

type validateDeployedContractCmd struct {
	deployTx *types.Transaction
}

func main() {
	err, showUsage := run()
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

const (
	weiPrecision = 18
)

func parseEthAsWei(str string) (*big.Int, error) {
	initialParts := strings.SplitN(str, ".", 2)
	if len(initialParts) == 1 {
		// a round number, simply multiply and go
		i, ok := big.NewInt(0).SetString(initialParts[0], 10)
		if !ok {
			return nil, errors.New("invalid round amount")
		}
		switch i.Cmp(big.NewInt(0)) {
		case -1:
			return nil, errors.New("invalid round amount: cannot be negative")
		case 0:
			return nil, errors.New("invalid round amount: cannot be nil")
		}
		return i.Mul(i, new(big.Int).Exp(big.NewInt(10), big.NewInt(weiPrecision), nil)), nil
	}

	whole := initialParts[0]
	dac := initialParts[1]
	sn := uint(weiPrecision)
	if l := uint(len(dac)); l < sn {
		sn = l
	}
	whole += initialParts[1][:sn]
	dac = dac[sn:]
	for i := range dac {
		if dac[i] != '0' {
			return nil, errors.New("invalid or too precise amount")
		}
	}
	i, ok := big.NewInt(0).SetString(whole, 10)
	if !ok {
		return nil, errors.New("invalid amount")
	}
	switch i.Cmp(big.NewInt(0)) {
	case -1:
		return nil, errors.New("invalid round amount: cannot be negative")
	case 0:
		return nil, errors.New("invalid round amount: cannot be nil")
	}
	i.Mul(i, big.NewInt(0).Exp(
		big.NewInt(10), big.NewInt(int64(weiPrecision-sn)), nil))

	switch i.Cmp(big.NewInt(0)) {
	case -1:
		return nil, errors.New("invalid round amount: cannot be negative")
	case 0:
		return nil, errors.New("invalid round amount: cannot be nil")
	}
	return i, nil
}

func formatWeiAsEthString(w *big.Int) string {
	if w.Cmp(big.NewInt(0)) == 0 {
		return "0"
	}

	str := w.String()
	l := uint(len(str))
	if l > weiPrecision {
		idx := l - weiPrecision
		str = strings.TrimRight(str[:idx]+"."+str[idx:], "0")
		str = strings.TrimRight(str, ".")
		if len(str) == 0 {
			return "0"
		}
		return str
	}
	str = "0." + strings.Repeat("0", int(weiPrecision-l)) + str
	str = strings.TrimRight(str, "0")
	str = strings.TrimRight(str, ".")
	return str
}

func run() (err error, showUsage bool) {
	flagset.Parse(os.Args[1:])
	args := flagset.Args()
	if len(args) == 0 {
		return nil, true
	}
	cmdArgs := 0
	switch args[0] {
	case "initiate":
		cmdArgs = 2
	case "participate":
		cmdArgs = 3
	case "redeem":
		cmdArgs = 2
	case "refund":
		cmdArgs = 1
	case "extractsecret":
		cmdArgs = 2
	case "auditcontract":
		cmdArgs = 1
	case "deploycontract":
		cmdArgs = 0
	case "validatedeployedcontract":
		cmdArgs = 1
	default:
		return fmt.Errorf("unknown command %v", args[0]), true
	}
	nArgs := checkCmdArgLength(args[1:], cmdArgs)
	flagset.Parse(args[1+nArgs:])
	if nArgs < cmdArgs {
		return fmt.Errorf("%s: too few arguments", args[0]), true
	}
	if flagset.NArg() != 0 {
		return fmt.Errorf("unexpected argument: %s", flagset.Arg(0)), true
	}

	if *testnetFlag {
		chainConfig = params.RinkebyChainConfig
	}

	var cmd command
	switch args[0] {
	case "initiate":
		cp2Addr := common.HexToAddress(args[1])
		amount, err := parseEthAsWei(args[2])
		if err != nil {
			return fmt.Errorf("unexpected amount argument (%v): %v", args[2], err), true
		}
		cmd = &initiateCmd{
			cp2Addr: cp2Addr,
			amount:  amount,
		}

	case "participate":
		cp1Addr := common.HexToAddress(args[1])
		amount, err := parseEthAsWei(args[2])
		if err != nil {
			return fmt.Errorf("unexpected amount argument (%v): %v", args[2], err), true
		}
		secretHash, err := hexDecodeSha256Hash("secret hash", args[3])
		if err != nil {
			return err, true
		}
		cmd = &participateCmd{
			cp1Addr:    cp1Addr,
			amount:     amount,
			secretHash: secretHash,
		}

	case "redeem":
		contractTx, err := hexDecodeTransaction(args[1])
		if err != nil {
			return err, true
		}
		secret, err := hexDecodeSha256Hash("secret", args[2])
		if err != nil {
			return err, true
		}
		cmd = &redeemCmd{
			contractTx: contractTx,
			secret:     secret,
		}

	case "refund":
		contractTx, err := hexDecodeTransaction(args[1])
		if err != nil {
			return err, true
		}
		cmd = &refundCmd{
			contractTx: contractTx,
		}

	case "extractsecret":
		redemptionTx, err := hexDecodeTransaction(args[1])
		if err != nil {
			return err, true
		}
		secretHash, err := hexDecodeSha256Hash("secret hash", args[2])
		if err != nil {
			return err, true
		}
		cmd = &extractSecretCmd{
			redemptionTx: redemptionTx,
			secretHash:   secretHash,
		}

	case "auditcontract":
		contractTx, err := hexDecodeTransaction(args[1])
		if err != nil {
			return err, true
		}
		cmd = &auditContractCmd{
			contractTx: contractTx,
		}

	case "deploycontract":
		cmd = new(deployContractCmd)

	case "validatedeployedcontract":
		deployTx, err := hexDecodeTransaction(args[1])
		if err != nil {
			return err, true
		}
		cmd = &validateDeployedContractCmd{
			deployTx: deployTx,
		}

	default:
		panic(fmt.Sprintf("unknown command %v", args[0]))
	}

	// Offline commands don't need to talk to the wallet.
	if cmd, ok := cmd.(offlineCommand); ok {
		return cmd.runOfflineCommand(), false
	}

	ctx := context.Background()
	client, err := eth.DialClient(ctx, *connectFlag)
	if err != nil {
		return fmt.Errorf("rpc connect: %v", err), false
	}
	defer client.Close()

	// create (swap) contract transactor
	contractAddr, err := getDeployedContractAddress()
	if err != nil {
		return fmt.Errorf("failed to get contract address: %v", err), false
	}
	key, err := loadAccount(*accountFlag)
	if err != nil {
		return errors.Wrap(err, "could not load account key"), false
	}
	sct, err := eth.NewSwapContractTransactor(ctx, client, contractAddr, key)
	if err != nil {
		return err, false
	}

	err = cmd.runCommand(sct)
	return err, false
}

func loadAccount(path string) (*ecdsa.PrivateKey, error) {

	json, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read encrypted account/key file (%s) content: %v", path, err)
	}
	passphrase, err := speakeasy.Ask("Account passphrase: ")
	if err != nil {
		return nil, fmt.Errorf("failed to get passphrase from STDIN: %v", err)
	}
	key, err := keystore.DecryptKey(json, passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt (JSON) account/key file (%s): %v", path, err)
	}
	return key.PrivateKey, nil
}

func getDeployedContractAddress() (common.Address, error) {
	contractAddress := *contractFlag
	if contractAddress != "" {
		return common.HexToAddress(contractAddress), nil
	}
	switch chainConfig {
	case params.MainnetChainConfig:
		return common.Address{}, errors.New("no default contract exist yet for the main net")
	case params.RinkebyChainConfig:
		return common.HexToAddress("2661CBAa149721f7c5FAB3FA88C1EA564A683631"), nil
	}

	panic("unknown chain config for chain ID: " + chainConfig.ChainID.String())
}

func sha256Hash(x []byte) [sha256.Size]byte {
	h := sha256.Sum256(x)
	return h
}

func hexDecodeSha256Hash(name, str string) (hash [sha256.Size]byte, err error) {
	slice, err := hex.DecodeString(strings.TrimPrefix(str, "0x"))
	if err != nil {
		err = errors.New(name + " must be hex encoded")
		return
	}
	if len(slice) != sha256.Size {
		err = errors.New(name + " has wrong size")
		return
	}
	copy(hash[:], slice)
	return
}

func hexDecodeTransaction(str string) (*types.Transaction, error) {
	slice, err := hex.DecodeString(strings.TrimPrefix(str, "0x"))
	if err != nil {
		return nil, errors.New("transaction must be hex encoded")
	}
	var tx types.Transaction
	err = rlp.DecodeBytes(slice, &tx)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction: %v", err)
	}
	return &tx, nil
}

func generateSecretHashPair() (secret, secretHash [sha256.Size]byte) {
	rand.Read(secret[:])
	secretHash = sha256Hash(secret[:])
	return
}

func promptPublishTx(name string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Publish %s transaction? [y/N] ", name)
		answer, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}
		answer = strings.TrimSpace(strings.ToLower(answer))

		switch answer {
		case "y", "yes":
			return true, nil
		case "n", "no", "":
			return false, nil
		default:
			fmt.Println("please answer y or n")
			continue
		}
	}
}

func calcGasCost(limit uint64, c *ethclient.Client) (*big.Int, error) {
	price, err := c.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	return price.Mul(price, big.NewInt(int64(limit))), nil
}

func unpackContractInputParams(abi abi.ABI, tx *types.Transaction) (params struct {
	LockDuration *big.Int
	SecretHash   [sha256.Size]byte
	ToAddress    common.Address
}, err error) {
	txData := tx.Data()

	// first 4 bytes contain the id, so let's get method using that ID
	method, err := abi.MethodById(txData[:4])
	if err != nil {
		err = fmt.Errorf("failed to get method using its parsed id: %v", err)
		return
	}

	args, err := method.Inputs.Unpack(txData[4:])
	if err != nil {
		err = fmt.Errorf("failed to unpack method's input params: %v", err)
	}
	if len(args) != 3 {
		err = errors.New("unexpected amount of transaction params")
		return
	}

	params.LockDuration = args[0].(*big.Int)
	params.SecretHash = args[1].([sha256.Size]byte)
	params.ToAddress = args[2].(common.Address)

	return
}

func (cmd *initiateCmd) runCommand(sct eth.SwapContractTransactor) error {
	output, err := eth.Initiate(context.Background(), sct, cmd.cp2Addr, cmd.amount)
	if err != nil {
		return errors.Wrap(err, "failed to create initiate TX")
	}

	fmt.Printf("Amount: %s Wei (%s ETH)\n\n",
		cmd.amount.String(), formatWeiAsEthString(cmd.amount))

	fmt.Printf("Author's refund address: %x\n\n", sct.FromAddr)

	fmt.Printf("Secret:      %x\n", output.Secret)
	fmt.Printf("Secret hash: %x\n\n", output.SecretHash)

	fmt.Printf("Chain ID:         %s\n", chainConfig.ChainID.String())
	fmt.Printf("Contract Address: %x\n", sct.ContractAddr)

	fmt.Printf("Contract transaction (%x):\n", output.ContractTransaction.Hash())
	txBytes, err := rlp.EncodeToBytes(output.ContractTransaction)
	if err != nil {
		return fmt.Errorf("failed to encode contract TX: %v", err)
	}
	fmt.Printf("%x\n\n", txBytes)

	publish, err := promptPublishTx("contract")
	if err != nil || !publish {
		return err
	}

	return nil
}

func (cmd *participateCmd) runCommand(sct eth.SwapContractTransactor) error {
	output, err := eth.Participate(context.Background(), sct, cmd.cp1Addr, cmd.amount, cmd.secretHash)
	if err != nil {
		return errors.Wrap(err, "failed to participate in atomic swap")
	}

	fmt.Printf("Amount: %s Wei (%s ETH)\n\n",
		cmd.amount.String(), formatWeiAsEthString(cmd.amount))

	fmt.Printf("Author's refund address: %x\n\n", sct.FromAddr)

	fmt.Printf("Chain ID:         %s\n", chainConfig.ChainID.String())
	fmt.Printf("Contract Address: %x\n", sct.ContractAddr)

	fmt.Printf("Contract transaction (%x):\n", output.ContractTransactionHash)

	return nil
}

func (cmd *redeemCmd) runCommand(sct eth.SwapContractTransactor) error {
	params, err := unpackContractInputParams(sct.Abi, cmd.contractTx)
	if err != nil {
		return err
	}
	output, err := eth.Redeem(context.Background(), sct, params.SecretHash, cmd.secret)
	if err != nil {
		return fmt.Errorf("failed to create redeem TX: %v", err)
	}

	fmt.Printf("Chain ID:         %s\n", chainConfig.ChainID.String())
	fmt.Printf("Contract Address: %x\n", sct.ContractAddr)

	fmt.Printf("Redeem transaction (%x):\n", output.RedeemTxHash)

	return nil
}

func (cmd *refundCmd) runCommand(sct eth.SwapContractTransactor) error {
	output, err := eth.Refund(context.Background(), sct, cmd.contractTx)
	if err != nil {
		return fmt.Errorf("failed to create refund TX: %v", err)
	}

	fmt.Printf("Chain ID:         %s\n", chainConfig.ChainID.String())
	fmt.Printf("Contract Address: %x\n", sct.ContractAddr)

	fmt.Printf("Refund transaction (%x):\n", output)
	return nil
}

func (cmd *extractSecretCmd) runCommand(eth.SwapContractTransactor) error {
	return cmd.runOfflineCommand()
}

func (cmd *extractSecretCmd) runOfflineCommand() error {
	abi, err := abi.JSON(strings.NewReader(contract.ContractABI))
	if err != nil {
		return fmt.Errorf("failed to read (smart) contract ABI: %v", err)
	}

	txData := cmd.redemptionTx.Data()

	// first 4 bytes contain the id, so let's get method using that ID
	method, err := abi.MethodById(txData[:4])
	if err != nil {
		return fmt.Errorf("failed to get method using its parsed id: %v", err)
	}
	if method.Name != "redeem" {
		return fmt.Errorf("unexpected name for unpacked method ID: %s", method.Name)
	}

	// prepare the params
	params := struct {
		Secret     [sha256.Size]byte
		SecretHash [sha256.Size]byte
	}{}

	// unpack the params
	args, err := method.Inputs.Unpack(txData[4:])
	if err != nil {
		return fmt.Errorf("failed to unpack method's input params: %v", err)
	}

	if len(args) > 2 {
		return errors.New("unexpected arguments count")
	}

	params.Secret = args[0].([sha256.Size]byte)
	params.SecretHash = args[1].([sha256.Size]byte)

	// ensure secret hash is the same as the given one
	if cmd.secretHash != params.SecretHash {
		return fmt.Errorf("unexpected secret hash found: %x", params.SecretHash)
	}
	secretHash := sha256Hash(params.Secret[:])
	if params.SecretHash != secretHash {
		return fmt.Errorf("unexpected secret found: %x", params.Secret)
	}

	// print secret
	fmt.Printf("Secret: %x\n", params.Secret)
	return nil
}

func (cmd *auditContractCmd) runCommand(sct eth.SwapContractTransactor) error {
	// unpack input params from contract tx
	params, err := unpackContractInputParams(sct.Abi, cmd.contractTx)
	if err != nil {
		return err
	}

	output, err := eth.AuditContract(context.Background(), sct, cmd.contractTx)
	if err != nil {
		return errors.Wrap(err, "could not audit conract")
	}

	// print contract info
	fmt.Printf("Contract address:        %x\n", cmd.contractTx.To())
	fmt.Printf("Contract value:          %s ETH\n", formatWeiAsEthString(cmd.contractTx.Value()))
	fmt.Printf("Recipient address:       %x\n", params.ToAddress)
	fmt.Printf("Author's refund address: %x\n\n", output.RefundAddress)

	fmt.Printf("Secret hash: %x\n\n", params.SecretHash)

	// NOTE:
	// the reason we require th node for this method,
	// is because we need to be able to know the transaction's timestamp

	lockTime := time.Unix(output.Locktime, 0)
	fmt.Printf("Locktime: %v\n", lockTime.UTC())
	reachedAt := lockTime.Sub(time.Now().UTC()).Truncate(time.Second)
	if reachedAt > 0 {
		fmt.Printf("Locktime reached in %v\n", reachedAt)
	} else {
		fmt.Printf("Contract refund time lock has expired\n")
	}
	return nil
}

func (cmd *deployContractCmd) runCommand(sct eth.SwapContractTransactor) error {
	ctx := context.Background()
	tx, err := sct.DeployTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to create deploy TX: %v", err)
	}

	deployTxCost := new(big.Int).Mul(tx.GasPrice(), new(big.Int).SetUint64(tx.Gas()))
	fmt.Printf("Deploy fee: %s ETH\n\n", formatWeiAsEthString(deployTxCost))

	fmt.Printf("Chain ID:         %s\n", chainConfig.ChainID.String())
	fmt.Printf("Contract Address: %x\n", sct.ContractAddr)

	fmt.Printf("Deploy transaction (%x):\n", tx.Hash())
	txBytes, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return fmt.Errorf("failed to encode deploy TX: %v", err)
	}
	fmt.Printf("%x\n\n", txBytes)

	publish, err := promptPublishTx("deploy")
	if err != nil || !publish {
		return err
	}

	err = tx.Send(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Published deploy transaction (%x)\n", tx.Hash())
	return nil
}

func (cmd *validateDeployedContractCmd) runCommand(eth.SwapContractTransactor) error {
	return cmd.runOfflineCommand()
}

func (cmd *validateDeployedContractCmd) runOfflineCommand() error {
	if !bytes.Equal(cmd.deployTx.Data(), contractBin) {
		return errors.New("deployed contract is invalid (make sure to use the same Solidity contract source code and Compiler version (0.4.24))")
	}
	fmt.Println("Contract is valid")
	return nil
}

type (
	// swapContractTransactor allows the creation of transactions for the different
	// atomic swap actions
	swapContractTransactor struct {
		abi          abi.ABI
		signer       bind.SignerFn
		client       *ethClient
		fromAddr     common.Address
		contractAddr common.Address
		autoAccount  bool // defines if an account is automatically selected

		_contract *contract.Contract // created only once
	}

	// swapTransaction adds send functionality to the transaction,
	// such that it can be send in an easy way
	swapTransaction struct {
		*types.Transaction
		client *ethClient
	}
)

func (sct *swapContractTransactor) initiateTx(amount *big.Int, secretHash [sha256.Size]byte, participant common.Address) (*swapTransaction, error) {
	// validate tx does not exist yet,
	// as to provide more meaningful error messages
	switch _, err := sct.getSwapContract(secretHash); err {
	case errNotExists:
		// this is what we want
	case nil:
		return nil, errors.New("secret hash is already used for another atomic swap contract")
	default:
		return nil, fmt.Errorf("unexpected error while checking for an existing contract: %v", err)
	}
	// create initiate tx
	return sct.newTransaction(
		amount, "initiate",
		// lock duration
		big.NewInt(initiateLockPeriodInSeconds),
		// secret hash
		secretHash,
		// participant
		participant,
	)
}

func (sct *swapContractTransactor) participateTx(amount *big.Int, secretHash [sha256.Size]byte, initiator common.Address) (*swapTransaction, error) {
	// validate tx does not exist yet,
	// as to provide more meaningful error messages
	switch _, err := sct.getSwapContract(secretHash); err {
	case errNotExists:
		// this is what we want
	case nil:
		return nil, errors.New("secret hash is already used for another atomic swap contract")
	default:
		return nil, fmt.Errorf("unexpected error while checking for an existing contract: %v", err)
	}
	return sct.newTransaction(
		amount, "participate",
		// lock duration
		big.NewInt(participateLockPeriodInSeconds),
		// secret hash
		secretHash,
		// participant
		initiator,
	)
}

func (sct *swapContractTransactor) redeemTx(secretHash, secret [sha256.Size]byte) (*swapTransaction, error) {
	// validate swap contract,
	// as to provide more meaningful errors
	sc, err := sct.getSwapContract(secretHash)
	if err != nil {
		return nil, err
	}
	if sc.SecretHash != secretHash {
		return nil, errors.New("invalid secret hash registered")
	}
	if userSecretHash := sha256Hash(secret[:]); sc.SecretHash != userSecretHash {
		return nil, errors.New("secret does not match secret hash")
	}
	switch sc.Kind {
	case swapKindInitiator:
		if sc.Participant != sct.fromAddr {
			return nil, fmt.Errorf("only the participant can redeem: unexpected address: %x", sct.fromAddr)
		}
	case swapKindParticipant:
		if sc.Initiator != sct.fromAddr {
			return nil, fmt.Errorf("only the initiator can redeem: unexpected address: %x", sct.fromAddr)
		}
	default:
		return nil, fmt.Errorf("invalid atomic swap contract kind: %d", sc.Kind)
	}
	if sc.State != swapStateFilled {
		return nil, errors.New("inactive atomic swap contract")
	}
	// create redeem tx
	return sct.newTransaction(
		nil, "redeem",
		// secret,
		secret,
		// secret hash
		secretHash,
	)
}

func (sct *swapContractTransactor) refundTx(secretHash [sha256.Size]byte) (*swapTransaction, error) {
	// validate swap contract,
	// as to provide more meaningful errors
	sc, err := sct.getSwapContract(secretHash)
	if err != nil {
		return nil, err
	}
	if sc.SecretHash != secretHash {
		return nil, errors.New("invalid secret hash registered")
	}
	switch sc.Kind {
	case swapKindInitiator:
		if sc.Initiator != sct.fromAddr {
			return nil, fmt.Errorf("only the participant can refund: unexpected address: %x", sct.fromAddr)
		}
	case swapKindParticipant:
		if sc.Participant != sct.fromAddr {
			return nil, fmt.Errorf("only the initiator can refund: unexpected address: %x", sct.fromAddr)
		}
	default:
		return nil, fmt.Errorf("invalid atomic swap contract kind: %d", sc.Kind)
	}
	if sc.State != swapStateFilled {
		return nil, errors.New("inactive atomic swap contract")
	}
	lockTime := time.Unix(bigIntPtrToUint64(sc.InitTimestamp)+bigIntPtrToUint64(sc.RefundTime), 0)
	if dur := time.Until(lockTime).Truncate(time.Second); dur >= 0 {
		return nil, fmt.Errorf("contract is still locked for %v", dur+time.Second)
	}
	// create refund tx
	return sct.newTransaction(
		nil, "refund",
		// secret hash
		secretHash,
	)
}

func bigIntPtrToUint64(i *big.Int) int64 {
	if i == nil {
		return 0
	}
	return i.Int64()
}

func (sct *swapContractTransactor) deployTx() (*swapTransaction, error) {
	return sct.newTransactionWithInput(nil, false, common.FromHex(contract.ContractBin))
}

func (sct *swapContractTransactor) maxGasCost() (*big.Int, error) {
	ctx := newContext()
	gasPrice, err := sct.client.SuggestGasPrice(ctx)
	ctx.Cancel()
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %v", err)
	}
	return gasPrice.Mul(gasPrice, big.NewInt(maxGasLimit)), nil
}

// states have to be mapped 1-to-1 with Enum AtomicSwap.State,
// as found in ./contract/src/contracts/AtomicSwap.sol
//
// This isn't part of the Ethereum-generated Go code found in the child "contract" pkg,
// given that the ABI does not export Enums.
const (
	swapStateEmpty uint8 = iota
	swapStateFilled
	swapStateRedeemed
	swapStateRefunded
)

// kinds have to be mapped 1-to-1 with Enum AtomicSwap.Kind,
// as found in ./contract/src/contracts/AtomicSwap.sol
//
// This isn't part of the Ethereum-generated Go code found in the child "contract" pkg,
// given that the ABI does not export Enums.
const (
	swapKindInitiator uint8 = iota
	swapKindParticipant
)

var (
	// error reported when an atomic swap contract (identified by a secret hash),
	// has the state Empty, indicating it doesn't exist yet.
	errNotExists = errors.New("atomic swap contract does not exist")
)

// getSwapContract is a free contract call,
// which allows us to retrieve an atomic swap contract from a deployed AtomicSwap smart contract,
// using the secret hash used in that atomic swap contract as this contract's identifier.
func (sct *swapContractTransactor) getSwapContract(secretHash [32]byte) (*struct {
	InitTimestamp *big.Int
	RefundTime    *big.Int
	SecretHash    [32]byte
	Secret        [32]byte
	Initiator     common.Address
	Participant   common.Address
	Value         *big.Int
	Kind          uint8
	State         uint8
}, error) {
	if sct._contract == nil {
		var err error
		sct._contract, err = contract.NewContract(sct.contractAddr, sct.client.Client)
		if err != nil {
			return nil, fmt.Errorf("failed to bind smart contract (at %x): %v", sct.contractAddr, err)
		}
	}
	ctx := newContext()
	sc, err := sct._contract.Swaps(&bind.CallOpts{
		Pending: false,
		From:    sct.fromAddr,
		Context: ctx,
	}, secretHash)
	ctx.Cancel()
	if err != nil {
		return nil, fmt.Errorf("failed to get swap contract from smart contract (at %x): %v", sct.contractAddr, err)
	}
	if sc.State == swapStateEmpty {
		return nil, errNotExists
	}
	return &sc, nil
}

func (sct *swapContractTransactor) newTransaction(amount *big.Int, name string, params ...interface{}) (*swapTransaction, error) {
	// pack up the parameters and contract name
	input, err := sct.abi.Pack(name, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack input")
	}
	return sct.newTransactionWithInput(amount, true, input)
}

func (sct *swapContractTransactor) newTransactionWithInput(amount *big.Int, contractCall bool, input []byte) (*swapTransaction, error) {
	// define the TransactOpts for binding
	opts, err := sct.calcBaseOpts(amount)
	if err != nil {
		return nil, err
	}
	opts.GasLimit, err = sct.calcGasLimit(opts.Value, opts.GasPrice, contractCall, input)
	if err != nil {
		return nil, err
	}

	// sign using daemon or do it client-side if desired
	var signedTx *types.Transaction
	if opts.Signer == nil {
		var toAddr *common.Address
		if contractCall {
			toAddr = &sct.contractAddr
		}
		// sign transaction using the daemon
		var result struct {
			Raw string            `json:"raw"`
			Tx  types.Transaction `json:"tx"`
		}
		ctx := newContext()
		err = sct.client.rpcClient.CallContext(ctx, &result, "eth_signTransaction", struct {
			From     common.Address  `json:"from"`
			To       *common.Address `json:"to"`
			Gas      hexutil.Uint64  `json:"gas"`
			GasPrice hexutil.Big     `json:"gasPrice"`
			Value    hexutil.Big     `json:"value"`
			Nonce    hexutil.Uint64  `json:"nonce"`
			Data     hexutil.Bytes   `json:"data"`
		}{
			From:     opts.From,
			To:       toAddr,
			Gas:      hexutil.Uint64(opts.GasLimit),
			GasPrice: hexutil.Big(*opts.GasPrice),
			Value: func() hexutil.Big {
				if amount == nil {
					return hexutil.Big{}
				}
				return hexutil.Big(*amount)
			}(),
			Nonce: hexutil.Uint64(opts.Nonce.Uint64()),
			Data:  hexutil.Bytes(input),
		})
		ctx.Cancel()
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction from daemon: %v", err)
		}
		signedTx = &result.Tx
	} else {
		var rawTx *types.Transaction
		if contractCall {
			rawTx = types.NewTransaction(
				opts.Nonce.Uint64(),
				sct.contractAddr,
				opts.Value,
				opts.GasLimit,
				opts.GasPrice,
				input,
			)
		} else {
			rawTx = types.NewContractCreation(
				opts.Nonce.Uint64(),
				opts.Value,
				opts.GasLimit,
				opts.GasPrice,
				input,
			)
		}
		// sign ourselves
		signedTx, err = opts.Signer(opts.From, rawTx)
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction from client: %v", err)
		}
	}
	return &swapTransaction{
		Transaction: signedTx,
		client:      sct.client,
	}, nil
}

func (sct *swapContractTransactor) calcBaseOpts(amount *big.Int) (*bind.TransactOpts, error) {
	ctx := newContext()
	nonce, err := sct.client.PendingNonceAt(ctx, sct.fromAddr)
	ctx.Cancel()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to retrieve account (%x) nonce: %v",
			sct.fromAddr, err)
	}
	ctx = newContext()
	gasPrice, err := sct.client.SuggestGasPrice(ctx)
	ctx.Cancel()
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %v", err)
	}
	if amount == nil {
		amount = new(big.Int)
	}
	return &bind.TransactOpts{
		From:     sct.fromAddr,
		Nonce:    new(big.Int).SetUint64(nonce),
		Signer:   sct.signer,
		Value:    amount,
		GasPrice: gasPrice,
	}, nil
}

func (sct *swapContractTransactor) calcGasLimit(amount, gasPrice *big.Int, contractCall bool, input []byte) (uint64, error) {
	if contractCall {
		ctx := newContext()
		code, err := sct.client.PendingCodeAt(ctx, sct.contractAddr)
		ctx.Cancel()
		if err != nil {
			return 0, fmt.Errorf("failed to estimate gas needed: %v", err)
		} else if len(code) == 0 {
			return 0, fmt.Errorf("failed to estimate gas needed: %v", bind.ErrNoCode)
		}
	}
	// If the contract surely has code (or code is not needed), estimate the transaction
	msg := ethereum.CallMsg{
		From:  sct.fromAddr,
		Value: amount,
		Data:  input,
	}
	if contractCall {
		msg.To = &sct.contractAddr
	}
	ctx := newContext()
	gasLimit, err := sct.client.EstimateGas(ctx, msg)
	ctx.Cancel()
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas needed: %v", err)
	}
	if contractCall && gasLimit > maxGasLimit {
		return 0, fmt.Errorf("%d exceeds the hardcoded code-call gas limit of %d", gasLimit, maxGasLimit)
	}
	return gasLimit, nil
}

func (st *swapTransaction) Send() error {
	ctx := newContext()
	err := st.client.SendTransaction(ctx, st.Transaction)
	ctx.Cancel()
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}
	return nil
}

func dialClient() (*ethClient, error) {
	c, err := rpc.DialContext(context.Background(), *connectFlag)
	if err != nil {
		return nil, err
	}
	return &ethClient{
		Client:    ethclient.NewClient(c),
		rpcClient: c,
	}, nil
}

type ethClient struct {
	*ethclient.Client
	rpcClient *rpc.Client
}

// newContext creates a context which HAS
// to be manually cancelled, as to not leak any resources
func newContext() *cancelableContext {
	if *timeoutFlag == 0 {
		ctx, cancelFn := context.WithCancel(context.Background())
		return &cancelableContext{
			Context:  ctx,
			cancelFn: cancelFn,
		}
	}
	ctx, cancelFn := context.WithTimeout(context.Background(), *timeoutFlag)
	return &cancelableContext{
		Context:  ctx,
		cancelFn: cancelFn,
	}
}

type cancelableContext struct {
	context.Context
	cancelFn context.CancelFunc
}

func (cc *cancelableContext) Cancel() {
	cc.cancelFn()
}

var (
	// decode the byte code of the smart contract used
	// during the initialisation phase of this CLI tool,
	// as to ensure the hex-encoded string is valid at all times.
	//
	// This prevents of having a hidden error,
	// due to the fact that it is only ever used in
	// our extra smart-contract-related commands.
	contractBin = func() []byte {
		b, err := hex.DecodeString(contract.ContractBin)
		if err != nil {
			panic("invalid binary contract: " + err.Error())
		}
		return b
	}()
)
