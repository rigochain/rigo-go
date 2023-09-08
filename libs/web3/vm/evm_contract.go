package vm

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"os"
)

type truffleBuildInfo struct {
	ABI              json.RawMessage `json:"abi"`
	Bytecode         hexutil.Bytes   `json:"bytecode"`
	DeployedBytecode hexutil.Bytes   `json:"deployedBytecode"`
}

type EVMContract struct {
	buildInfo truffleBuildInfo
	abi       abi.ABI
	addr      types.Address
}

func NewEVMContract(path string) (*EVMContract, error) {
	bz, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	bi := truffleBuildInfo{}
	if err := json.Unmarshal(bz, &bi); err != nil {
		return nil, err
	}

	_abi, err := abi.JSON(bytes.NewReader(bi.ABI))
	if err != nil {
		return nil, err
	}
	return &EVMContract{
		buildInfo: bi,
		abi:       _abi,
	}, nil
}

func (ec *EVMContract) SetAddress(addr types.Address) {
	ec.addr = addr
}

func (ec *EVMContract) GetAddress() types.Address {
	return ec.addr
}

func (ec *EVMContract) Call(name string, args []interface{}, from types.Address, height int64, rweb3 *web3.RigoWeb3) ([]interface{}, error) {
	if ec.addr == nil {
		return nil, errors.New("no contract address")
	}
	data, err := ec.pack(name, args...)
	if err != nil {
		return nil, err
	}
	ret0, xerr := rweb3.VmCall(from, ec.addr, height, data)
	if xerr != nil {
		return nil, xerr
	}
	return ec.unpack(name, ret0.ReturnData)
}

func (ec *EVMContract) ExecAsync(name string, args []interface{}, from *web3.Wallet, nonce, gas uint64, gasPrice, amt *uint256.Int, rweb3 *web3.RigoWeb3) (*coretypes.ResultBroadcastTx, error) {
	to := ec.addr

	data, err := ec.pack(name, args...)
	if err != nil {
		return nil, err
	}

	if name == "" {
		// constructor
		to = types.ZeroAddress()
		data = append(ec.buildInfo.Bytecode, data...)
	}
	tx := web3.NewTrxContract(from.Address(), to, nonce, gas, gasPrice, amt, data)
	_, _, err = from.SignTrxRLP(tx)
	if err != nil {
		return nil, err
	}

	ret, err := rweb3.SendTransactionAsync(tx)
	if err != nil {
		return nil, err
	}
	if ret.Code == xerrors.ErrCodeSuccess && len(ret.Data) == types.AddrSize {
		ec.addr = types.Address(ret.Data)
	}

	return ret, nil
}

func (ec *EVMContract) ExecSync(name string, args []interface{}, from *web3.Wallet, nonce, gas uint64, gasPrice, amt *uint256.Int, rweb3 *web3.RigoWeb3) (*coretypes.ResultBroadcastTx, error) {
	to := ec.addr

	data, err := ec.pack(name, args...)
	if err != nil {
		return nil, err
	}

	if name == "" {
		// constructor
		to = types.ZeroAddress()
		data = append(ec.buildInfo.Bytecode, data...)
	}
	tx := web3.NewTrxContract(from.Address(), to, nonce, gas, gasPrice, amt, data)
	_, _, err = from.SignTrxRLP(tx)
	if err != nil {
		return nil, err
	}

	ret, err := rweb3.SendTransactionSync(tx)
	if err != nil {
		return nil, err
	}
	if ret.Code == xerrors.ErrCodeSuccess && len(ret.Data) == types.AddrSize {
		ec.addr = types.Address(ret.Data)
	}

	return ret, nil
}

func (ec *EVMContract) ExecCommit(name string, args []interface{}, from *web3.Wallet, nonce, gas uint64, gasPrice, amt *uint256.Int, rweb3 *web3.RigoWeb3) (*coretypes.ResultBroadcastTxCommit, error) {
	to := ec.addr

	data, err := ec.pack(name, args...)
	if err != nil {
		return nil, err
	}

	if name == "" {
		// constructor
		to = types.ZeroAddress()
		data = append(ec.buildInfo.Bytecode, data...)
	}
	tx := web3.NewTrxContract(from.Address(), to, nonce, gas, gasPrice, amt, data)
	_, _, err = from.SignTrxRLP(tx)
	if err != nil {
		return nil, err
	}

	ret, err := rweb3.SendTransactionCommit(tx)
	if err != nil {
		return nil, err
	}
	if ret.DeliverTx.Code == xerrors.ErrCodeSuccess && len(ret.DeliverTx.Data) == types.AddrSize {
		ec.addr = types.Address(ret.DeliverTx.Data)
	}

	return ret, nil
}

func (ec *EVMContract) pack(name string, args ...interface{}) ([]byte, error) {
	data, err := ec.abi.Pack(name, args...)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (ec *EVMContract) unpack(name string, bz []byte) ([]interface{}, error) {
	r, err := ec.abi.Unpack(name, bz)
	if err != nil {
		return nil, err
	}
	return r, nil
}
