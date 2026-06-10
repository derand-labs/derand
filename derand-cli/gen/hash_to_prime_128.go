// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package gen

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

// HashToPrime128MetaData contains all meta data concerning the HashToPrime128 contract.
var HashToPrime128MetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"hashToPrime\",\"inputs\":[{\"name\":\"seed\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"transcript\",\"type\":\"uint128[]\",\"internalType\":\"uint128[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint128\",\"internalType\":\"uint128\"}],\"stateMutability\":\"view\"},{\"type\":\"error\",\"name\":\"InvalidTranscript\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTranscriptAttemptCompositeNotAComposite\",\"inputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"type\":\"error\",\"name\":\"InvalidTranscriptAttemptCompositeNotDivisible\",\"inputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"type\":\"error\",\"name\":\"InvalidTranscriptAttemptElement\",\"inputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"type\":\"error\",\"name\":\"NotAPrime\",\"inputs\":[]}]",
	Bin: "0x60808060405234601557610556908161001a8239f35b5f80fdfe6080806040526004361015610012575f80fd5b5f3560e01c63d2b8580b14610025575f80fd5b346101f95760403660031901126101f95760043567ffffffffffffffff81116101f957366023820112156101f95780600401359067ffffffffffffffff82116101f95736602483830101116101f9576024359167ffffffffffffffff83116101f957366023840112156101f95782600401359367ffffffffffffffff85116101f9573660248660051b860101116101f95780825f9360246020960183378101838152039060025afa15610266575f51905f5b63ffffffff8116848110156101fd576100f08185610271565b6001810180911161017e5761011c906040519060208201526020815261011760408261027e565b6102b4565b6024641fffffffe0600585901b16850101356001600160801b038116908190036101f9576001811480156101e6575b6101d357806101a4575061015e90610300565b610192575063ffffffff905b1663ffffffff811461017e576001016100d7565b634e487b7160e01b5f52601160045260245ffd5b635d9fc11560e11b5f5260045260245ffd5b6001600160801b0391821606166101c1575063ffffffff9061016a565b633f0c5e5160e11b5f5260045260245ffd5b82630513e2bb60e31b5f5260045260245ffd5b506001600160801b03821681101561014b565b5f80fd5b61020d63ffffffff831685610271565b6001810180911161017e57610234906040519060208201526020815261011760408261027e565b61023d81610300565b15610257576040516001600160801b039091168152602090f35b63df6bafaf60e01b5f5260045ffd5b6040513d5f823e3d90fd5b9190820180921161017e57565b90601f8019910116810190811067ffffffffffffffff8211176102a057604052565b634e487b7160e01b5f52604160045260245ffd5b5f60208092604051918183925191829101835e8101838152039060025afa15610266575f5160801c60016001607f1b011790565b6001600160801b039081165f19019190821161017e57565b906001600160801b0382166002811061040c57600281148015610419575b61041257600183161561040c57610334836102e8565b5f915b60018216156103e0575f5b600c60ff821610610357575060019450505050565b604051608087901b6001600160801b0319166020820190815260f883901b6001600160f81b0319166030830152601182528591859161039760318361027e565b905190206001600160801b0316848110156103d357906103b8929189610423565b156103ca57600160ff915b0116610342565b505f9450505050565b505050600160ff916103c3565b9160019190911c60016001607f1b0316906001600160801b0390811690811461017e5760010191610337565b505f9150565b5060019150565b506003811461031e565b9160906010925f929695966040519185835285602084015285604084015260801b606083015260801b60708201528460801b608082015260055afa156101f9575f5160801c600181148015610506575b6104fe576001905b6001600160801b038581169083161061049657505f93505050565b6001600160801b0390811690831680156104ea576001600160801b0391800916906001600160801b036104c8846102e8565b1682146104e1576001016001600160801b03169061047b565b50600193505050565b634e487b7160e01b5f52601260045260245ffd5b506001925050565b506001600160801b03610518836102e8565b16811461047356fea26469706673582212207188b79c744e92d05cb322d357eefe9676f0bf0bb64d7d036df87136bc15256864736f6c63430008220033",
}

// HashToPrime128ABI is the input ABI used to generate the binding from.
// Deprecated: Use HashToPrime128MetaData.ABI instead.
var HashToPrime128ABI = HashToPrime128MetaData.ABI

// HashToPrime128Bin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use HashToPrime128MetaData.Bin instead.
var HashToPrime128Bin = HashToPrime128MetaData.Bin

// DeployHashToPrime128 deploys a new Ethereum contract, binding an instance of HashToPrime128 to it.
func DeployHashToPrime128(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *HashToPrime128, error) {
	parsed, err := HashToPrime128MetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(HashToPrime128Bin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &HashToPrime128{HashToPrime128Caller: HashToPrime128Caller{contract: contract}, HashToPrime128Transactor: HashToPrime128Transactor{contract: contract}, HashToPrime128Filterer: HashToPrime128Filterer{contract: contract}}, nil
}

// HashToPrime128 is an auto generated Go binding around an Ethereum contract.
type HashToPrime128 struct {
	HashToPrime128Caller     // Read-only binding to the contract
	HashToPrime128Transactor // Write-only binding to the contract
	HashToPrime128Filterer   // Log filterer for contract events
}

// HashToPrime128Caller is an auto generated read-only Go binding around an Ethereum contract.
type HashToPrime128Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HashToPrime128Transactor is an auto generated write-only Go binding around an Ethereum contract.
type HashToPrime128Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HashToPrime128Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type HashToPrime128Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HashToPrime128Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type HashToPrime128Session struct {
	Contract     *HashToPrime128   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HashToPrime128CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type HashToPrime128CallerSession struct {
	Contract *HashToPrime128Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// HashToPrime128TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type HashToPrime128TransactorSession struct {
	Contract     *HashToPrime128Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// HashToPrime128Raw is an auto generated low-level Go binding around an Ethereum contract.
type HashToPrime128Raw struct {
	Contract *HashToPrime128 // Generic contract binding to access the raw methods on
}

// HashToPrime128CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type HashToPrime128CallerRaw struct {
	Contract *HashToPrime128Caller // Generic read-only contract binding to access the raw methods on
}

// HashToPrime128TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type HashToPrime128TransactorRaw struct {
	Contract *HashToPrime128Transactor // Generic write-only contract binding to access the raw methods on
}

// NewHashToPrime128 creates a new instance of HashToPrime128, bound to a specific deployed contract.
func NewHashToPrime128(address common.Address, backend bind.ContractBackend) (*HashToPrime128, error) {
	contract, err := bindHashToPrime128(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &HashToPrime128{HashToPrime128Caller: HashToPrime128Caller{contract: contract}, HashToPrime128Transactor: HashToPrime128Transactor{contract: contract}, HashToPrime128Filterer: HashToPrime128Filterer{contract: contract}}, nil
}

// NewHashToPrime128Caller creates a new read-only instance of HashToPrime128, bound to a specific deployed contract.
func NewHashToPrime128Caller(address common.Address, caller bind.ContractCaller) (*HashToPrime128Caller, error) {
	contract, err := bindHashToPrime128(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &HashToPrime128Caller{contract: contract}, nil
}

// NewHashToPrime128Transactor creates a new write-only instance of HashToPrime128, bound to a specific deployed contract.
func NewHashToPrime128Transactor(address common.Address, transactor bind.ContractTransactor) (*HashToPrime128Transactor, error) {
	contract, err := bindHashToPrime128(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &HashToPrime128Transactor{contract: contract}, nil
}

// NewHashToPrime128Filterer creates a new log filterer instance of HashToPrime128, bound to a specific deployed contract.
func NewHashToPrime128Filterer(address common.Address, filterer bind.ContractFilterer) (*HashToPrime128Filterer, error) {
	contract, err := bindHashToPrime128(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &HashToPrime128Filterer{contract: contract}, nil
}

// bindHashToPrime128 binds a generic wrapper to an already deployed contract.
func bindHashToPrime128(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := HashToPrime128MetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HashToPrime128 *HashToPrime128Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _HashToPrime128.Contract.HashToPrime128Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HashToPrime128 *HashToPrime128Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HashToPrime128.Contract.HashToPrime128Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HashToPrime128 *HashToPrime128Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HashToPrime128.Contract.HashToPrime128Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HashToPrime128 *HashToPrime128CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _HashToPrime128.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HashToPrime128 *HashToPrime128TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HashToPrime128.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HashToPrime128 *HashToPrime128TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HashToPrime128.Contract.contract.Transact(opts, method, params...)
}

// HashToPrime is a free data retrieval call binding the contract method 0xd2b8580b.
//
// Solidity: function hashToPrime(bytes seed, uint128[] transcript) view returns(uint128)
func (_HashToPrime128 *HashToPrime128Caller) HashToPrime(opts *bind.CallOpts, seed []byte, transcript []*big.Int) (*big.Int, error) {
	var out []interface{}
	err := _HashToPrime128.contract.Call(opts, &out, "hashToPrime", seed, transcript)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// HashToPrime is a free data retrieval call binding the contract method 0xd2b8580b.
//
// Solidity: function hashToPrime(bytes seed, uint128[] transcript) view returns(uint128)
func (_HashToPrime128 *HashToPrime128Session) HashToPrime(seed []byte, transcript []*big.Int) (*big.Int, error) {
	return _HashToPrime128.Contract.HashToPrime(&_HashToPrime128.CallOpts, seed, transcript)
}

// HashToPrime is a free data retrieval call binding the contract method 0xd2b8580b.
//
// Solidity: function hashToPrime(bytes seed, uint128[] transcript) view returns(uint128)
func (_HashToPrime128 *HashToPrime128CallerSession) HashToPrime(seed []byte, transcript []*big.Int) (*big.Int, error) {
	return _HashToPrime128.Contract.HashToPrime(&_HashToPrime128.CallOpts, seed, transcript)
}
