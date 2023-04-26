// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// ContractMetaData contains all meta data concerning the Contract contract.
var ContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"initTimestamp\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"refundTime\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"secretHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"initiator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Initiated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"initTimestamp\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"refundTime\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"secretHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"initiator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Participated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"redeemTime\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"secretHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"secret\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"redeemer\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Redeemed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"refundTime\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"secretHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"refunder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Refunded\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"refundTime\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"secretHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"}],\"name\":\"initiate\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"refundTime\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"secretHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"initiator\",\"type\":\"address\"}],\"name\":\"participate\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"secret\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"secretHash\",\"type\":\"bytes32\"}],\"name\":\"redeem\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"secretHash\",\"type\":\"bytes32\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"swaps\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"initTimestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"refundTime\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"secretHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"secret\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"initiator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"enumAtomicSwap.Kind\",\"name\":\"kind\",\"type\":\"uint8\"},{\"internalType\":\"enumAtomicSwap.State\",\"name\":\"state\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b506111f4806100206000396000f3fe60806040526004361061004a5760003560e01c80631aa028531461004f5780637249fbb61461006b578063ae05214714610094578063b31597ad146100b0578063eb84e7f2146100d9575b600080fd5b61006960048036038101906100649190610d23565b61011e565b005b34801561007757600080fd5b50610092600480360381019061008d9190610d76565b610361565b005b6100ae60048036038101906100a99190610d23565b610633565b005b3480156100bc57600080fd5b506100d760048036038101906100d29190610da3565b610876565b005b3480156100e557600080fd5b5061010060048036038101906100fb9190610d76565b610bac565b60405161011599989796959493929190610ecf565b60405180910390f35b826000341161012c57600080fd5b6000811161013957600080fd5b826000600381111561014e5761014d610e10565b5b60008083815260200190815260200160002060070160019054906101000a900460ff16600381111561018357610182610e10565b5b1461018d57600080fd5b4260008086815260200190815260200160002060000181905550846000808681526020019081526020016000206001018190555083600080868152602001908152602001600020600201819055508260008086815260200190815260200160002060040160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055503360008086815260200190815260200160002060050160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055503460008086815260200190815260200160002060060181905550600160008086815260200190815260200160002060070160006101000a81548160ff021916908360018111156102d6576102d5610e10565b5b0217905550600160008086815260200190815260200160002060070160016101000a81548160ff0219169083600381111561031457610313610e10565b5b02179055507fe5571d467a528d7481c0e3bdd55ad528d0df6b457b07bab736c3e245c3aa16f442868686333460405161035296959493929190610f5c565b60405180910390a15050505050565b80336001600381111561037757610376610e10565b5b60008084815260200190815260200160002060070160019054906101000a900460ff1660038111156103ac576103ab610e10565b5b146103b657600080fd5b6001808111156103c9576103c8610e10565b5b60008084815260200190815260200160002060070160009054906101000a900460ff1660018111156103fe576103fd610e10565b5b03610475578073ffffffffffffffffffffffffffffffffffffffff1660008084815260200190815260200160002060050160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff161461047057600080fd5b6104e3565b8073ffffffffffffffffffffffffffffffffffffffff1660008084815260200190815260200160002060040160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16146104e257600080fd5b5b6000806000848152602001908152602001600020600001549050600080848152602001908152602001600020600101548161051e9190610fec565b905080421161052c57600080fd5b3373ffffffffffffffffffffffffffffffffffffffff166108fc600080878152602001908152602001600020600601549081150290604051600060405180830381858888f19350505050158015610587573d6000803e3d6000fd5b50600360008086815260200190815260200160002060070160016101000a81548160ff021916908360038111156105c1576105c0610e10565b5b02179055507fadb1dca52dfad065e50a1e25c2ee47ae54013a1f2d6f8ea5abace52eb4b7a4c8426000808781526020019081526020016000206002015433600080898152602001908152602001600020600601546040516106259493929190611020565b60405180910390a150505050565b826000341161064157600080fd5b6000811161064e57600080fd5b826000600381111561066357610662610e10565b5b60008083815260200190815260200160002060070160019054906101000a900460ff16600381111561069857610697610e10565b5b146106a257600080fd5b4260008086815260200190815260200160002060000181905550846000808681526020019081526020016000206001018190555083600080868152602001908152602001600020600201819055503360008086815260200190815260200160002060040160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508260008086815260200190815260200160002060050160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055503460008086815260200190815260200160002060060181905550600080600086815260200190815260200160002060070160006101000a81548160ff021916908360018111156107eb576107ea610e10565b5b0217905550600160008086815260200190815260200160002060070160016101000a81548160ff0219169083600381111561082957610828610e10565b5b02179055507f75501a491c11746724d18ea6e5ac6a53864d886d653da6b846fdecda837cf57642868633873460405161086796959493929190610f5c565b60405180910390a15050505050565b8082336001600381111561088d5761088c610e10565b5b60008085815260200190815260200160002060070160019054906101000a900460ff1660038111156108c2576108c1610e10565b5b146108cc57600080fd5b6001808111156108df576108de610e10565b5b60008085815260200190815260200160002060070160009054906101000a900460ff16600181111561091457610913610e10565b5b0361098b578073ffffffffffffffffffffffffffffffffffffffff1660008085815260200190815260200160002060040160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff161461098657600080fd5b6109f9565b8073ffffffffffffffffffffffffffffffffffffffff1660008085815260200190815260200160002060050160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16146109f857600080fd5b5b82600283604051602001610a0d9190611086565b604051602081830303815290604052604051610a299190611112565b602060405180830381855afa158015610a46573d6000803e3d6000fd5b5050506040513d601f19601f82011682018060405250810190610a69919061113e565b14610a7357600080fd5b3373ffffffffffffffffffffffffffffffffffffffff166108fc600080878152602001908152602001600020600601549081150290604051600060405180830381858888f19350505050158015610ace573d6000803e3d6000fd5b50600260008086815260200190815260200160002060070160016101000a81548160ff02191690836003811115610b0857610b07610e10565b5b021790555084600080868152602001908152602001600020600301819055507fe4da013d8c42cdfa76ab1d5c08edcdc1503d2da88d7accc854f0e57ebe45c591426000808781526020019081526020016000206002015460008088815260200190815260200160002060030154336000808a815260200190815260200160002060060154604051610b9d95949392919061116b565b60405180910390a15050505050565b60006020528060005260406000206000915090508060000154908060010154908060020154908060030154908060040160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16908060050160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16908060060154908060070160009054906101000a900460ff16908060070160019054906101000a900460ff16905089565b600080fd5b6000819050919050565b610c6c81610c59565b8114610c7757600080fd5b50565b600081359050610c8981610c63565b92915050565b6000819050919050565b610ca281610c8f565b8114610cad57600080fd5b50565b600081359050610cbf81610c99565b92915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610cf082610cc5565b9050919050565b610d0081610ce5565b8114610d0b57600080fd5b50565b600081359050610d1d81610cf7565b92915050565b600080600060608486031215610d3c57610d3b610c54565b5b6000610d4a86828701610c7a565b9350506020610d5b86828701610cb0565b9250506040610d6c86828701610d0e565b9150509250925092565b600060208284031215610d8c57610d8b610c54565b5b6000610d9a84828501610cb0565b91505092915050565b60008060408385031215610dba57610db9610c54565b5b6000610dc885828601610cb0565b9250506020610dd985828601610cb0565b9150509250929050565b610dec81610c59565b82525050565b610dfb81610c8f565b82525050565b610e0a81610ce5565b82525050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b60028110610e5057610e4f610e10565b5b50565b6000819050610e6182610e3f565b919050565b6000610e7182610e53565b9050919050565b610e8181610e66565b82525050565b60048110610e9857610e97610e10565b5b50565b6000819050610ea982610e87565b919050565b6000610eb982610e9b565b9050919050565b610ec981610eae565b82525050565b600061012082019050610ee5600083018c610de3565b610ef2602083018b610de3565b610eff604083018a610df2565b610f0c6060830189610df2565b610f196080830188610e01565b610f2660a0830187610e01565b610f3360c0830186610de3565b610f4060e0830185610e78565b610f4e610100830184610ec0565b9a9950505050505050505050565b600060c082019050610f716000830189610de3565b610f7e6020830188610de3565b610f8b6040830187610df2565b610f986060830186610e01565b610fa56080830185610e01565b610fb260a0830184610de3565b979650505050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000610ff782610c59565b915061100283610c59565b925082820190508082111561101a57611019610fbd565b5b92915050565b60006080820190506110356000830187610de3565b6110426020830186610df2565b61104f6040830185610e01565b61105c6060830184610de3565b95945050505050565b6000819050919050565b61108061107b82610c8f565b611065565b82525050565b6000611092828461106f565b60208201915081905092915050565b600081519050919050565b600081905092915050565b60005b838110156110d55780820151818401526020810190506110ba565b60008484015250505050565b60006110ec826110a1565b6110f681856110ac565b93506111068185602086016110b7565b80840191505092915050565b600061111e82846110e1565b915081905092915050565b60008151905061113881610c99565b92915050565b60006020828403121561115457611153610c54565b5b600061116284828501611129565b91505092915050565b600060a0820190506111806000830188610de3565b61118d6020830187610df2565b61119a6040830186610df2565b6111a76060830185610e01565b6111b46080830184610de3565b969550505050505056fea2646970667358221220cf88348547b8c65b4bb6ca4783a030cbd46367a1b1974b85abfb62ee9f099a5264736f6c63430008130033",
}

// ContractABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractMetaData.ABI instead.
var ContractABI = ContractMetaData.ABI

// ContractBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractMetaData.Bin instead.
var ContractBin = ContractMetaData.Bin

// DeployContract deploys a new Ethereum contract, binding an instance of Contract to it.
func DeployContract(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Contract, error) {
	parsed, err := ContractMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// Contract is an auto generated Go binding around an Ethereum contract.
type Contract struct {
	ContractCaller     // Read-only binding to the contract
	ContractTransactor // Write-only binding to the contract
	ContractFilterer   // Log filterer for contract events
}

// ContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractSession struct {
	Contract     *Contract         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractCallerSession struct {
	Contract *ContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractTransactorSession struct {
	Contract     *ContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractRaw struct {
	Contract *Contract // Generic contract binding to access the raw methods on
}

// ContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractCallerRaw struct {
	Contract *ContractCaller // Generic read-only contract binding to access the raw methods on
}

// ContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractTransactorRaw struct {
	Contract *ContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContract creates a new instance of Contract, bound to a specific deployed contract.
func NewContract(address common.Address, backend bind.ContractBackend) (*Contract, error) {
	contract, err := bindContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// NewContractCaller creates a new read-only instance of Contract, bound to a specific deployed contract.
func NewContractCaller(address common.Address, caller bind.ContractCaller) (*ContractCaller, error) {
	contract, err := bindContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractCaller{contract: contract}, nil
}

// NewContractTransactor creates a new write-only instance of Contract, bound to a specific deployed contract.
func NewContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractTransactor, error) {
	contract, err := bindContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractTransactor{contract: contract}, nil
}

// NewContractFilterer creates a new log filterer instance of Contract, bound to a specific deployed contract.
func NewContractFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractFilterer, error) {
	contract, err := bindContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractFilterer{contract: contract}, nil
}

// bindContract binds a generic wrapper to an already deployed contract.
func bindContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.ContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transact(opts, method, params...)
}

// Swaps is a free data retrieval call binding the contract method 0xeb84e7f2.
//
// Solidity: function swaps(bytes32 ) view returns(uint256 initTimestamp, uint256 refundTime, bytes32 secretHash, bytes32 secret, address initiator, address participant, uint256 value, uint8 kind, uint8 state)
func (_Contract *ContractCaller) Swaps(opts *bind.CallOpts, arg0 [32]byte) (struct {
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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "swaps", arg0)

	outstruct := new(struct {
		InitTimestamp *big.Int
		RefundTime    *big.Int
		SecretHash    [32]byte
		Secret        [32]byte
		Initiator     common.Address
		Participant   common.Address
		Value         *big.Int
		Kind          uint8
		State         uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.InitTimestamp = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.RefundTime = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.SecretHash = *abi.ConvertType(out[2], new([32]byte)).(*[32]byte)
	outstruct.Secret = *abi.ConvertType(out[3], new([32]byte)).(*[32]byte)
	outstruct.Initiator = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.Participant = *abi.ConvertType(out[5], new(common.Address)).(*common.Address)
	outstruct.Value = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)
	outstruct.Kind = *abi.ConvertType(out[7], new(uint8)).(*uint8)
	outstruct.State = *abi.ConvertType(out[8], new(uint8)).(*uint8)

	return *outstruct, err

}

// Swaps is a free data retrieval call binding the contract method 0xeb84e7f2.
//
// Solidity: function swaps(bytes32 ) view returns(uint256 initTimestamp, uint256 refundTime, bytes32 secretHash, bytes32 secret, address initiator, address participant, uint256 value, uint8 kind, uint8 state)
func (_Contract *ContractSession) Swaps(arg0 [32]byte) (struct {
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
	return _Contract.Contract.Swaps(&_Contract.CallOpts, arg0)
}

// Swaps is a free data retrieval call binding the contract method 0xeb84e7f2.
//
// Solidity: function swaps(bytes32 ) view returns(uint256 initTimestamp, uint256 refundTime, bytes32 secretHash, bytes32 secret, address initiator, address participant, uint256 value, uint8 kind, uint8 state)
func (_Contract *ContractCallerSession) Swaps(arg0 [32]byte) (struct {
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
	return _Contract.Contract.Swaps(&_Contract.CallOpts, arg0)
}

// Initiate is a paid mutator transaction binding the contract method 0xae052147.
//
// Solidity: function initiate(uint256 refundTime, bytes32 secretHash, address participant) payable returns()
func (_Contract *ContractTransactor) Initiate(opts *bind.TransactOpts, refundTime *big.Int, secretHash [32]byte, participant common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "initiate", refundTime, secretHash, participant)
}

// Initiate is a paid mutator transaction binding the contract method 0xae052147.
//
// Solidity: function initiate(uint256 refundTime, bytes32 secretHash, address participant) payable returns()
func (_Contract *ContractSession) Initiate(refundTime *big.Int, secretHash [32]byte, participant common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Initiate(&_Contract.TransactOpts, refundTime, secretHash, participant)
}

// Initiate is a paid mutator transaction binding the contract method 0xae052147.
//
// Solidity: function initiate(uint256 refundTime, bytes32 secretHash, address participant) payable returns()
func (_Contract *ContractTransactorSession) Initiate(refundTime *big.Int, secretHash [32]byte, participant common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Initiate(&_Contract.TransactOpts, refundTime, secretHash, participant)
}

// Participate is a paid mutator transaction binding the contract method 0x1aa02853.
//
// Solidity: function participate(uint256 refundTime, bytes32 secretHash, address initiator) payable returns()
func (_Contract *ContractTransactor) Participate(opts *bind.TransactOpts, refundTime *big.Int, secretHash [32]byte, initiator common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "participate", refundTime, secretHash, initiator)
}

// Participate is a paid mutator transaction binding the contract method 0x1aa02853.
//
// Solidity: function participate(uint256 refundTime, bytes32 secretHash, address initiator) payable returns()
func (_Contract *ContractSession) Participate(refundTime *big.Int, secretHash [32]byte, initiator common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Participate(&_Contract.TransactOpts, refundTime, secretHash, initiator)
}

// Participate is a paid mutator transaction binding the contract method 0x1aa02853.
//
// Solidity: function participate(uint256 refundTime, bytes32 secretHash, address initiator) payable returns()
func (_Contract *ContractTransactorSession) Participate(refundTime *big.Int, secretHash [32]byte, initiator common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Participate(&_Contract.TransactOpts, refundTime, secretHash, initiator)
}

// Redeem is a paid mutator transaction binding the contract method 0xb31597ad.
//
// Solidity: function redeem(bytes32 secret, bytes32 secretHash) returns()
func (_Contract *ContractTransactor) Redeem(opts *bind.TransactOpts, secret [32]byte, secretHash [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "redeem", secret, secretHash)
}

// Redeem is a paid mutator transaction binding the contract method 0xb31597ad.
//
// Solidity: function redeem(bytes32 secret, bytes32 secretHash) returns()
func (_Contract *ContractSession) Redeem(secret [32]byte, secretHash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Redeem(&_Contract.TransactOpts, secret, secretHash)
}

// Redeem is a paid mutator transaction binding the contract method 0xb31597ad.
//
// Solidity: function redeem(bytes32 secret, bytes32 secretHash) returns()
func (_Contract *ContractTransactorSession) Redeem(secret [32]byte, secretHash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Redeem(&_Contract.TransactOpts, secret, secretHash)
}

// Refund is a paid mutator transaction binding the contract method 0x7249fbb6.
//
// Solidity: function refund(bytes32 secretHash) returns()
func (_Contract *ContractTransactor) Refund(opts *bind.TransactOpts, secretHash [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "refund", secretHash)
}

// Refund is a paid mutator transaction binding the contract method 0x7249fbb6.
//
// Solidity: function refund(bytes32 secretHash) returns()
func (_Contract *ContractSession) Refund(secretHash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Refund(&_Contract.TransactOpts, secretHash)
}

// Refund is a paid mutator transaction binding the contract method 0x7249fbb6.
//
// Solidity: function refund(bytes32 secretHash) returns()
func (_Contract *ContractTransactorSession) Refund(secretHash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Refund(&_Contract.TransactOpts, secretHash)
}

// ContractInitiatedIterator is returned from FilterInitiated and is used to iterate over the raw logs and unpacked data for Initiated events raised by the Contract contract.
type ContractInitiatedIterator struct {
	Event *ContractInitiated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractInitiated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractInitiated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractInitiated represents a Initiated event raised by the Contract contract.
type ContractInitiated struct {
	InitTimestamp *big.Int
	RefundTime    *big.Int
	SecretHash    [32]byte
	Initiator     common.Address
	Participant   common.Address
	Value         *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterInitiated is a free log retrieval operation binding the contract event 0x75501a491c11746724d18ea6e5ac6a53864d886d653da6b846fdecda837cf576.
//
// Solidity: event Initiated(uint256 initTimestamp, uint256 refundTime, bytes32 secretHash, address initiator, address participant, uint256 value)
func (_Contract *ContractFilterer) FilterInitiated(opts *bind.FilterOpts) (*ContractInitiatedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "Initiated")
	if err != nil {
		return nil, err
	}
	return &ContractInitiatedIterator{contract: _Contract.contract, event: "Initiated", logs: logs, sub: sub}, nil
}

// WatchInitiated is a free log subscription operation binding the contract event 0x75501a491c11746724d18ea6e5ac6a53864d886d653da6b846fdecda837cf576.
//
// Solidity: event Initiated(uint256 initTimestamp, uint256 refundTime, bytes32 secretHash, address initiator, address participant, uint256 value)
func (_Contract *ContractFilterer) WatchInitiated(opts *bind.WatchOpts, sink chan<- *ContractInitiated) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "Initiated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractInitiated)
				if err := _Contract.contract.UnpackLog(event, "Initiated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitiated is a log parse operation binding the contract event 0x75501a491c11746724d18ea6e5ac6a53864d886d653da6b846fdecda837cf576.
//
// Solidity: event Initiated(uint256 initTimestamp, uint256 refundTime, bytes32 secretHash, address initiator, address participant, uint256 value)
func (_Contract *ContractFilterer) ParseInitiated(log types.Log) (*ContractInitiated, error) {
	event := new(ContractInitiated)
	if err := _Contract.contract.UnpackLog(event, "Initiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractParticipatedIterator is returned from FilterParticipated and is used to iterate over the raw logs and unpacked data for Participated events raised by the Contract contract.
type ContractParticipatedIterator struct {
	Event *ContractParticipated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractParticipatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractParticipated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractParticipated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractParticipatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractParticipatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractParticipated represents a Participated event raised by the Contract contract.
type ContractParticipated struct {
	InitTimestamp *big.Int
	RefundTime    *big.Int
	SecretHash    [32]byte
	Initiator     common.Address
	Participant   common.Address
	Value         *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterParticipated is a free log retrieval operation binding the contract event 0xe5571d467a528d7481c0e3bdd55ad528d0df6b457b07bab736c3e245c3aa16f4.
//
// Solidity: event Participated(uint256 initTimestamp, uint256 refundTime, bytes32 secretHash, address initiator, address participant, uint256 value)
func (_Contract *ContractFilterer) FilterParticipated(opts *bind.FilterOpts) (*ContractParticipatedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "Participated")
	if err != nil {
		return nil, err
	}
	return &ContractParticipatedIterator{contract: _Contract.contract, event: "Participated", logs: logs, sub: sub}, nil
}

// WatchParticipated is a free log subscription operation binding the contract event 0xe5571d467a528d7481c0e3bdd55ad528d0df6b457b07bab736c3e245c3aa16f4.
//
// Solidity: event Participated(uint256 initTimestamp, uint256 refundTime, bytes32 secretHash, address initiator, address participant, uint256 value)
func (_Contract *ContractFilterer) WatchParticipated(opts *bind.WatchOpts, sink chan<- *ContractParticipated) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "Participated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractParticipated)
				if err := _Contract.contract.UnpackLog(event, "Participated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseParticipated is a log parse operation binding the contract event 0xe5571d467a528d7481c0e3bdd55ad528d0df6b457b07bab736c3e245c3aa16f4.
//
// Solidity: event Participated(uint256 initTimestamp, uint256 refundTime, bytes32 secretHash, address initiator, address participant, uint256 value)
func (_Contract *ContractFilterer) ParseParticipated(log types.Log) (*ContractParticipated, error) {
	event := new(ContractParticipated)
	if err := _Contract.contract.UnpackLog(event, "Participated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractRedeemedIterator is returned from FilterRedeemed and is used to iterate over the raw logs and unpacked data for Redeemed events raised by the Contract contract.
type ContractRedeemedIterator struct {
	Event *ContractRedeemed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractRedeemedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractRedeemed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractRedeemed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractRedeemedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractRedeemedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractRedeemed represents a Redeemed event raised by the Contract contract.
type ContractRedeemed struct {
	RedeemTime *big.Int
	SecretHash [32]byte
	Secret     [32]byte
	Redeemer   common.Address
	Value      *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterRedeemed is a free log retrieval operation binding the contract event 0xe4da013d8c42cdfa76ab1d5c08edcdc1503d2da88d7accc854f0e57ebe45c591.
//
// Solidity: event Redeemed(uint256 redeemTime, bytes32 secretHash, bytes32 secret, address redeemer, uint256 value)
func (_Contract *ContractFilterer) FilterRedeemed(opts *bind.FilterOpts) (*ContractRedeemedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "Redeemed")
	if err != nil {
		return nil, err
	}
	return &ContractRedeemedIterator{contract: _Contract.contract, event: "Redeemed", logs: logs, sub: sub}, nil
}

// WatchRedeemed is a free log subscription operation binding the contract event 0xe4da013d8c42cdfa76ab1d5c08edcdc1503d2da88d7accc854f0e57ebe45c591.
//
// Solidity: event Redeemed(uint256 redeemTime, bytes32 secretHash, bytes32 secret, address redeemer, uint256 value)
func (_Contract *ContractFilterer) WatchRedeemed(opts *bind.WatchOpts, sink chan<- *ContractRedeemed) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "Redeemed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractRedeemed)
				if err := _Contract.contract.UnpackLog(event, "Redeemed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRedeemed is a log parse operation binding the contract event 0xe4da013d8c42cdfa76ab1d5c08edcdc1503d2da88d7accc854f0e57ebe45c591.
//
// Solidity: event Redeemed(uint256 redeemTime, bytes32 secretHash, bytes32 secret, address redeemer, uint256 value)
func (_Contract *ContractFilterer) ParseRedeemed(log types.Log) (*ContractRedeemed, error) {
	event := new(ContractRedeemed)
	if err := _Contract.contract.UnpackLog(event, "Redeemed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractRefundedIterator is returned from FilterRefunded and is used to iterate over the raw logs and unpacked data for Refunded events raised by the Contract contract.
type ContractRefundedIterator struct {
	Event *ContractRefunded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractRefundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractRefunded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractRefunded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractRefundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractRefundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractRefunded represents a Refunded event raised by the Contract contract.
type ContractRefunded struct {
	RefundTime *big.Int
	SecretHash [32]byte
	Refunder   common.Address
	Value      *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterRefunded is a free log retrieval operation binding the contract event 0xadb1dca52dfad065e50a1e25c2ee47ae54013a1f2d6f8ea5abace52eb4b7a4c8.
//
// Solidity: event Refunded(uint256 refundTime, bytes32 secretHash, address refunder, uint256 value)
func (_Contract *ContractFilterer) FilterRefunded(opts *bind.FilterOpts) (*ContractRefundedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "Refunded")
	if err != nil {
		return nil, err
	}
	return &ContractRefundedIterator{contract: _Contract.contract, event: "Refunded", logs: logs, sub: sub}, nil
}

// WatchRefunded is a free log subscription operation binding the contract event 0xadb1dca52dfad065e50a1e25c2ee47ae54013a1f2d6f8ea5abace52eb4b7a4c8.
//
// Solidity: event Refunded(uint256 refundTime, bytes32 secretHash, address refunder, uint256 value)
func (_Contract *ContractFilterer) WatchRefunded(opts *bind.WatchOpts, sink chan<- *ContractRefunded) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "Refunded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractRefunded)
				if err := _Contract.contract.UnpackLog(event, "Refunded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRefunded is a log parse operation binding the contract event 0xadb1dca52dfad065e50a1e25c2ee47ae54013a1f2d6f8ea5abace52eb4b7a4c8.
//
// Solidity: event Refunded(uint256 refundTime, bytes32 secretHash, address refunder, uint256 value)
func (_Contract *ContractFilterer) ParseRefunded(log types.Log) (*ContractRefunded, error) {
	event := new(ContractRefunded)
	if err := _Contract.contract.UnpackLog(event, "Refunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
