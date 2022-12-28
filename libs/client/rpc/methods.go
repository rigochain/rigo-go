package rpc

import (
	"encoding/json"
	"github.com/kysee/arcanus/ctrlers/stake"
	types2 "github.com/kysee/arcanus/ctrlers/types"
	"github.com/kysee/arcanus/rpc"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
	tmjson "github.com/tendermint/tendermint/libs/json"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"sync"
)

var (
	rpcClient = []*JSONRpcClient{
		NewRpcClient("http://localhost:26657"),
		//NewRpcClient("http://3.37.191.127:26657"),
		//NewRpcClient("http://3.34.201.6:26657"),
		//NewRpcClient("http://15.165.45.176:26657"),
		//NewRpcClient("http://15.165.38.111:26657"),
	}

	idx = 0
)

// todo: add Method object

var mtx sync.RWMutex

func getRpcClient() *JSONRpcClient {
	mtx.Lock()
	defer mtx.Unlock()

	idx++
	return rpcClient[idx%len(rpcClient)]
}

func GetAccount(addr types.Address) (*types2.Account, error) {
	queryResp := &rpc.QueryResult{}
	acct := types2.NewAccount(nil)
	if req, err := NewRequest("account", addr.String()); err != nil {
		panic(err)
	} else if resp, err := getRpcClient().call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, xerrors.New("rpc error: " + string(resp.Error))
	} else if err := tmjson.Unmarshal(resp.Result, queryResp); err != nil {
		return nil, err
	} else if err := tmjson.Unmarshal(queryResp.Value, acct); err != nil {
		return nil, err
	} else {
		return acct, nil
	}
}

func GetDelegatee(addr types.Address) (*stake.Delegatee, error) {
	queryResp := &rpc.QueryResult{}
	delegatee := stake.NewDelegatee(nil, nil)
	if req, err := NewRequest("delegatee", addr.String()); err != nil {
		panic(err)
	} else if resp, err := getRpcClient().call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, xerrors.New("rpc error: " + string(resp.Error))
	} else if err := tmjson.Unmarshal(resp.Result, queryResp); err != nil {
		return nil, err
	} else if err := tmjson.Unmarshal(queryResp.Value, delegatee); err != nil {
		return nil, err
	} else {
		return delegatee, nil
	}
}

func GetStakes(addr types.Address) ([]*stake.Stake, error) {
	queryResp := &rpc.QueryResult{}
	var stakes []*stake.Stake
	if req, err := NewRequest("stakes", addr.String()); err != nil {
		panic(err)
	} else if resp, err := getRpcClient().call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, xerrors.New("rpc error: " + string(resp.Error))
	} else if err := tmjson.Unmarshal(resp.Result, queryResp); err != nil {
		return nil, err
	} else if err := tmjson.Unmarshal(queryResp.Value, &stakes); err != nil {
		return nil, err
	} else {
		return stakes, nil
	}
}

func SendTransactionAsync(tx *types2.Trx) (*coretypes.ResultBroadcastTx, error) {
	return sendTransaction(tx, "broadcast_tx_async")
}
func SendTransactionSync(tx *types2.Trx) (*coretypes.ResultBroadcastTx, error) {
	return sendTransaction(tx, "broadcast_tx_sync")
}
func SendTransactionCommit(tx *types2.Trx) (*coretypes.ResultBroadcastTx, error) {
	return sendTransaction(tx, "broadcast_tx_commit")
}

func sendTransaction(tx *types2.Trx, method string) (*coretypes.ResultBroadcastTx, error) {

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
	TxDetail *types2.Trx `json:"tx_detail"`
}

func GetTransaction(txhash []byte) (*TrxResult, error) {
	txRet := &TrxResult{
		ResultTx: &coretypes.ResultTx{},
		TxDetail: &types2.Trx{},
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
