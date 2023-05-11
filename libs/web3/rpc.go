package web3

import (
	"encoding/json"
	"errors"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/ctrlers/stake"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	rweb3types "github.com/rigochain/rigo-go/libs/web3/types"
	"github.com/rigochain/rigo-go/rpc"
	"github.com/rigochain/rigo-go/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"strconv"
	"strings"
)

func (rweb3 *RigoWeb3) GetAccount(addr types.Address) (*ctrlertypes.Account, error) {
	queryResp := &rpc.QueryResult{}

	if req, err := rweb3.NewRequest("account", addr.String()); err != nil {
		panic(err)
	} else if resp, err := rweb3.provider.Call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, errors.New("provider error: " + string(resp.Error))
	} else if err := tmjson.Unmarshal(resp.Result, queryResp); err != nil {
		return nil, err
	}

	_acct := &struct {
		Address types.Address `json:"address"`
		Name    string        `json:"name,omitempty"`
		Nonce   uint64        `json:"nonce,string"`
		Balance string        `json:"balance"`
		Code    []byte        `json:"code,omitempty"`
	}{}

	if err := tmjson.Unmarshal(queryResp.Value, _acct); err != nil {
		return nil, err
	} else {
		var bal *uint256.Int
		if strings.HasPrefix(_acct.Balance, "0x") {
			bal = uint256.MustFromHex(_acct.Balance)
		} else {
			bal = uint256.MustFromDecimal(_acct.Balance)
		}

		return &ctrlertypes.Account{
			Address: _acct.Address,
			Name:    _acct.Name,
			Nonce:   _acct.Nonce,
			Balance: bal,
			Code:    _acct.Code,
		}, nil
	}
}

func (rweb3 *RigoWeb3) GetDelegatee(addr types.Address) (*stake.Delegatee, error) {
	queryResp := &rpc.QueryResult{}
	delegatee := stake.NewDelegatee(nil, nil)

	if req, err := rweb3.NewRequest("delegatee", addr.String()); err != nil {
		panic(err)
	} else if resp, err := rweb3.provider.Call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, errors.New("provider error: " + string(resp.Error))
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
	if req, err := rweb3.NewRequest("stakes", addr.String()); err != nil {
		panic(err)
	} else if resp, err := rweb3.provider.Call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, errors.New("provider error: " + string(resp.Error))
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
	} else if req, err := rweb3.NewRequest(method, txbz); err != nil {
		return nil, err
	} else if resp, err := rweb3.provider.Call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, errors.New("provider error: " + string(resp.Error))
	} else {
		switch method {
		case "broadcast_tx_async":
			return nil, errors.New("not supported method: " + method)
		case "broadcast_tx_sync":
			ret := &coretypes.ResultBroadcastTx{}
			if err := json.Unmarshal(resp.Result, ret); err != nil {
				return nil, err
			}
			return ret, nil
		case "broadcast_tx_commit":
			return nil, errors.New("not supported method: " + method)
		default:
			return nil, errors.New("unknown method: " + method)
		}
	}
}

func (rweb3 *RigoWeb3) GetTransaction(txhash []byte) (*rweb3types.TrxResult, error) {
	txRet := &rweb3types.TrxResult{
		ResultTx: &coretypes.ResultTx{},
		TxDetail: &ctrlertypes.Trx{},
	}

	if req, err := rweb3.NewRequest("tx", txhash, false); err != nil {
		return nil, err
	} else if resp, err := rweb3.provider.Call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, errors.New("provider error: " + string(resp.Error))
	} else if err := tmjson.Unmarshal(resp.Result, txRet.ResultTx); err != nil {
		return nil, err
	} else if err := txRet.TxDetail.Decode(txRet.ResultTx.Tx); err != nil {
		return nil, err
	} else {
		return txRet, nil
	}
}

func (rweb3 *RigoWeb3) GetValidators(height int64, page, perPage int) (*coretypes.ResultValidators, error) {
	retVals := &coretypes.ResultValidators{}

	_height := strconv.FormatInt(height, 10)
	if page == 0 {
		page = 1
	}
	_page := strconv.Itoa(page)
	_perPage := strconv.Itoa(perPage)

	if req, err := rweb3.NewRequest("validators", _height, _page, _perPage); err != nil {
		return nil, err
	} else if resp, err := rweb3.provider.Call(req); err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, errors.New("provider error: " + string(resp.Error))
	} else if err := tmjson.Unmarshal(resp.Result, retVals); err != nil {
		return nil, err
	}
	return retVals, nil
}

func (rweb3 *RigoWeb3) VmCall(from, to types.Address, height int64, data []byte) (*ctrlertypes.VMCallResult, error) {
	req, err := rweb3.NewRequest("vm_call", from, to, strconv.FormatInt(height, 10), data)
	if err != nil {
		return nil, err
	}
	resp, err := rweb3.provider.Call(req)
	if err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, errors.New("provider error: " + string(resp.Error))
	}

	qryResp := &rpc.QueryResult{}
	if err := tmjson.Unmarshal(resp.Result, qryResp); err != nil {
		return nil, err
	}

	vmRet := &ctrlertypes.VMCallResult{}
	if err := tmjson.Unmarshal(qryResp.Value, vmRet); err != nil {
		return nil, err
	}
	return vmRet, nil
}
