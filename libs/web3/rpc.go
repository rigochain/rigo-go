package web3

import (
	"encoding/json"
	"github.com/rigochain/rigo-go/ctrlers/stake"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	rweb3types "github.com/rigochain/rigo-go/libs/web3/types"
	"github.com/rigochain/rigo-go/rpc"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	tmjson "github.com/tendermint/tendermint/libs/json"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

func (rweb3 *RigoWeb3) GetAccount(addr types.Address) (*ctrlertypes.Account, error) {
	queryResp := &rpc.QueryResult{}
	acct := ctrlertypes.NewAccount(nil)

	rweb3.callId++

	if req, err := rweb3types.NewRequest(rweb3.callId, "account", addr.String()); err != nil {
		panic(err)
	} else if resp, err := rweb3.provider.Call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, xerrors.NewOrdinary("provider error: " + string(resp.Error))
	} else if err := tmjson.Unmarshal(resp.Result, queryResp); err != nil {
		return nil, err
	} else if err := tmjson.Unmarshal(queryResp.Value, acct); err != nil {
		return nil, err
	} else {
		return acct, nil
	}
}

func (rweb3 *RigoWeb3) GetDelegatee(addr types.Address) (*stake.Delegatee, error) {
	queryResp := &rpc.QueryResult{}
	delegatee := stake.NewDelegatee(nil, nil)

	rweb3.callId++

	if req, err := rweb3types.NewRequest(rweb3.callId, "delegatee", addr.String()); err != nil {
		panic(err)
	} else if resp, err := rweb3.provider.Call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, xerrors.NewOrdinary("provider error: " + string(resp.Error))
	} else if err := tmjson.Unmarshal(resp.Result, queryResp); err != nil {
		return nil, err
	} else if err := tmjson.Unmarshal(queryResp.Value, delegatee); err != nil {
		return nil, err
	} else {
		return delegatee, nil
	}
}

func (rweb3 *RigoWeb3) GetStakes(addr types.Address) ([]*stake.Stake, error) {
	queryResp := &rpc.QueryResult{}
	var stakes []*stake.Stake
	if req, err := rweb3types.NewRequest(rweb3.callId, "stakes", addr.String()); err != nil {
		panic(err)
	} else if resp, err := rweb3.provider.Call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, xerrors.NewOrdinary("provider error: " + string(resp.Error))
	} else if err := tmjson.Unmarshal(resp.Result, queryResp); err != nil {
		return nil, err
	} else if err := tmjson.Unmarshal(queryResp.Value, &stakes); err != nil {
		return nil, err
	} else {
		return stakes, nil
	}
}

func (rweb3 *RigoWeb3) SendTransactionAsync(tx *ctrlertypes.Trx) (*coretypes.ResultBroadcastTx, error) {
	return rweb3.sendTransaction(tx, "broadcast_tx_async")
}
func (rweb3 *RigoWeb3) SendTransactionSync(tx *ctrlertypes.Trx) (*coretypes.ResultBroadcastTx, error) {
	return rweb3.sendTransaction(tx, "broadcast_tx_sync")
}
func (rweb3 *RigoWeb3) SendTransactionCommit(tx *ctrlertypes.Trx) (*coretypes.ResultBroadcastTx, error) {
	return rweb3.sendTransaction(tx, "broadcast_tx_commit")
}

func (rweb3 *RigoWeb3) sendTransaction(tx *ctrlertypes.Trx, method string) (*coretypes.ResultBroadcastTx, error) {

	if txbz, err := tx.Encode(); err != nil {
		return nil, err
	} else if req, err := rweb3types.NewRequest(rweb3.callId, method, txbz); err != nil {
		return nil, err
	} else if resp, err := rweb3.provider.Call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, xerrors.NewOrdinary("provider error: " + string(resp.Error))
	} else {
		switch method {
		case "broadcast_tx_async":
			return nil, xerrors.NewOrdinary("not supported method: " + method)
		case "broadcast_tx_sync":
			ret := &coretypes.ResultBroadcastTx{}
			if err := json.Unmarshal(resp.Result, ret); err != nil {
				return nil, err
			}
			return ret, nil
		case "broadcast_tx_commit":
			return nil, xerrors.NewOrdinary("not supported method: " + method)
		default:
			return nil, xerrors.NewOrdinary("unknown method: " + method)
		}
	}
}

func (rweb3 *RigoWeb3) GetTransaction(txhash []byte) (*rweb3types.TrxResult, error) {
	txRet := &rweb3types.TrxResult{
		ResultTx: &coretypes.ResultTx{},
		TxDetail: &ctrlertypes.Trx{},
	}

	if req, err := rweb3types.NewRequest(rweb3.callId, "tx", txhash, false); err != nil {
		return nil, err
	} else if resp, err := rweb3.provider.Call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, xerrors.NewOrdinary("provider error: " + string(resp.Error))
	} else if err := tmjson.Unmarshal(resp.Result, txRet.ResultTx); err != nil {
		return nil, err
	} else if err := txRet.TxDetail.Decode(txRet.ResultTx.Tx); err != nil {
		return nil, err
	} else {
		return txRet, nil
	}
}
