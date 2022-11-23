package rpc

import (
	"encoding/json"
	"github.com/kysee/arcanus/ctrlers/account"
	accttypes "github.com/kysee/arcanus/types/account"
	"github.com/kysee/arcanus/types/trxs"
	"github.com/kysee/arcanus/types/xerrors"
	tmjson "github.com/tendermint/tendermint/libs/json"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"math/big"
	"sync"
)

var (
	rpcClient = []*JSONRpcClient{
		NewRpcClient("http://3.37.191.127:26657"),
		NewRpcClient("http://3.34.201.6:26657"),
		NewRpcClient("http://15.165.45.176:26657"),
		NewRpcClient("http://15.165.38.111:26657"),
	}

	idx = 0
)

// todo: add Method object

var mtx sync.RWMutex

func getRpcClient() *JSONRpcClient {
	mtx.Lock()
	defer mtx.Unlock()

	idx++
	return rpcClient[idx%4]
}

func GetBalance(addr accttypes.Address) *big.Int {
	if req, err := NewRequest("account", addr.String()); err != nil {
		panic(err)
	} else if resp, err := getRpcClient().call(req); err != nil {
		panic(err)
	} else if acct, err := account.DecodeAccount(resp.Result); err != nil {
		return big.NewInt(0)
	} else {
		return acct.GetBalance()
	}
}

func GetAccount(addr accttypes.Address) (*accttypes.Account, error) {
	acct := accttypes.NewAccount(nil)
	if req, err := NewRequest("account", addr.String()); err != nil {
		panic(err)
	} else if resp, err := getRpcClient().call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, xerrors.New("rpc error: " + string(resp.Error))
	} else if err := tmjson.Unmarshal(resp.Result, acct); err != nil {
		return nil, err
	} else {
		return acct, nil
	}
}

func SendTransactionAsync(tx *trxs.Trx) (*coretypes.ResultBroadcastTx, error) {
	return sendTransaction(tx, "broadcast_tx_async")
}
func SendTransactionSync(tx *trxs.Trx) (*coretypes.ResultBroadcastTx, error) {
	return sendTransaction(tx, "broadcast_tx_sync")
}
func SendTransactionCommit(tx *trxs.Trx) (*coretypes.ResultBroadcastTx, error) {
	return sendTransaction(tx, "broadcast_tx_commit")
}

func sendTransaction(tx *trxs.Trx, method string) (*coretypes.ResultBroadcastTx, error) {

	if txbz, err := tx.Encode(); err != nil {
		return nil, err
	} else if req, err := NewRequest(method, txbz); err != nil {
		return nil, err
	} else if resp, err := getRpcClient().call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, xerrors.New("rpc error: " + string(resp.Error))
	} else {
		switch method {
		case "broadcast_tx_async":
			return nil, xerrors.New("not supported method: " + method)
		case "broadcast_tx_sync":
			ret := &coretypes.ResultBroadcastTx{}
			if err := json.Unmarshal(resp.Result, ret); err != nil {
				return nil, err
			}
			return ret, nil
		case "broadcast_tx_commit":
			return nil, xerrors.New("not supported method: " + method)
		default:
			return nil, xerrors.New("unknown method: " + method)
		}
	}
}

type TrxResult struct {
	*coretypes.ResultTx
	TxDetail *trxs.Trx `json:"tx_detail"`
}

func GetTransaction(txhash []byte) (*TrxResult, error) {
	txRet := &TrxResult{
		ResultTx: &coretypes.ResultTx{},
		TxDetail: &trxs.Trx{},
	}

	if req, err := NewRequest("tx", txhash, false); err != nil {
		return nil, err
	} else if resp, err := getRpcClient().call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, xerrors.New("rpc error: " + string(resp.Error))
	} else if err := tmjson.Unmarshal(resp.Result, txRet.ResultTx); err != nil {
		return nil, err
	} else if err := txRet.TxDetail.Decode(txRet.ResultTx.Tx); err != nil {
		return nil, err
	} else {
		return txRet, nil
	}
}
