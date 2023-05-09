package evm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/uint256"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	contractAddr types.Address
)

func Test_callEVM_Deploy(t *testing.T) {
	input, err := abiContract.Pack("", "TokenOnRigo", "TOR")
	require.NoError(t, err)

	// deploy code = contract byte code + input parameters
	input = append(buildInfo.Bytecode, input...)

	// make transaction
	fromAcct := acctHandler.walletsArr[0].GetAccount()
	to := types.ZeroAddress()

	ret, xerr := rigoEVM.execVM(fromAcct.Address, to, fromAcct.GetNonce(), uint256.NewInt(0), input, 1, time.Now().UnixNano())
	require.NoError(t, xerr)
	require.NoError(t, ret.Err)

	contractAddr = ret.ReturnData
	fmt.Println("TestDeploy", "contract address", contractAddr)

	retUnpack, xerr := execMethod(fromAcct.Address, contractAddr, fromAcct.GetNonce(), uint256.NewInt(0), 2, time.Now().UnixNano(), "name")
	require.NoError(t, xerr)
	require.NoError(t, ret.Err)
	fmt.Println("TestDeploy", "name", retUnpack[0])

	retUnpack, xerr = execMethod(fromAcct.Address, contractAddr, fromAcct.GetNonce(), uint256.NewInt(0), 2, time.Now().UnixNano(), "symbol")
	require.NoError(t, xerr)
	require.NoError(t, ret.Err)
	fmt.Println("TestDeploy", "symbol", retUnpack[0])

	_, height, xerr := rigoEVM.Commit()
	require.NoError(t, xerr)
	fmt.Println("TestDeploy", "Commit block", height)
}

func Test_callEVM_Transfer(t *testing.T) {
	fromAcct := acctHandler.walletsArr[0].GetAccount()
	toAcct := acctHandler.walletsArr[1].GetAccount()

	ret, xerr := queryMethod(fromAcct.Address, contractAddr, fromAcct.GetNonce(), uint256.NewInt(0), rigoEVM.lastBlockHeight, time.Now().UnixNano(),
		"balanceOf", toAddrArr(fromAcct.Address))
	require.NoError(t, xerr)
	fmt.Println("(BEFORE) balanceOf", fromAcct.Address, ret[0], "nonce", rigoEVM.stateDBWrapper.GetNonce(fromAcct.Address.Array20()))

	ret, xerr = queryMethod(fromAcct.Address, contractAddr, fromAcct.GetNonce(), uint256.NewInt(0), rigoEVM.lastBlockHeight, time.Now().UnixNano(),
		"balanceOf", toAddrArr(toAcct.Address))
	require.NoError(t, xerr)
	fmt.Println("(BEFORE) balanceOf", toAcct.Address, ret[0], "nonce", rigoEVM.stateDBWrapper.GetNonce(fromAcct.Address.Array20()))

	ret, xerr = execMethod(fromAcct.Address, contractAddr, fromAcct.GetNonce(), uint256.NewInt(0), rigoEVM.lastBlockHeight, time.Now().UnixNano(),
		"transfer", toAddrArr(toAcct.Address), toWei(100000000))
	require.NoError(t, xerr)
	fmt.Println("<transferred>")

	_, height, xerr := rigoEVM.Commit()
	require.NoError(t, xerr)
	fmt.Println("Commit block", height)

	ret, xerr = queryMethod(fromAcct.Address, contractAddr, fromAcct.GetNonce(), uint256.NewInt(0), rigoEVM.lastBlockHeight, time.Now().UnixNano(),
		"balanceOf", toAddrArr(fromAcct.Address))
	require.NoError(t, xerr)
	fmt.Println(" (AFTER) balanceOf", fromAcct.Address, ret[0], "nonce", rigoEVM.stateDBWrapper.GetNonce(fromAcct.Address.Array20()))

	ret, xerr = queryMethod(fromAcct.Address, contractAddr, fromAcct.GetNonce(), uint256.NewInt(0), rigoEVM.lastBlockHeight, time.Now().UnixNano(),
		"balanceOf", toAddrArr(toAcct.Address))
	require.NoError(t, xerr)
	fmt.Println(" (AFTER) balanceOf", toAcct.Address, ret[0], "nonce", rigoEVM.stateDBWrapper.GetNonce(fromAcct.Address.Array20()))

	_, height, xerr = rigoEVM.Commit()
	require.NoError(t, xerr)
	fmt.Println("Commit block", height)
	xerr = rigoEVM.Close()
	require.NoError(t, xerr)

	rigoEVM = NewEVMCtrler(dbPath, &acctHandler, tmlog.NewNopLogger())

	ret, xerr = queryMethod(fromAcct.Address, contractAddr, fromAcct.GetNonce(), uint256.NewInt(0), rigoEVM.lastBlockHeight, time.Now().UnixNano(),
		"balanceOf", toAddrArr(fromAcct.Address))
	require.NoError(t, xerr)
	fmt.Println("(REOPEN) balanceOf", fromAcct.Address, ret[0], "nonce", rigoEVM.stateDBWrapper.GetNonce(fromAcct.Address.Array20()))

	ret, xerr = queryMethod(fromAcct.Address, contractAddr, fromAcct.GetNonce(), uint256.NewInt(0), rigoEVM.lastBlockHeight, time.Now().UnixNano(),
		"balanceOf", toAddrArr(toAcct.Address))
	require.NoError(t, xerr)
	fmt.Println("(REOPEN) balanceOf", toAcct.Address, ret[0], "nonce", rigoEVM.stateDBWrapper.GetNonce(fromAcct.Address.Array20()))

	_, height, xerr = rigoEVM.Commit()
	require.NoError(t, xerr)
	fmt.Println("Commit block", height)
	xerr = rigoEVM.Close()
	require.NoError(t, xerr)
}

func execMethod(from, to types.Address, nonce uint64, amt *uint256.Int, bn, bt int64, methodName string, args ...interface{}) ([]interface{}, xerrors.XError) {
	input, err := abiContract.Pack(methodName, args...)
	if err != nil {
		return nil, xerrors.From(err)
	}

	ret, xerr := rigoEVM.execVM(from, to, nonce, amt, input, bn, bt)
	if xerr != nil {
		return nil, xerr
	}
	if ret.Err != nil {
		return nil, xerrors.From(ret.Err)
	}

	retUnpack, err := abiContract.Unpack(methodName, ret.ReturnData)
	if err != nil {
		return nil, xerrors.From(err)
	}
	return retUnpack, nil
}

func queryMethod(from, to types.Address, nonce uint64, amt *uint256.Int, bn, bt int64, methodName string, args ...interface{}) ([]interface{}, xerrors.XError) {
	input, err := abiContract.Pack(methodName, args...)
	if err != nil {
		return nil, xerrors.From(err)
	}

	ret, xerr := rigoEVM.queryVM(from, to, input, bn, bt)
	if xerr != nil {
		return nil, xerr
	}
	if ret.Err != nil {
		return nil, xerrors.From(ret.Err)
	}

	retUnpack, err := abiContract.Unpack(methodName, ret.ReturnData)
	if err != nil {
		return nil, xerrors.From(err)
	}
	return retUnpack, nil
}

func toWei(c int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(c), big.NewInt(1000000000000000000))
}

func toAddrArr(addr []byte) common.Address {
	var ret common.Address
	copy(ret[:], addr)
	return ret
}

var (
	rigoEVM     *EVMCtrler
	buildInfo   TruffleBuild
	abiContract abi.ABI
	abiFile     = "./erc20_test_contract.json"
	acctHandler acctHandlerMock
	dbPath      = filepath.Join(os.TempDir(), "rigo-evm-test")
)

type TruffleBuild struct {
	ABI              json.RawMessage `json:"abi"`
	Bytecode         hexutil.Bytes   `json:"bytecode"`
	DeployedBytecode hexutil.Bytes   `json:"deployedBytecode"`
}

func init() {
	// initialize acctHandler
	acctHandler.origin = true
	acctHandler.walletsMap = make(map[string]*web3.Wallet)
	for i := 0; i < 10; i++ {
		w := web3.NewWallet(nil)
		w.GetAccount().AddBalance(uint256.MustFromDecimal("1000000000000000000000000000"))
		acctHandler.walletsMap[w.Address().String()] = w
		acctHandler.walletsArr = append(acctHandler.walletsArr, w)
	}

	// load erc20 contract
	if bz, err := ioutil.ReadFile(abiFile); err != nil {
		panic(err)
	} else if err := json.Unmarshal(bz, &buildInfo); err != nil {
		panic(err)
	} else if abiContract, err = abi.JSON(bytes.NewReader(buildInfo.ABI)); err != nil {
		panic(err)
	} else {
		for _, method := range abiContract.Methods {
			fmt.Printf("%x: %s\n", method.ID, method.Sig)
		}
	}

	rigoEVM = NewEVMCtrler(dbPath, &acctHandler, tmlog.NewNopLogger())
}

type acctHandlerMock struct {
	walletsMap map[string]*web3.Wallet
	walletsArr []*web3.Wallet
	contAccts  []*ctrlertypes.Account
	origin     bool
}

func (handler *acctHandlerMock) FindOrNewAccount(addr types.Address, exec bool) *ctrlertypes.Account {
	ret := handler.FindAccount(addr, exec)
	if ret != nil {
		return ret
	}
	ret = ctrlertypes.NewAccount(addr)
	handler.contAccts = append(handler.contAccts, ret)
	return ret
}

func (handler *acctHandlerMock) FindAccount(addr types.Address, exec bool) *ctrlertypes.Account {
	if w, ok := handler.walletsMap[addr.String()]; ok {
		return w.GetAccount()
	}
	for _, a := range handler.contAccts {
		if bytes.Compare(addr, a.Address) == 0 {
			return a
		}
	}
	return nil
}
func (a *acctHandlerMock) Transfer(from, to types.Address, amt *uint256.Int, exec bool) xerrors.XError {
	if sender := a.FindAccount(from, exec); sender == nil {
		return xerrors.ErrNotFoundAccount
	} else if receiver := a.FindAccount(to, exec); receiver == nil {
		return xerrors.ErrNotFoundAccount
	} else if xerr := sender.SubBalance(amt); xerr != nil {
		return xerr
	} else if xerr := receiver.AddBalance(amt); xerr != nil {
		return xerr
	}
	return nil
}
func (a *acctHandlerMock) Reward(to types.Address, amt *uint256.Int, exec bool) xerrors.XError {
	if receiver := a.FindAccount(to, exec); receiver == nil {
		return xerrors.ErrNotFoundAccount
	} else if xerr := receiver.AddBalance(amt); xerr != nil {
		return xerr
	}
	return nil
}

func (handler *acctHandlerMock) ImmutableAcctCtrlerAt(i int64) (ctrlertypes.IAccountHandler, xerrors.XError) {
	walletsMap := make(map[string]*web3.Wallet)
	walletsArr := make([]*web3.Wallet, len(handler.walletsArr))
	for i, w := range handler.walletsArr {
		w0 := w.Clone()
		walletsMap[w.Address().String()] = w0
		walletsArr[i] = w0
	}
	contAccts := make([]*ctrlertypes.Account, len(handler.contAccts))
	for i, a := range handler.contAccts {
		contAccts[i] = a.Clone()
	}

	return &acctHandlerMock{
		walletsMap: walletsMap,
		walletsArr: walletsArr,
		contAccts:  contAccts,
		origin:     false,
	}, nil
}

var _ ctrlertypes.IAccountHandler = (*acctHandlerMock)(nil)
