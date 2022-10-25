package rpc

import (
	"encoding/hex"
	"github.com/kysee/arcanus/ctrlers/account"
	"github.com/kysee/arcanus/ctrlers/stake"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/tendermint/tendermint/libs/json"
	tmrpccore "github.com/tendermint/tendermint/rpc/core"
	tmrpccoretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmrpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	"regexp"
	"strings"
)

var hexReg = regexp.MustCompile(`(?i)[a-f0-9]{40,}`)

func queryAccount(ctx *tmrpctypes.Context, addr string) (types.IAccount, error) {
	bzAddr, err := types.AddressFromHex(addr)
	if err != nil {
		return nil, xerrors.NewFrom(err)
	}

	qd := &types.QueryData{
		Command: types.QUERY_ACCOUNT,
		Params:  bzAddr,
	}
	resp, err := tmrpccore.ABCIQuery(ctx, ctx.HTTPReq.RequestURI, qd.Encode(), 0, false)
	if err != nil {
		// not reachable
		// ABCIQuery never returns error
		return nil, err
	}

	if resp.Response.Code == xerrors.ErrCodeSuccess {
		acct, err := account.DecodeAccount(resp.Response.Value)
		if err != nil {
			return nil, err
		}
		return acct, nil
	} else {
		return nil, xerrors.NewWith(resp.Response.Code, resp.Response.Log)
	}
}

func QueryAccount(ctx *tmrpctypes.Context, addr string) (*account.Account, error) {
	acct, err := queryAccount(ctx, addr)
	if err != nil {
		if xerr, ok := err.(xerrors.XError); ok && xerr.Code() == xerrors.ErrNotFoundAccount.Code() {
			addrbz, _ := hex.DecodeString(addr)
			acct = account.NewAccountWithName(addrbz, "")
		} else {
			return nil, err
		}
	}

	switch acct.Type() {
	case types.ACCT_COMMON_TYPE:
		return acct.(*account.Account), nil
	default:
		return nil, xerrors.New("error account type: unknown account type")
	}
}

func QueryAcctNonce(ctx *tmrpctypes.Context, addr string) (*ResponseAcctNonce, error) {
	acct, err := queryAccount(ctx, addr)
	if err != nil {
		return nil, err
	}
	return &ResponseAcctNonce{
		Nonce: acct.GetNonce(),
	}, nil
}

func queryStakes(ctx *tmrpctypes.Context, addr string) (*stake.Delegatee, error) {
	bzAddr, err := types.AddressFromHex(addr)
	if err != nil {
		return nil, xerrors.NewFrom(err)
	}

	qd := &types.QueryData{
		Command: types.QUERY_STAKES,
		Params:  bzAddr,
	}

	resp, err := tmrpccore.ABCIQuery(ctx, ctx.HTTPReq.RequestURI, qd.Encode(), 0, false)
	if err != nil {
		// not reachable
		// ABCIQuery never returns error
		return nil, err
	}

	if resp.Response.Code == xerrors.ErrCodeSuccess {
		sset := &stake.Delegatee{}
		err := json.Unmarshal(resp.Response.Value, sset)
		if err != nil {
			return nil, err
		}
		return sset, nil
	} else {
		return nil, xerrors.NewWith(resp.Response.Code, resp.Response.Log)
	}
}

func QueryStakes(ctx *tmrpctypes.Context, addr string) (*stake.Delegatee, error) {
	staker, err := queryStakes(ctx, addr)
	if err != nil {
		return nil, err
	}
	return staker, nil
}

func Subscribe(ctx *tmrpctypes.Context, query string) (*tmrpccoretypes.ResultSubscribe, error) {
	// return error when the event subscription request is received over http session.
	// related to: #103
	if ctx.WSConn == nil || ctx.JSONReq == nil {
		return nil, xerrors.New("error connection type: no websocket connection")
	}
	// make hex string like address or hash be uppercase
	//  address's size is 20bytes(40characters)
	//  hash's size is 32bytes(64characters)
	return tmrpccore.Subscribe(ctx, hexReg.ReplaceAllStringFunc(query, strings.ToUpper))
}

func Unsubscribe(ctx *tmrpctypes.Context, query string) (*tmrpccoretypes.ResultUnsubscribe, error) {
	return tmrpccore.Unsubscribe(ctx, hexReg.ReplaceAllStringFunc(query, strings.ToUpper))
}

func TxSearch(
	ctx *tmrpctypes.Context,
	query string,
	prove bool,
	pagePtr, perPagePtr *int,
	orderBy string,
) (*tmrpccoretypes.ResultTxSearch, error) {
	return tmrpccore.TxSearch(ctx, hexReg.ReplaceAllStringFunc(query, strings.ToUpper), prove, pagePtr, perPagePtr, orderBy)
}
