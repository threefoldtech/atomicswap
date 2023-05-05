package eth

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/threefoldtech/atomicswap/eth/contract"
)

type (
	// SwapContractTransactor allows the creation of transactions for the different
	// atomic swap actions
	SwapContractTransactor struct {
		Abi          abi.ABI
		signer       bind.SignerFn
		Client       *EthClient
		FromAddr     common.Address
		ContractAddr common.Address
		autoAccount  bool // defines if an account is automatically selected

		_contract *contract.Contract // created only once
	}

	// swapTransaction adds send functionality to the transaction,
	// such that it can be send in an easy way
	swapTransaction struct {
		*types.Transaction
		client *EthClient
	}
)

const (
	initiateLockPeriodInSeconds    = 48 * 60 * 60
	participateLockPeriodInSeconds = 24 * 60 * 60

	maxGasLimit = 210000
)

func (sct *SwapContractTransactor) initiateTx(ctx context.Context, amount *big.Int, secretHash [sha256.Size]byte, participant common.Address) (*swapTransaction, error) {
	// validate tx does not exist yet,
	// as to provide more meaningful error messages
	switch _, err := sct.getSwapContract(ctx, secretHash); err {
	case errNotExists:
		// this is what we want
	case nil:
		return nil, errors.New("secret hash is already used for another atomic swap contract")
	default:
		return nil, fmt.Errorf("unexpected error while checking for an existing contract: %v", err)
	}
	// create initiate tx
	return sct.newTransaction(
		ctx,
		amount, "initiate",
		// lock duration
		big.NewInt(initiateLockPeriodInSeconds),
		// secret hash
		secretHash,
		// participant
		participant,
	)
}

func (sct *SwapContractTransactor) participateTx(ctx context.Context, amount *big.Int, secretHash [sha256.Size]byte, initiator common.Address) (*swapTransaction, error) {
	// validate tx does not exist yet,
	// as to provide more meaningful error messages
	switch _, err := sct.getSwapContract(ctx, secretHash); err {
	case errNotExists:
		// this is what we want
	case nil:
		return nil, errors.New("secret hash is already used for another atomic swap contract")
	default:
		return nil, fmt.Errorf("unexpected error while checking for an existing contract: %v", err)
	}
	return sct.newTransaction(
		ctx,
		amount, "participate",
		// lock duration
		big.NewInt(participateLockPeriodInSeconds),
		// secret hash
		secretHash,
		// participant
		initiator,
	)
}

func (sct *SwapContractTransactor) redeemTx(ctx context.Context, secretHash, secret [sha256.Size]byte) (*swapTransaction, error) {
	// validate swap contract,
	// as to provide more meaningful errors
	sc, err := sct.getSwapContract(ctx, secretHash)
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
		if sc.Participant != sct.FromAddr {
			return nil, fmt.Errorf("only the participant can redeem: unexpected address: %x", sct.FromAddr)
		}
	case swapKindParticipant:
		if sc.Initiator != sct.FromAddr {
			return nil, fmt.Errorf("only the initiator can redeem: unexpected address: %x", sct.FromAddr)
		}
	default:
		return nil, fmt.Errorf("invalid atomic swap contract kind: %d", sc.Kind)
	}
	if sc.State != swapStateFilled {
		return nil, errors.New("inactive atomic swap contract")
	}
	// create redeem tx
	return sct.newTransaction(
		ctx,
		nil, "redeem",
		// secret,
		secret,
		// secret hash
		secretHash,
	)
}

func (sct *SwapContractTransactor) refundTx(ctx context.Context, secretHash [sha256.Size]byte) (*swapTransaction, error) {
	// validate swap contract,
	// as to provide more meaningful errors
	sc, err := sct.getSwapContract(ctx, secretHash)
	if err != nil {
		return nil, err
	}
	if sc.SecretHash != secretHash {
		return nil, errors.New("invalid secret hash registered")
	}
	switch sc.Kind {
	case swapKindInitiator:
		if sc.Initiator != sct.FromAddr {
			return nil, fmt.Errorf("only the participant can refund: unexpected address: %x", sct.FromAddr)
		}
	case swapKindParticipant:
		if sc.Participant != sct.FromAddr {
			return nil, fmt.Errorf("only the initiator can refund: unexpected address: %x", sct.FromAddr)
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
		ctx,
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

func (sct *SwapContractTransactor) DeployTx(ctx context.Context) (*swapTransaction, error) {
	return sct.newTransactionWithInput(ctx, nil, false, common.FromHex(contract.ContractBin))
}

func (sct *SwapContractTransactor) maxGasCost(ctx context.Context) (*big.Int, error) {
	gasPrice, err := sct.Client.SuggestGasPrice(ctx)
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
func (sct *SwapContractTransactor) getSwapContract(ctx context.Context, secretHash [32]byte) (*struct {
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
		sct._contract, err = contract.NewContract(sct.ContractAddr, sct.Client.Client)
		if err != nil {
			return nil, fmt.Errorf("failed to bind smart contract (at %x): %v", sct.ContractAddr, err)
		}
	}
	sc, err := sct._contract.Swaps(&bind.CallOpts{
		Pending: false,
		From:    sct.FromAddr,
		Context: ctx,
	}, secretHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get swap contract from smart contract (at %x): %v", sct.ContractAddr, err)
	}
	if sc.State == swapStateEmpty {
		return nil, errNotExists
	}
	return &sc, nil
}

func (sct *SwapContractTransactor) newTransaction(ctx context.Context, amount *big.Int, name string, params ...interface{}) (*swapTransaction, error) {
	// pack up the parameters and contract name
	input, err := sct.Abi.Pack(name, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack input")
	}
	return sct.newTransactionWithInput(ctx, amount, true, input)
}

func (sct *SwapContractTransactor) newTransactionWithInput(ctx context.Context, amount *big.Int, contractCall bool, input []byte) (*swapTransaction, error) {
	// define the TransactOpts for binding
	opts, err := sct.calcBaseOpts(ctx, amount)
	if err != nil {
		return nil, err
	}
	opts.GasLimit, err = sct.calcGasLimit(ctx, opts.Value, opts.GasPrice, contractCall, input)
	if err != nil {
		return nil, err
	}

	// sign using daemon or do it client-side if desired
	var signedTx *types.Transaction
	if opts.Signer == nil {
		var toAddr *common.Address
		if contractCall {
			toAddr = &sct.ContractAddr
		}
		// sign transaction using the daemon
		var result struct {
			Raw string            `json:"raw"`
			Tx  types.Transaction `json:"tx"`
		}
		err = sct.Client.rpcClient.CallContext(ctx, &result, "eth_signTransaction", struct {
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
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction from daemon: %v", err)
		}
		signedTx = &result.Tx
	} else {
		var rawTx *types.Transaction
		if contractCall {
			rawTx = types.NewTransaction(
				opts.Nonce.Uint64(),
				sct.ContractAddr,
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
		client:      sct.Client,
	}, nil
}

func (sct *SwapContractTransactor) calcBaseOpts(ctx context.Context, amount *big.Int) (*bind.TransactOpts, error) {
	nonce, err := sct.Client.PendingNonceAt(ctx, sct.FromAddr)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to retrieve account (%x) nonce: %v",
			sct.FromAddr, err)
	}
	gasPrice, err := sct.Client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %v", err)
	}
	if amount == nil {
		amount = new(big.Int)
	}
	return &bind.TransactOpts{
		From:     sct.FromAddr,
		Nonce:    new(big.Int).SetUint64(nonce),
		Signer:   sct.signer,
		Value:    amount,
		GasPrice: gasPrice,
	}, nil
}

func (sct *SwapContractTransactor) calcGasLimit(ctx context.Context, amount, gasPrice *big.Int, contractCall bool, input []byte) (uint64, error) {
	if contractCall {
		code, err := sct.Client.PendingCodeAt(ctx, sct.ContractAddr)
		if err != nil {
			return 0, fmt.Errorf("failed to estimate gas needed: %v", err)
		} else if len(code) == 0 {
			return 0, fmt.Errorf("failed to estimate gas needed: %v", bind.ErrNoCode)
		}
	}
	// If the contract surely has code (or code is not needed), estimate the transaction
	msg := ethereum.CallMsg{
		From:  sct.FromAddr,
		Value: amount,
		Data:  input,
	}
	if contractCall {
		msg.To = &sct.ContractAddr
	}
	gasLimit, err := sct.Client.EstimateGas(ctx, msg)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas needed: %v", err)
	}
	if contractCall && gasLimit > maxGasLimit {
		return 0, fmt.Errorf("%d exceeds the hardcoded code-call gas limit of %d", gasLimit, maxGasLimit)
	}
	return gasLimit, nil
}

func (st *swapTransaction) Send(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	err := st.client.SendTransaction(ctx, st.Transaction)
	cancel()
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}
	return nil
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

func sha256Hash(x []byte) [sha256.Size]byte {
	h := sha256.Sum256(x)
	return h
}

// newSwapContractTransactor creates a new swapContract instance,
// see swapContractTransactor for more information
func NewSwapContractTransactor(ctx context.Context, c *EthClient, contractAddr common.Address, key *ecdsa.PrivateKey, chainID *big.Int) (SwapContractTransactor, error) {
	parsed, err := abi.JSON(strings.NewReader(contract.ContractABI))
	if err != nil {
		return SwapContractTransactor{}, fmt.Errorf("failed to read (smart) contract ABI: %v", err)
	}
	// sign using given key
	var signer bind.SignerFn
	var fromAddr common.Address
	if key != nil {
		signer, fromAddr, err = newSigner(key, chainID)
		if err != nil {
			return SwapContractTransactor{}, fmt.Errorf("failed to create tx signer: %v", err)
		}
	}
	return SwapContractTransactor{
		Abi:          parsed,
		signer:       signer,
		Client:       c,
		FromAddr:     fromAddr,
		ContractAddr: contractAddr,
	}, nil
}

// newSigner creates a signer func using the flag-passed
// private credentials of the sender
func newSigner(privKey *ecdsa.PrivateKey, chainID *big.Int) (bind.SignerFn, common.Address, error) {
	// json, err := ioutil.ReadFile(path)
	// if err != nil {
	// 	return nil, common.Address{}, fmt.Errorf("failed to read encrypted account/key file (%s) content: %v", path, err)
	// }
	// passphrase, err := speakeasy.Ask("Account passphrase: ")
	// if err != nil {
	// 	return nil, common.Address{}, fmt.Errorf("failed to get passphrase from STDIN: %v", err)
	// }
	// key, err := keystore.DecryptKey(json, passphrase)
	// if err != nil {
	// 	return nil, common.Address{}, fmt.Errorf("failed to decrypt (JSON) account/key file (%s): %v", path, err)
	// }
	keyAddr := crypto.PubkeyToAddress(privKey.PublicKey)
	return func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if address != keyAddr {
			return nil, errors.New("not authorized to sign this account")
		}
		s := types.NewEIP155Signer(chainID)
		signature, err := crypto.Sign(s.Hash(tx).Bytes(), privKey)
		if err != nil {
			return nil, err
		}
		return tx.WithSignature(s, signature)
	}, keyAddr, nil
}
